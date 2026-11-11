package provider

import (
	"context"
	"fmt"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                   = &DeployedModelResource{}
	_ resource.ResourceWithConfigure      = &DeployedModelResource{}
	_ resource.ResourceWithImportState    = &DeployedModelResource{}
	_ resource.ResourceWithValidateConfig = &DeployedModelResource{}
)

func NewDeployedModelResource() resource.Resource {
	return &DeployedModelResource{}
}

// DeployedModelResource manages a LoRA addon loaded onto a deployment or
// onto serverless.
type DeployedModelResource struct {
	client *client.Client
}

type DeployedModelResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Model              types.String `tfsdk:"model"`
	Deployment         types.String `tfsdk:"deployment"`
	Serverless         types.Bool   `tfsdk:"serverless"`
	DisplayName        types.String `tfsdk:"display_name"`
	Description        types.String `tfsdk:"description"`
	Default            types.Bool   `tfsdk:"default"`
	Public             types.Bool   `tfsdk:"public"`
	ReplaceMergedAddon types.Bool   `tfsdk:"replace_merged_addon"`
	WaitForReady       types.Bool   `tfsdk:"wait_for_ready"`
	State              types.String `tfsdk:"state"`
	CreateTime         types.String `tfsdk:"create_time"`
}

func (r *DeployedModelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployed_model"
}

func (r *DeployedModelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A LoRA model loaded onto a dedicated deployment or onto serverless. " +
			"Set exactly one of deployment or serverless = true.",
		MarkdownDescription: "A LoRA model loaded onto a dedicated deployment or onto serverless " +
			"(the API calls this a deployed model). Set exactly one of `deployment` or `serverless = true`.\n\n" +
			"~> A dedicated deployment must have addons enabled to accept a LoRA. " +
			"Destroying this resource unloads the LoRA; the model itself stays in the account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Deployed model ID (last segment of the resource name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Full resource name, e.g. accounts/my-account/deployedModels/abc123.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"model": schema.StringAttribute{
				Description: "LoRA model to load, e.g. accounts/my-account/models/my-lora. " +
					"Changing this forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment": schema.StringAttribute{
				Description: "Deployment to load onto, e.g. accounts/my-account/deployments/abc123. " +
					"Conflicts with serverless. Changing this forces a new resource.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"serverless": schema.BoolAttribute{
				Description: "Load onto the shared serverless pool instead of a deployment. " +
					"Conflicts with deployment. Changing this forces a new resource.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable name (max 64 characters).",
				Optional:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description (max 1000 characters).",
				Optional:    true,
			},
			"default": schema.BoolAttribute{
				Description: "Whether this copy answers requests to the bare model name, without a #<deployment> suffix. " +
					"The API sets this to true for the first copy of a model.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"public": schema.BoolAttribute{
				Description: "Whether the deployed model is publicly reachable.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"replace_merged_addon": schema.BoolAttribute{
				Description: "Merge this LoRA into the base model, unloading any addon already merged there. " +
					"Required for hot-reload deployments. Only used at create time. Defaults to false.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"wait_for_ready": schema.BoolAttribute{
				Description: "Whether to wait until the deployed model reaches DEPLOYED state. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"state": schema.StringAttribute{
				Description: "Current state (DEPLOYING, DEPLOYED, UPDATING, UNDEPLOYING).",
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

func (r *DeployedModelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig enforces the API rule that exactly one of deployment or
// serverless = true is set.
func (r *DeployedModelResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var deployment types.String
	var serverless types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("deployment"), &deployment)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("serverless"), &serverless)...)
	if resp.Diagnostics.HasError() || deployment.IsUnknown() || serverless.IsUnknown() {
		return
	}

	hasDeployment := !deployment.IsNull()
	isServerless := serverless.ValueBool()
	switch {
	case hasDeployment && isServerless:
		resp.Diagnostics.AddAttributeError(path.Root("serverless"), "Conflicting attributes",
			"Set either deployment or serverless = true, not both.")
	case !hasDeployment && !isServerless:
		resp.Diagnostics.AddAttributeError(path.Root("deployment"), "Missing attribute",
			"Set either deployment or serverless = true.")
	}
}

