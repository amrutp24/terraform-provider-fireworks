package provider

import (
	"context"
	"fmt"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &DatasetResource{}
	_ resource.ResourceWithConfigure   = &DatasetResource{}
	_ resource.ResourceWithImportState = &DatasetResource{}
)

func NewDatasetResource() resource.Resource {
	return &DatasetResource{}
}

// DatasetResource manages a Fireworks dataset (used for fine-tuning and
// evaluations). Datasets are immutable after creation, so every attribute
// change forces replacement.
type DatasetResource struct {
	client *client.Client
}

type DatasetResourceModel struct {
	ID           types.String `tfsdk:"id"`
	DatasetID    types.String `tfsdk:"dataset_id"`
	Name         types.String `tfsdk:"name"`
	DisplayName  types.String `tfsdk:"display_name"`
	ExternalURL  types.String `tfsdk:"external_url"`
	Format       types.String `tfsdk:"format"`
	ExampleCount types.Int64  `tfsdk:"example_count"`
	State        types.String `tfsdk:"state"`
}

func (r *DatasetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dataset"
}

func (r *DatasetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A dataset for fine-tuning or evaluation. Attributes are immutable; changes force replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Same as dataset_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dataset_id": schema.StringAttribute{
				Description: "User-chosen dataset identifier.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Full resource name, e.g. accounts/my-account/datasets/my-dataset.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable name.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"external_url": schema.StringAttribute{
				Description: "External URI of the dataset contents, e.g. gs://bucket/data.jsonl.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"format": schema.StringAttribute{
				Description: "Dataset format: CHAT, COMPLETION, or RL.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"example_count": schema.Int64Attribute{
				Description: "Number of examples in the dataset.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Description: "Dataset state (UPLOADING, READY).",
				Computed:    true,
			},
		},
	}
}

func (r *DatasetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (m *DatasetResourceModel) fromAPI(ds *client.Dataset) {
	m.ID = types.StringValue(client.ResourceID(ds.Name))
	m.Name = types.StringValue(ds.Name)
	m.State = types.StringValue(ds.State)
	if ds.DisplayName != "" || !m.DisplayName.IsNull() {
		m.DisplayName = stringOrNull(ds.DisplayName)
	}
	if ds.ExternalURL != "" || !m.ExternalURL.IsNull() {
		m.ExternalURL = stringOrNull(ds.ExternalURL)
	}
	m.Format = stringOrNull(ds.Format)
	if ds.ExampleCount != 0 {
		m.ExampleCount = types.Int64Value(ds.ExampleCount)
	} else if m.ExampleCount.IsUnknown() {
		m.ExampleCount = types.Int64Null()
	}
}

func (r *DatasetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatasetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds := &client.Dataset{
		DisplayName: data.DisplayName.ValueString(),
		ExternalURL: data.ExternalURL.ValueString(),
	}
	if !data.Format.IsUnknown() {
		ds.Format = data.Format.ValueString()
	}
	if !data.ExampleCount.IsNull() && !data.ExampleCount.IsUnknown() {
		ds.ExampleCount = data.ExampleCount.ValueInt64()
	}

	out, err := r.client.CreateDataset(ctx, data.DatasetID.ValueString(), ds)
	if err != nil {
		resp.Diagnostics.AddError("Error creating dataset", err.Error())
		return
	}

	data.fromAPI(out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatasetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatasetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := r.client.GetDataset(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading dataset", err.Error())
		return
	}

	// dataset_id is not returned by the API; derive it from the name.
	data.DatasetID = types.StringValue(client.ResourceID(ds.Name))
	data.fromAPI(ds)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is never called: every attribute is Required/RequiresReplace or
// Computed, so changes always plan as replacement.
func (r *DatasetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatasetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatasetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatasetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDataset(ctx, data.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting dataset", err.Error())
	}
}

func (r *DatasetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
