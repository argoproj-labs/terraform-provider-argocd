package provider

import (
	"fmt"

	"github.com/argoproj-labs/terraform-provider-argocd/internal/utils"
	"github.com/argoproj-labs/terraform-provider-argocd/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type objectMeta struct {
	Name            types.String            `tfsdk:"name"`
	Namespace       types.String            `tfsdk:"namespace"`
	Annotations     map[string]types.String `tfsdk:"annotations"`
	Labels          map[string]types.String `tfsdk:"labels"`
	Finalizers      []types.String          `tfsdk:"finalizers"`
	Generation      types.Int64             `tfsdk:"generation"`
	ResourceVersion types.String            `tfsdk:"resource_version"`
	UID             types.String            `tfsdk:"uid"`
}

func objectMetaSchemaAttribute(objectName string, computed bool) schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Standard Kubernetes object metadata. For more info see the [Kubernetes reference](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata).",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Name of the %s, must be unique. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names", objectName),
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validators.IsDNSSubdomain(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("Namespace of the %s, must be unique. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/", objectName),
				Optional:            true,
				Computed:            computed,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validators.IsDNSSubdomain(),
				},
			},
			"annotations": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf("An unstructured key value map stored with the %s that may be used to store arbitrary metadata. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/", objectName),
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.MetadataAnnotations(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf("Map of string keys and values that can be used to organize and categorize (scope and select) the %s. May match selectors of replication controllers and services. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels", objectName),
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.MetadataLabels(),
				},
			},
			"finalizers": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf("List of finalizers attached to the %s. Finalizers are used to ensure proper resource cleanup. Must be empty before the object is deleted from the registry.", objectName),
				Computed:            computed,
				Optional:            true, // Always optional - users don't need to provide for data sources
				ElementType:         types.StringType,
			},
			"generation": schema.Int64Attribute{
				MarkdownDescription: "A sequence number representing a specific generation of the desired state.",
				Computed:            true,
			},
			"resource_version": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("An opaque value that represents the internal version of this %s that can be used by clients to determine when the %s has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency", objectName, objectName),
				Computed:            true,
			},
			"uid": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The unique in time and space value for this %s. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids", objectName),
				Computed:            true,
			},
		},
	}
}

func objectMetaSchemaListBlock(objectName string) schema.Block {
	return schema.ListNestedBlock{
		MarkdownDescription: "Standard Kubernetes object metadata. For more info see the [Kubernetes reference](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata).",
		Validators: []validator.List{
			listvalidator.IsRequired(),
			listvalidator.SizeAtLeast(1),
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: fmt.Sprintf("Name of the %s, must be unique. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names", objectName),
					Required:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Validators: []validator.String{
						validators.IsDNSSubdomain(),
					},
				},
				"namespace": schema.StringAttribute{
					MarkdownDescription: fmt.Sprintf("Namespace of the %s, must be unique. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/", objectName),
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Validators: []validator.String{
						validators.IsDNSSubdomain(),
					},
				},
				"annotations": schema.MapAttribute{
					MarkdownDescription: fmt.Sprintf("An unstructured key value map stored with the %s that may be used to store arbitrary metadata. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/", objectName),
					Optional:            true,
					ElementType:         types.StringType,
					Validators: []validator.Map{
						validators.MetadataAnnotations(),
					},
				},
				"labels": schema.MapAttribute{
					MarkdownDescription: fmt.Sprintf("Map of string keys and values that can be used to organize and categorize (scope and select) the %s. May match selectors of replication controllers and services. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels", objectName),
					Optional:            true,
					ElementType:         types.StringType,
					Validators: []validator.Map{
						validators.MetadataLabels(),
					},
				},
				"finalizers": schema.ListAttribute{
					MarkdownDescription: fmt.Sprintf("List of finalizers attached to the %s. Finalizers are used to ensure proper resource cleanup. Must be empty before the object is deleted from the registry.", objectName),
					Optional:            true,
					ElementType:         types.StringType,
				},
				"generation": schema.Int64Attribute{
					MarkdownDescription: "A sequence number representing a specific generation of the desired state.",
					Computed:            true,
				},
				"resource_version": schema.StringAttribute{
					MarkdownDescription: fmt.Sprintf("An opaque value that represents the internal version of this %s that can be used by clients to determine when the %s has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency", objectName, objectName),
					Computed:            true,
				},
				"uid": schema.StringAttribute{
					MarkdownDescription: fmt.Sprintf("The unique in time and space value for this %s. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids", objectName),
					Computed:            true,
				},
			},
		},
	}
}

