package provider

import (
	"context"
	"fmt"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &SecretResource{}
	_ resource.ResourceWithConfigure   = &SecretResource{}
	_ resource.ResourceWithImportState = &SecretResource{}
)

func NewSecretResource() resource.Resource {
	return &SecretResource{}
}

// SecretResource manages an account-level secret. The API never returns the
// value, so the provider keeps the configured value in state and cannot detect
// changes made outside Terraform.
type SecretResource struct {
	client *client.Client
}

type SecretResourceModel struct {
	ID       types.String `tfsdk:"id"`
	SecretID types.String `tfsdk:"secret_id"`
	Name     types.String `tfsdk:"name"`
	KeyName  types.String `tfsdk:"key_name"`
	Value    types.String `tfsdk:"value"`
}

func (r *SecretResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An account-level secret. The API never returns the value, so it is kept in Terraform state " +
			"(protect your state) and changes made outside Terraform are not detected.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Same as secret_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_id": schema.StringAttribute{
				Description: "Secret identifier, the last segment of the resource name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Full resource name, e.g. accounts/my-account/secrets/my-secret.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_name": schema.StringAttribute{
				Description: "Name the secret is exposed under, e.g. WANDB_API_KEY.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Secret value. Stored in state; never read back from the API.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *SecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// fromAPI copies server-side fields. Value is left alone: the API never
// returns it.
func (m *SecretResourceModel) fromAPI(s *client.Secret) {
	id := client.ResourceID(s.Name)
	m.ID = types.StringValue(id)
	m.SecretID = types.StringValue(id)
	m.Name = types.StringValue(s.Name)
	m.KeyName = types.StringValue(s.KeyName)
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.CreateSecret(ctx, data.SecretID.ValueString(), data.KeyName.ValueString(), data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating secret", err.Error())
		return
	}
	if out.Name == "" {
		out.Name = r.client.SecretName(data.SecretID.ValueString())
	}
	if out.KeyName == "" {
		out.KeyName = data.KeyName.ValueString()
	}

	data.fromAPI(out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetSecret(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading secret", err.Error())
		return
	}

	data.fromAPI(s)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.client.UpdateSecret(ctx, data.ID.ValueString(), data.KeyName.ValueString(), data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating secret", err.Error())
		return
	}
	if out.Name == "" {
		out.Name = r.client.SecretName(data.ID.ValueString())
	}
	if out.KeyName == "" {
		out.KeyName = data.KeyName.ValueString()
	}

	data.fromAPI(out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSecret(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting secret", err.Error())
	}
}

// ImportState sets the ID only. The value can't be imported; the next apply
// writes the configured value to the secret.
func (r *SecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
