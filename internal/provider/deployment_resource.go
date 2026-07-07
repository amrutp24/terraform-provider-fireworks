package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// createTimeout bounds how long we poll for a deployment to become READY.
const createTimeout = 30 * time.Minute

var (
	_ resource.Resource                = &DeploymentResource{}
	_ resource.ResourceWithConfigure   = &DeploymentResource{}
	_ resource.ResourceWithImportState = &DeploymentResource{}
)

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

// DeploymentResource manages a Fireworks on-demand GPU inference deployment.
type DeploymentResource struct {
	client *client.Client
}

type DeploymentResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	BaseModel        types.String `tfsdk:"base_model"`
	DisplayName      types.String `tfsdk:"display_name"`
	Description      types.String `tfsdk:"description"`
	MinReplicaCount  types.Int64  `tfsdk:"min_replica_count"`
	MaxReplicaCount  types.Int64  `tfsdk:"max_replica_count"`
	AcceleratorType  types.String `tfsdk:"accelerator_type"`
	AcceleratorCount types.Int64  `tfsdk:"accelerator_count"`
	Precision        types.String `tfsdk:"precision"`
	Region           types.String `tfsdk:"region"`
	WaitForReady     types.Bool   `tfsdk:"wait_for_ready"`
	State            types.String `tfsdk:"state"`
	CreateTime       types.String `tfsdk:"create_time"`
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An on-demand GPU inference deployment. Billed per GPU-minute while replicas are running; " +
			"set min_replica_count = 0 to allow scale-to-zero.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Deployment ID (last segment of the resource name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Full resource name, e.g. accounts/my-account/deployments/abc123.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"base_model": schema.StringAttribute{
				Description: "Model to deploy, e.g. accounts/fireworks/models/llama-v3p1-8b-instruct.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable name (max 64 characters).",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the deployment.",
				Optional:    true,
			},
			"min_replica_count": schema.Int64Attribute{
				Description: "Minimum replicas. 0 enables scale-to-zero. API default: 0.",
				Optional:    true,
				Computed:    true,
			},
			"max_replica_count": schema.Int64Attribute{
				Description: "Maximum replicas for autoscaling. API default: max(min_replica_count, 1).",
				Optional:    true,
				Computed:    true,
			},
			"accelerator_type": schema.StringAttribute{
				Description: "GPU type, e.g. NVIDIA_A100_80GB, NVIDIA_H100_80GB, NVIDIA_H200_141GB, AMD_MI300X_192GB.",
				Optional:    true,
				Computed:    true,
			},
			"accelerator_count": schema.Int64Attribute{
				Description: "GPUs per replica.",
				Optional:    true,
				Computed:    true,
			},
			"precision": schema.StringAttribute{
				Description: "Model serving precision, e.g. FP16, FP8, BF16, NF4.",
				Optional:    true,
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Placement region, e.g. US_IOWA_1, EU_FRANKFURT_1.",
				Optional:    true,
			},
			"wait_for_ready": schema.BoolAttribute{
				Description: "Whether to wait until the deployment reaches READY state. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"state": schema.StringAttribute{
				Description: "Current deployment state (CREATING, READY, UPDATING, DELETING, FAILED).",
				Computed:    true,
			},
			"create_time": schema.StringAttribute{
				Description: "Creation timestamp (RFC 3339).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = c
}

