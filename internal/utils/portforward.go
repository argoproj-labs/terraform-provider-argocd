package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/util/podutils"
)

// stolen from argocd pkg kube portforwarder.go
//
//	an attempt to implement retry into port-forwarding
//
// in the future we might want to file a PR for this upstream so that the CLI and others can have this too, but it needs to work for that
// guarantees that the port returned keeps the same for the entire lifetime of the provider running, but the connection to the remote pod might be recreated a couple of times
// depending on pod restarts
func SetupPortForward(ctx context.Context, targetPort int, namespace string, overrides *clientcmd.ConfigOverrides, podSelectors ...string) (int, error) {
	// first create a listener to get a random free port (but close the listener again, not needed)
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return -1, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil { // could also ignore the error, not relevant for us
		return 0, fmt.Errorf("failed to close %v: %v", ln, err)
	}

	// setup some channels for our background listener
	readyChan := make(chan struct{}) // a channel to write to whenver a portforward was setup successfully (buffered for asyncness with the reader)
	setupChan := make(chan any)      // a channel to write to whenever a new setup needs to be performed (buffered for asyncness with the reader)
	stopChan := make(chan struct{})  // a channel to control the port-forward operation in the background
	failedChan := make(chan error)   // errors returned when setting up port-forwards are written to this channel (buffered for asyncness with the reader)
	out := new(bytes.Buffer)         // stdout of portforward
	errOut := new(bytes.Buffer)      // stderr of portforward

	// TODO: reusing the same forwarder over and over again could be potentially dangerous due to concurrent access?
	var forwarder *portforward.PortForwarder // holds the forwarder we are going to use

	// forward is a function that runs the port-forward in the background and dies when the port-forward dies
	// in case the port-forward died without an error a write to the setupChan, if there was an error, writes to the failedChan
	forward := func() {
		err = forwarder.ForwardPorts()
		log.Printf("port-forward returned with: %q", err)
		// TODO: XOR or write to setupChan anyway?
		if err == nil {
			log.Printf("port-forward returned without error")
			setupChan <- struct{}{}
		} else {
			log.Printf("port-forward returned with error: %q", err)
			failedChan <- err
		}
	}

	// setup recreates the forwarder object and thus also reselects a suitable pod as remote target
	setup := func() error {
		log.Println("SEARCHME: call for new port-forward setup")
		close(stopChan) // close previous port-forward explicitly (in case it was still running)
		stopChan = make(chan struct{})
		readyChan = make(chan struct{}) // recreate for next forward
		forwarder, err = newForwarder(ctx, targetPort, namespace, overrides, port, stopChan, readyChan, out, errOut, podSelectors...)
		if err != nil {
			log.Printf("SEARCHME: creating forwarder object failed: %q", err)
			return err // when setup of the new forwarder failed, it's pointless to continue
		}

		go forward() // forward in the background (blocking operation)

		select {
		case err = <-failedChan:
			log.Printf("write to failedChan, likely a port-forward setup failure: %q", err)
			return err
		case <-readyChan:
			log.Printf("readyChan was closed, port-forward successsfully setup")
		}

		return nil
	}

	clientSet, _, err := newClientSet(namespace, overrides)
	if err != nil {
		return -1, err
	}

	go func(namespace string, clientSet *kubernetes.Clientset, setupChan chan any) {
		log.Printf("Setting up watcher for namespace: %s\n", namespace)
		// Create a SharedInformerFactory. This is the starting point for creating informers.
		// It takes our clientset, a resync period, and optional settings.
		// SharedInformerFactory efficiently shares a single connection to the API server
		// across multiple informers if you were watching different resource types.
		factory := informers.NewSharedInformerFactoryWithOptions(
			clientSet,                          // Our connection to Kubernetes API
			10*time.Second,                     // Resync period: How often the informer re-lists resources from the API server to ensure its cache is fresh.
			informers.WithNamespace(namespace), // Configure this factory to watch only this specific namespace.
			informers.WithTweakListOptions(func(opt *metav1.ListOptions) {
				// This allows us to apply additional filters to the initial LIST call and subsequent WATCH calls.
				// fields.Everything().String() means we're not applying any field-based filters here, watching all Pods.
				opt.FieldSelector = fields.Everything().String()
				opt.LabelSelector = podSelectors[0]
			}),
		)

		// Get the Pod informer from the factory.
		// `Core().V1().Pods().Informer()` is the specific informer for Pods in the "core/v1" API group.
		informer := factory.Core().V1().Pods().Informer()

		// Add event handlers. These are functions that will be called by the informer
		// whenever a specific type of event (Add, Update, Delete) occurs for a Pod.
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*corev1.Pod)
				log.Printf("Pod was deleted: %+v", pod)
				setupChan <- pod
			},
		})

		// Create a channel to signal when to stop the informer.
		// When this channel is closed, the informer will gracefully shut down.
		// stopCh := make(chan struct{})
		stopCh := context.Background().Done()
		// defer close(stopCh)

		// Start the informer factory. This begins the process of listing and watching events.
		// This also runs in a goroutine, so it doesn't block the current goroutine.
		go factory.Start(stopCh)

		// Wait for the informer's caches to be synced. This is important!
		// It ensures the informer has retrieved the initial state of all Pods
		// before it starts processing real-time events. This prevents missing initial events.
		// It will block until caches are synced or stopCh is closed.
		factory.WaitForCacheSync(stopCh)
		log.Printf("Cache synced for namespace: %s. Ready to watch events.\n", namespace)

		// This line keeps the goroutine running indefinitely.
		// It will block until the 'stopCh' channel is closed, allowing the informer to run in the background.
		// If `stopCh` is closed, this goroutine will exit.
		<-stopCh
	}(namespace, clientSet, setupChan)

	// our background watcher for events on the setupChan, triggers a new setup on event or dies when the overall context is done
	go func() {
		for {
			select {
			// case <-ctx.Done():
			case <-context.Background().Done():
				close(stopChan) // cancel current port-forward operation
				log.Printf("SEARCHME: context was marked as done with reason: %q", ctx.Err())
				return
			case <-setupChan:
				errOut.Reset()
				out.Reset()
				err := setup()
				if err != nil {
					// TODO: the failedChan would indicate fatal errors that cannot be recovered, how does the main routine get notified about such things?
					log.Fatalf("SEARCHME: setup of a new port-forward failed, pointless to continue with the provider: %q", err)
				}
			}
		}
	}()

	// initial port-forward setup
	return port, setup()
}