func (m *DeployedModelResourceModel) toAPI() *client.DeployedModel {
	dm := &client.DeployedModel{
		Model:       m.Model.ValueString(),
		Deployment:  m.Deployment.ValueString(),
		DisplayName: m.DisplayName.ValueString(),
		Description: m.Description.ValueString(),
		Default:     knownBoolPointer(m.Default),
		Public:      knownBoolPointer(m.Public),
	}
	if m.Serverless.ValueBool() {
		dm.Serverless = m.Serverless.ValueBoolPointer()
	}
	return dm
}

func (m *DeployedModelResourceModel) toUpdateAPI() *client.DeployedModelUpdate {
	return &client.DeployedModelUpdate{
		DisplayName: m.DisplayName.ValueString(),
		Description: m.Description.ValueString(),
		Default:     knownBoolPointer(m.Default),
		Public:      knownBoolPointer(m.Public),
	}
}

func (m *DeployedModelResourceModel) fromAPI(dm *client.DeployedModel) {
	m.ID = types.StringValue(client.ResourceID(dm.Name))
	m.Name = types.StringValue(dm.Name)
	m.Model = types.StringValue(dm.Model)
	m.Deployment = stringOrNull(dm.Deployment)
	m.State = types.StringValue(dm.State)
	m.CreateTime = types.StringValue(dm.CreateTime)

	if dm.DisplayName != "" || !m.DisplayName.IsNull() {
		m.DisplayName = stringOrNull(dm.DisplayName)
	}
	if dm.Description != "" || !m.Description.IsNull() {
		m.Description = stringOrNull(dm.Description)
	}

	// The API omits false booleans, so nil reads as false.
	m.Serverless = types.BoolValue(dm.Serverless != nil && *dm.Serverless)
	m.Default = types.BoolValue(dm.Default != nil && *dm.Default)
	m.Public = types.BoolValue(dm.Public != nil && *dm.Public)
}

// knownBoolPointer returns nil for null or unknown values so they are not sent.
func knownBoolPointer(b types.Bool) *bool {
	if b.IsNull() || b.IsUnknown() {
		return nil
	}
	return b.ValueBoolPointer()
}

func (r *DeployedModelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeployedModelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dm, err := r.client.CreateDeployedModel(ctx, data.toAPI(), data.ReplaceMergedAddon.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error creating deployed model", err.Error())
		return
	}

	id := client.ResourceID(dm.Name)
	if data.WaitForReady.ValueBool() && dm.State != client.DeployedModelStateDeployed {
		waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
		defer cancel()
		ready, err := r.client.WaitForDeployedModelReady(waitCtx, id)
		if err != nil {
			resp.Diagnostics.AddError(
				"Deployed model did not become ready",
				fmt.Sprintf("Deployed model %s was created but did not reach DEPLOYED: %s. "+
					"It is now tracked in state; fix the issue and re-apply, or destroy it.", id, err),
			)
		} else {
			dm = ready
		}
	}

	data.fromAPI(dm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeployedModelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeployedModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dm, err := r.client.GetDeployedModel(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading deployed model", err.Error())
		return
	}

	data.fromAPI(dm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeployedModelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeployedModelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state DeployedModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	dm, err := r.client.UpdateDeployedModel(ctx, id, data.toUpdateAPI())
	if err != nil {
		resp.Diagnostics.AddError("Error updating deployed model", err.Error())
		return
	}

	if data.WaitForReady.ValueBool() && dm.State != client.DeployedModelStateDeployed {
		waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
		defer cancel()
		dm, err = r.client.WaitForDeployedModelReady(waitCtx, id)
		if err != nil {
			resp.Diagnostics.AddError("Deployed model did not become ready after update", err.Error())
			return
		}
	}

	data.fromAPI(dm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeployedModelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeployedModelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if err := r.client.DeleteDeployedModel(ctx, id); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error unloading deployed model", err.Error())
		return
	}

	waitCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()
	if err := r.client.WaitForDeployedModelGone(waitCtx, id); err != nil {
		resp.Diagnostics.AddError("Error waiting for deployed model to unload", err.Error())
	}
}

func (r *DeployedModelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
