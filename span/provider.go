package span

import (
	"context"
	"fmt"
	"os"

	"github.com/attuned-corp/terraform-provider-span/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type spanProvider struct {
	version string
}

var (
	_ provider.Provider = &spanProvider{}
)

// NewProvider is our main factory instantiation function
func NewProviderFactory(version string) func() provider.Provider {
	return func() provider.Provider {
		return &spanProvider{
			version: version,
		}
	}
}

// Metadata is a call that auto-populates resource/datasource.{MetadataRequest} properties
// Metadata should return the metadata for the provider, such as
// a type name and version data.
func (p *spanProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "span"
	resp.Version = p.version
}

// Schema returns a complete schema for the provider.
func (p *spanProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				Description: "A Span PAT for API authn & authz.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_endpoint": schema.StringAttribute{
				Description: "Span API's base endpoint for client communication.",
				Optional:    true,
			},
		},
	}
}

// PluginProviderConfiguration describes the provider data model.
type ProviderConfiguration struct {
	AccessToken types.String `tfsdk:"access_token"`
	APIEndpoint types.String `tfsdk:"api_endpoint"`
}

// Configure is a start of lifecycle hook which terraform uses to insert all values
// at instantiation. We are going to initialize & inject our API client
func (p *spanProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg ProviderConfiguration

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("SPAN_ACCESS_TOKEN")
	if cfg.AccessToken.ValueString() != "" {
		token = cfg.AccessToken.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing Span Access token",
			"The SPAN_ACCESS_TOKEN was not correctly initialized. It needs to be provided within a configuration block or via the environment.",
		)
		return
	} else if len(token) != 64 {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Incorrect Span Access token",
			fmt.Sprintf("The SPAN_ACCESS_TOKEN needs to be exactly 64 characters. Incorrect length encountered [%d]", len(token)),
		)
		return
	}

	fnOpts := []api.ClientOption{api.WithToken(token)}

	endpoint := os.Getenv("SPAN_API_ENDPOINT")
	if cfg.APIEndpoint.ValueString() != "" {
		endpoint = cfg.APIEndpoint.ValueString()
	}

	if endpoint != "" {
		fnOpts = append(fnOpts, api.WithEndpoint(endpoint))
	}

	client, err := api.NewSpanAPIClient(fnOpts...)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed instantiating Span API client",
			fmt.Sprintf("Unexpected error %s", err.Error()),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "terraform-provider-span - version information", map[string]any{"version": p.version})
}

// DataSources returns a slice of functions to instantiate supported data source callouts.
func (p *spanProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPersonDataSource,
		NewPeopleDataSource,
		NewTeamDataSource,
		NewTeamsDataSource,
	}
}

// Resources returns a slice of functions to instantiate supported Resource implementations
func (p *spanProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