// toAPI converts the Terraform model into the API request object.
func (m *DeploymentResourceModel) toAPI() *client.Deployment {
	dep := &client.Deployment{
		BaseModel:       m.BaseModel.ValueString(),
		DisplayName:     m.DisplayName.ValueString(),
		Description:     m.Description.ValueString(),
		AcceleratorType: m.AcceleratorType.ValueString(),
		Precision:       m.Precision.ValueString(),
	}
	if !m.MinReplicaCount.IsNull() && !m.MinReplicaCount.IsUnknown() {
		dep.MinReplicaCount = m.MinReplicaCount.ValueInt64Pointer()
	}
	if !m.MaxReplicaCount.IsNull() && !m.MaxReplicaCount.IsUnknown() {
		dep.MaxReplicaCount = m.MaxReplicaCount.ValueInt64Pointer()
	}
	if !m.AcceleratorCount.IsNull() && !m.AcceleratorCount.IsUnknown() {
		dep.AcceleratorCount = m.AcceleratorCount.ValueInt64Pointer()
	}
	if !m.Region.IsNull() && m.Region.ValueString() != "" {
		dep.Placement = &client.Placement{Region: m.Region.ValueString()}
	}
	// Unknown (computed, not yet decided) optional strings must not be sent.
	if m.AcceleratorType.IsUnknown() {
		dep.AcceleratorType = ""
	}
	if m.Precision.IsUnknown() {
		dep.Precision = ""
	}
	return dep
}

// fromAPI writes API response fields back into the Terraform model.
func (m *DeploymentResourceModel) fromAPI(dep *client.Deployment) {
	m.ID = types.StringValue(client.ResourceID(dep.Name))
	m.Name = types.StringValue(dep.Name)
	m.BaseModel = types.StringValue(dep.BaseModel)
	m.State = types.StringValue(dep.State)
	m.CreateTime = types.StringValue(dep.CreateTime)

	// Optional attributes: only overwrite when the API reports a value, so a
	// config that omits them keeps null instead of flapping to "".
	if dep.DisplayName != "" || !m.DisplayName.IsNull() {
		m.DisplayName = stringOrNull(dep.DisplayName)
	}
	if dep.Description != "" || !m.Description.IsNull() {
		m.Description = stringOrNull(dep.Description)
	}
	if dep.Placement != nil && dep.Placement.Region != "" {
		m.Region = types.StringValue(dep.Placement.Region)
	}

	// Optional+computed attributes always take the server's value.
	m.MinReplicaCount = int64OrNull(dep.MinReplicaCount)
	m.MaxReplicaCount = int64OrNull(dep.MaxReplicaCount)
	m.AcceleratorCount = int64OrNull(dep.AcceleratorCount)
	m.AcceleratorType = stringOrNull(dep.AcceleratorType)
	m.Precision = stringOrNull(dep.Precision)
}

func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func int64OrNull(v *int64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(*v)
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dep, err := r.client.CreateDeployment(ctx, data.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error creating deployment", err.Error())
		return
	}

	id := client.ResourceID(dep.Name)
	if data.WaitForReady.ValueBool() {
		waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
		defer cancel()
		dep, err = r.client.WaitForDeploymentReady(waitCtx, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Deployment did not become ready",
				fmt.Sprintf("Deployment %s was created but did not reach READY: %s. "+
					"It is now tracked in state; fix the issue and re-apply, or destroy it.", id, err),
			)
			// Fall through: still record what we know so the resource is tracked.
			dep = &client.Deployment{Name: fmt.Sprintf("accounts/-/deployments/%s", id)}
		}
	}

	data.fromAPI(dep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dep, err := r.client.GetDeployment(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading deployment", err.Error())
		return
	}
	if dep.State == client.DeploymentStateDeleted {
		resp.State.RemoveResource(ctx)
		return
	}

	data.fromAPI(dep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	dep, err := r.client.UpdateDeployment(ctx, id, data.toAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating deployment", err.Error())
		return
	}

	if data.WaitForReady.ValueBool() && dep.State != client.DeploymentStateReady {
		waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
		defer cancel()
		dep, err = r.client.WaitForDeploymentReady(waitCtx, id)
		if err != nil {
			resp.Diagnostics.AddError("Deployment did not become ready after update", err.Error())
			return
		}
	}

	data.fromAPI(dep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if err := r.client.DeleteDeployment(ctx, id); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting deployment", err.Error())
		return
	}

	waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()
	if err := r.client.WaitForDeploymentGone(waitCtx, id); err != nil {
		resp.Diagnostics.AddError("Error waiting for deployment deletion", err.Error())
	}
}

func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
