package span

import (
	"context"
	"fmt"

	"github.com/attuned-corp/terraform-provider-span/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &TeamsDataSource{}

func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

// TeamsDataSource is the concrete implementation
type TeamsDataSource struct {
	apiClient api.SpanAPIClient
}

type TeamsResourceData struct {
	Teams types.List `tfsdk:"teams"`
}

func (d *TeamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *TeamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List of teams available in Span.",
		Attributes: map[string]schema.Attribute{
			"teams": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Complete list of people within Span.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: TeamResourceData{}.Attributes(),
				},
			},
		},
	}
}

func (d *TeamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(api.SpanAPIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected a SpanAPIClient but got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.apiClient = apiClient
}

func newTeamsResourceData(ctx context.Context, in []api.Team, diags *diag.Diagnostics) TeamsResourceData {
	var data TeamsResourceData

	if len(in) == 0 {
		return data
	}

	teams := make([]TeamResourceData, len(in))
	for i, incoming := range in {
		teams[i] = newTeamResourceData(ctx, &incoming)
	}

	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: TeamResourceData{}.AttrTypes()}, teams)

	diags.Append(d...)

	data.Teams = result

	return data
}

func (d *TeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamsResourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.apiClient.FindTeams(api.FindTeamsRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Unexpected API error", fmt.Sprintf("Raw: %s\n", err.Error()))
		return
	}

	data = newTeamsResourceData(ctx, response, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