// TODO: can we simplify / update this logic? (copied from argo-cd's portforward.go file)
func newForwarder(ctx context.Context, targetPort int, namespace string, overrides *clientcmd.ConfigOverrides, port int, forwardChan chan struct{}, readyChan chan struct{}, out, errOut *bytes.Buffer, podSelectors ...string) (*portforward.PortForwarder, error) {
	clientSet, config, err := newClientSet(namespace, overrides)
	if err != nil {
		return nil, err
	}

	pod, err := selectPodForPortForward(clientSet, namespace, podSelectors...)
	if err != nil {
		return nil, err
	}

	url := clientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("portforward").URL()

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, fmt.Errorf("could not create round tripper: %w", err)
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	tunnelingDialer, err := portforward.NewSPDYOverWebsocketDialer(url, config)
	if err != nil {
		return nil, fmt.Errorf("could not create tunneling dialer: %w", err)
	}
	// First attempt tunneling (websocket) dialer, then fallback to spdy dialer.
	dialer = portforward.NewFallbackDialer(tunnelingDialer, dialer, func(err error) bool {
		return httpstream.IsUpgradeFailure(err) || httpstream.IsHTTPSProxyError(err)
	})

	forwarder, err := portforward.NewOnAddresses(dialer, []string{"localhost"}, []string{fmt.Sprintf("%d:%d", port, targetPort)}, forwardChan, readyChan, out, errOut)
	if err != nil {
		return nil, err
	}

	return forwarder, nil
}

func selectPodForPortForward(clientSet kubernetes.Interface, namespace string, podSelectors ...string) (*corev1.Pod, error) {
	for _, podSelector := range podSelectors {
		pods, err := clientSet.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: podSelector,
		})
		if err != nil {
			return nil, err
		}

		for _, po := range pods.Items {
			if po.Status.Phase == corev1.PodRunning && podutils.IsPodReady(&po) {
				return &po, nil
			}
		}
	}
	return nil, fmt.Errorf("cannot find ready pod with selector: %v - use the --{component}-name flag in this command or set the environmental variable (Refer to https://argo-cd.readthedocs.io/en/stable/user-guide/environment-variables), to change the Argo CD component name in the CLI", podSelectors)
}

func newClientSet(namespace string, overrides *clientcmd.ConfigOverrides) (*kubernetes.Clientset, *rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, os.Stdin)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	if namespace == "" {
		namespace, _, err = clientConfig.Namespace()
		if err != nil {
			return nil, config, err
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, config, err
	}

	return clientSet, config, err

}
