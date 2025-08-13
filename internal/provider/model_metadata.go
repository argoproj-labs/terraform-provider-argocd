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
				MarkdownDescription: fmt.Sprintf("Name of the %s, must be unique. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names", objectName),
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
				MarkdownDescription: "An unstructured key value map stored with the cluster secret that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.MetadataAnnotations(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Map of string keys and values that can be used to organize and categorize (scope and select) the cluster secret. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels",
				Computed:            computed,
				Optional:            !computed,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.MetadataLabels(),
				},
			},
			"generation": schema.Int64Attribute{
				MarkdownDescription: "A sequence number representing a specific generation of the desired state.",
				Computed:            true,
			},
			"resource_version": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("An opaque value that represents the internal version of this %s that can be used by clients to determine when %s has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency", objectName, objectName),
				Computed:            true,
			},
			"uid": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The unique in time and space value for this %s. More info: http://kubernetes.io/docs/user-guide/identifiers#uids", objectName),
				Computed:            true,
			},
		},
	}
}

func objectMetaSchemaListBlock(objectName string, computed bool) schema.Block {
	return schema.ListNestedBlock{
		MarkdownDescription: "Standard Kubernetes object metadata. For more info see the [Kubernetes reference](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata).",
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					MarkdownDescription: fmt.Sprintf("Name of the %s, must be unique. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names", objectName),
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
					MarkdownDescription: "An unstructured key value map stored with the cluster secret that may be used to store arbitrary metadata. More info: http://kubernetes.io/docs/user-guide/annotations",
					Computed:            computed,
					Optional:            !computed,
					ElementType:         types.StringType,
					Validators: []validator.Map{
						validators.MetadataAnnotations(),
					},
				},
				"labels": schema.MapAttribute{
					MarkdownDescription: "Map of string keys and values that can be used to organize and categorize (scope and select) the cluster secret. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels",
					Computed:            computed,
					Optional:            !computed,
					ElementType:         types.StringType,
					Validators: []validator.Map{
						validators.MetadataLabels(),
					},
				},
				"generation": schema.Int64Attribute{
					MarkdownDescription: "A sequence number representing a specific generation of the desired state.",
					Computed:            true,
				},
				"resource_version": schema.StringAttribute{
					MarkdownDescription: fmt.Sprintf("An opaque value that represents the internal version of this %s that can be used by clients to determine when %s has changed. Read more: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency", objectName, objectName),
					Computed:            true,
				},
				"uid": schema.StringAttribute{
					MarkdownDescription: fmt.Sprintf("The unique in time and space value for this %s. More info: http://kubernetes.io/docs/user-guide/identifiers#uids", objectName),
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

	return obj
}
