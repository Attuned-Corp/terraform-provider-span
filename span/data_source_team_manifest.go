package span

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/attuned-corp/terraform-provider-span/internal/api"
	dynamic "github.com/attuned-corp/terraform-provider-span/span/internal/serde"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &TeamManifestDataSource{}

func NewTeamManifestDataSource() datasource.DataSource {
	return &TeamManifestDataSource{}
}

// TeamManifestDataSource is the implementation for a manifest.
type TeamManifestDataSource struct {
	apiClient api.SpanAPIClient
}

type teamManifestDataSourceData struct {
	TeamID    types.String  `tfsdk:"team_id"`
	TeamName  types.String  `tfsdk:"team_name"`
	Reference types.String  `tfsdk:"reference"`
	TechLead  types.String  `tfsdk:"tech_lead"`
	Vendors   types.Dynamic `tfsdk:"vendors"`
}

func newTeamManifestDataSourceData(_ context.Context, in *api.TeamManifest) (*teamManifestDataSourceData, error) {
	var data teamManifestDataSourceData

	data.TeamID = types.StringValue(in.TeamID)
	data.TeamName = types.StringValue(in.TeamName)
	data.Reference = types.StringValue(in.TeamReference)
	data.TechLead = types.StringValue(in.TechLead)

	// Could not find a way to do this conversion without additional serde
	// full step. ;o/
	vendorsInput, err := json.Marshal(in.Vendors)
	if err != nil {
		return nil, err
	}

	data.Vendors, err = dynamic.FromJSON(vendorsInput)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (tmr teamManifestDataSourceData) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"team_id": schema.StringAttribute{
			MarkdownDescription: "The team id owner for the manifest resource.",
			Optional:            true,
		},
		"team_name": schema.StringAttribute{
			MarkdownDescription: "The name of the team owner for the manifest resource.",
			Optional:            true,
		},
		"reference": schema.StringAttribute{
			MarkdownDescription: "Human formatted reference for the team.",
			Optional:            true,
		},
		"tech_lead": schema.StringAttribute{
			MarkdownDescription: "Email of the tech lead for said team.",
			Optional:            true,
			Computed:            true,
		},
		"vendors": schema.DynamicAttribute{
			Optional: true,
		},
	}
}

func (d *TeamManifestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_manifest"
}

func (d *TeamManifestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A data source representation for manifest details stored for a team resource.",
		Attributes:          teamManifestDataSourceData{}.Attributes(),
	}
}

func (d *TeamManifestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamManifestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data teamManifestDataSourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.TeamID.IsNull() {
		resp.Diagnostics.AddError("Missing required parameter - please provide a team ID for the manifest resource", "")
		return
	}

	response, err := d.apiClient.FindTeamManifestByTeamID(data.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unexpected API error", fmt.Sprintf("Raw: %s\n", err.Error()))
		return
	}

	if response == nil {
		resp.Diagnostics.AddError("Missing data source", fmt.Sprintf("Could not load data source for team manifest with team owner id of ID %s", data.TeamID.ValueString()))
		return
	}

	resource, err := newTeamManifestDataSourceData(ctx, response)

	if err != nil {
		resp.Diagnostics.AddError("Could not load manifest", fmt.Sprintf("Schema mapping for manifiest failed with %v", err))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, resource)...)
}