func newObjectMeta(om metav1.ObjectMeta) objectMeta {
	obj := objectMeta{
		Annotations:     utils.MapMap(om.Annotations, types.StringValue),
		Labels:          utils.MapMap(om.Labels, types.StringValue),
		Generation:      types.Int64Value(om.Generation),
		Name:            types.StringValue(om.Name),
		ResourceVersion: types.StringValue(om.ResourceVersion),
	}

	// Handle namespace
	if om.Namespace != "" {
		obj.Namespace = types.StringValue(om.Namespace)
	} else {
		obj.Namespace = types.StringNull()
	}

	// Handle UID
	if string(om.UID) != "" {
		obj.UID = types.StringValue(string(om.UID))
	} else {
		obj.UID = types.StringNull()
	}

	// Handle finalizers - convert all from API
	if len(om.Finalizers) > 0 {
		obj.Finalizers = make([]types.String, len(om.Finalizers))
		for i, f := range om.Finalizers {
			obj.Finalizers[i] = types.StringValue(f)
		}
	}

	return obj
}

// newObjectMetaWithConfiguredFinalizers creates an objectMeta from the API response,
// but only includes finalizers that were configured by the user. This prevents
// system-managed finalizers from appearing in state and causing drift.
func newObjectMetaWithConfiguredFinalizers(om metav1.ObjectMeta, configuredFinalizers []types.String) objectMeta {
	obj := newObjectMeta(om)

	// Filter finalizers to only include those that were configured
	if len(configuredFinalizers) > 0 {
		// Build a set of API finalizers for quick lookup
		apiSet := make(map[string]bool)
		for _, f := range om.Finalizers {
			apiSet[f] = true
		}

		// Preserve the user's configured order by iterating through configured finalizers
		filtered := make([]types.String, 0, len(configuredFinalizers))
		for _, f := range configuredFinalizers {
			if apiSet[f.ValueString()] {
				filtered = append(filtered, f)
			}
		}

		obj.Finalizers = filtered
	} else if configuredFinalizers != nil {
		// User explicitly configured an empty list - return empty list, not nil
		obj.Finalizers = []types.String{}
	} else {
		// No configured finalizers at all (nil) - don't show any in state
		obj.Finalizers = nil
	}

	return obj
}

// mergeFinalizersForUpdate merges user-configured finalizers with existing system finalizers.
// It uses the previous state to determine which finalizers Terraform manages, allowing users
// to remove their own finalizers while preserving system-managed ones.
func mergeFinalizersForUpdate(existingFinalizers []string, oldConfigured, newConfigured []types.String) []string {
	// Build sets for efficient lookup
	oldSet := make(map[string]bool)
	for _, f := range oldConfigured {
		oldSet[f.ValueString()] = true
	}

	newSet := make(map[string]bool)
	for _, f := range newConfigured {
		newSet[f.ValueString()] = true
	}

	merged := make([]string, 0, len(existingFinalizers)+len(newConfigured))

	// Process existing finalizers from the API
	for _, existing := range existingFinalizers {
		if oldSet[existing] && !newSet[existing] {
			// This finalizer was previously managed by Terraform but user removed it
			// from config - don't preserve it (allow removal)
			continue
		}

		if !newSet[existing] {
			// This finalizer exists on API but was never in Terraform state
			// It's a system-managed finalizer - preserve it
			merged = append(merged, existing)
		} else {
			// User wants this finalizer and it exists - keep it
			merged = append(merged, existing)
			delete(newSet, existing)
		}
	}

	// Add any new finalizers the user configured that don't exist yet
	for f := range newSet {
		merged = append(merged, f)
	}

	return merged
}
