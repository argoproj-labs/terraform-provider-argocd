package utils

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
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
	setupChan := make(chan any, 1)      // a channel to write to whenever a new setup needs to be performed TODO: buffered or unbuffered?
	readyChan := make(chan struct{}, 1) // a channel to write to whenver a portforward was setup successfully TODO: buffered or unbuffered?
	failedChan := make(chan error, 1)   // errors returned when setting up port-forwards are written to this channel (buffered for asyncness with the reader)
	stopChan := make(chan struct{}, 1)  // a channel to control the port-forward operation in the background
	out := new(bytes.Buffer)            // stdout of portforward
	errOut := new(bytes.Buffer)         // stderr of portforward

	// TODO: reusing the same forwarder over and over again could be potentially dangerous due to concurrent access?
	var forwarder *portforward.PortForwarder // holds the forwarder we are going to use

	// a background job that watches for errors in the port-forward that would trigger a new setup
	checkError := func(ctx context.Context, errOut *bytes.Buffer) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if errOut.String() != "" {
					failedChan <- fmt.Errorf(errOut.String())
				}
			}
		}
	}

	// forward is a function that runs the port-forward in the background and dies when the port-forward dies
	// in case the port-forward died without an error a write to the setupChan, if there was an error, writes to the failedChan
	forward := func() {
		err = forwarder.ForwardPorts()
		log.Printf("port-forward returned with: %q", err)
		// TODO: XOR or write to setupChan anyway?
		if err == nil {
			log.Printf("port-forward returned without error")
			setupChan <- 1
		} else {
			log.Printf("port-forward returned with error: %q", err)
			failedChan <- err
		}
	}

	// setup recreates the forwarder object and thus also reselects a suitable pod as remote target
	// writes to the failedChan in case there was an error and dies by calling forward()
	setup := func() error {
		log.Println("SEARCHME: call for new port-forward setup")
		close(stopChan) // close previous port-forward explicitly (in case it was still running, setup can also happen if the other encountered an error)
		stopChan = make(chan struct{}, 1)
		readyChan = make(chan struct{}, 1) // recreate for next forward
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
		// TODO: is this additional check really needed?
		if errOut.String() != "" {
			return fmt.Errorf("%s", errOut.String())
		}

		return nil
	}

	// TODO: the failedChan would indicate fatal errors that cannot be recovered, how does the main routine get notified about such things?
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
				err := setup()
				if err != nil {
					log.Fatalf("SEARCHME: setup of a new port-forward failed, pointless to continue with the provider: %q", err)
				}
			case <-failedChan:
				// TODO: some logic because we don't want to restart on every error
				setupChan <- 1
			}
		}
	}()

	// background error checker
	go checkError(context.Background(), errOut)

	// initial port-forward setup
	return port, setup()
}

// TODO: can we simplify / update this logic? (copied from argo-cd's portforward.go file)
func newForwarder(ctx context.Context, targetPort int, namespace string, overrides *clientcmd.ConfigOverrides, port int, forwardChan chan struct{}, readyChan chan struct{}, out, errOut *bytes.Buffer, podSelectors ...string) (*portforward.PortForwarder, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, os.Stdin)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	if namespace == "" {
		namespace, _, err = clientConfig.Namespace()
		if err != nil {
			return nil, err
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
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
