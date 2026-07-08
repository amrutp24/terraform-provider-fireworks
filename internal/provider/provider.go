package provider

import (
	"context"
	"os"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &FireworksProvider{}

// FireworksProvider manages Fireworks AI control-plane resources
// (deployments, datasets) via https://api.fireworks.ai.
type FireworksProvider struct {
	version string
}

type FireworksProviderModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	AccountID types.String `tfsdk:"account_id"`
	BaseURL   types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FireworksProvider{version: version}
	}
}

func (p *FireworksProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fireworks"
	resp.Version = p.version
}

func (p *FireworksProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Fireworks AI resources: GPU inference deployments and datasets.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Fireworks API key. Can also be set via the FIREWORKS_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"account_id": schema.StringAttribute{
				Description: "Fireworks account ID. Can also be set via the FIREWORKS_ACCOUNT_ID environment variable.",
				Optional:    true,
			},
			"base_url": schema.StringAttribute{
				Description: "Override the API base URL. Defaults to https://api.fireworks.ai. Mainly for testing.",
				Optional:    true,
			},
		},
	}
}

func (p *FireworksProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data FireworksProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := data.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("FIREWORKS_API_KEY")
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Fireworks API key",
			"Set the api_key provider attribute or the FIREWORKS_API_KEY environment variable.",
		)
	}

	accountID := data.AccountID.ValueString()
	if accountID == "" {
		accountID = os.Getenv("FIREWORKS_ACCOUNT_ID")
	}
	if accountID == "" {
		resp.Diagnostics.AddError(
			"Missing Fireworks account ID",
			"Set the account_id provider attribute or the FIREWORKS_ACCOUNT_ID environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(data.BaseURL.ValueString(), accountID, apiKey)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *FireworksProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDeploymentResource,
		NewDatasetResource,
		NewSFTJobResource,
	}
}

func (p *FireworksProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeploymentDataSource,
		NewModelDataSource,
	}
}
