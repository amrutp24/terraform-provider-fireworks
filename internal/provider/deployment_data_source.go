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
	_ datasource.DataSource              = &DeploymentDataSource{}
	_ datasource.DataSourceWithConfigure = &DeploymentDataSource{}
)

func NewDeploymentDataSource() datasource.DataSource {
	return &DeploymentDataSource{}
}

// DeploymentDataSource looks up an existing deployment by ID.
type DeploymentDataSource struct {
	client *client.Client
}

type DeploymentDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	BaseModel       types.String `tfsdk:"base_model"`
	DisplayName     types.String `tfsdk:"display_name"`
	MinReplicaCount types.Int64  `tfsdk:"min_replica_count"`
	MaxReplicaCount types.Int64  `tfsdk:"max_replica_count"`
	AcceleratorType types.String `tfsdk:"accelerator_type"`
	Precision       types.String `tfsdk:"precision"`
	State           types.String `tfsdk:"state"`
}

func (d *DeploymentDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (d *DeploymentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Look up an existing deployment by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Deployment ID.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Full resource name, e.g. accounts/my-account/deployments/abc123.",
				Computed:    true,
			},
			"base_model": schema.StringAttribute{
				Description: "Model the deployment serves.",
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable name.",
				Computed:    true,
			},
			"min_replica_count": schema.Int64Attribute{
				Description: "Minimum replicas; 0 means scale-to-zero is enabled.",
				Computed:    true,
			},
			"max_replica_count": schema.Int64Attribute{
				Description: "Maximum replicas for autoscaling.",
				Computed:    true,
			},
			"accelerator_type": schema.StringAttribute{
				Description: "GPU type serving the deployment.",
				Computed:    true,
			},
			"precision": schema.StringAttribute{
				Description: "Model serving precision.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "Current deployment state (CREATING, READY, UPDATING, DELETING, FAILED).",
				Computed:    true,
			},
		},
	}
}

func (d *DeploymentDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeploymentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dep, err := d.client.GetDeployment(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployment", err.Error())
		return
	}

	data.Name = types.StringValue(dep.Name)
	data.BaseModel = types.StringValue(dep.BaseModel)
	data.DisplayName = stringOrNull(dep.DisplayName)
	data.MinReplicaCount = int64OrNull(dep.MinReplicaCount)
	data.MaxReplicaCount = int64OrNull(dep.MaxReplicaCount)
	data.AcceleratorType = stringOrNull(dep.AcceleratorType)
	data.Precision = stringOrNull(dep.Precision)
	data.State = types.StringValue(dep.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
