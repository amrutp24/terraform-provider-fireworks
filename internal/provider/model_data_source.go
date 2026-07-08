package provider

import (
	"context"
	"fmt"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &ModelDataSource{}
	_ datasource.DataSourceWithConfigure = &ModelDataSource{}
)

func NewModelDataSource() datasource.DataSource {
	return &ModelDataSource{}
}

// ModelDataSource looks up a model in the configured account by ID —
// typically the output model of a fine-tuning job or an uploaded model.
type ModelDataSource struct {
	client *client.Client
}

type ModelDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	State       types.String `tfsdk:"state"`
	Kind        types.String `tfsdk:"kind"`
	CreateTime  types.String `tfsdk:"create_time"`
}

func (d *ModelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_model"
}

func (d *ModelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up a model in the configured account by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Model ID (last segment of the resource name).",
				Required:    true,
			},
			"name":         schema.StringAttribute{Computed: true, Description: "Full resource name."},
			"display_name": schema.StringAttribute{Computed: true, Description: "Human-readable name."},
			"description":  schema.StringAttribute{Computed: true, Description: "Model description."},
			"state":        schema.StringAttribute{Computed: true, Description: "Model state (UPLOADING, READY)."},
			"kind":         schema.StringAttribute{Computed: true, Description: "Model kind, e.g. HF_BASE_MODEL, HF_PEFT_ADDON."},
			"create_time":  schema.StringAttribute{Computed: true, Description: "Creation timestamp (RFC 3339)."},
		},
	}
}

func (d *ModelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *ModelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model, err := d.client.GetModel(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading model", err.Error())
		return
	}

	data.Name = types.StringValue(model.Name)
	data.DisplayName = stringOrNull(model.DisplayName)
	data.Description = stringOrNull(model.Description)
	data.State = types.StringValue(model.State)
	data.Kind = types.StringValue(model.Kind)
	data.CreateTime = stringOrNull(model.CreateTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
