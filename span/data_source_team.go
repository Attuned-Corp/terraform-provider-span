package span

import (
	"context"
	"fmt"

	"github.com/attuned-corp/terraform-provider-span/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &TeamDataSource{}

func NewTeamDataSource() datasource.DataSource {
	return &TeamDataSource{}
}

// TeamDataSource is the concrete implementation
type TeamDataSource struct {
	apiClient api.SpanAPIClient
}

type TeamMember struct {
	Email    types.String `tfsdk:"email"`
	Name     types.String `tfsdk:"name"`
	TeamLead types.Bool   `tfsdk:"team_lead"`
}

func (pt TeamMember) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"email":     types.StringType,
		"name":      types.StringType,
		"team_lead": types.BoolType,
	}
}

type TeamResourceData struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Slug types.String `tfsdk:"slug"`
}

func (tr TeamResourceData) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Immutable ID for the Span team resource",
			Optional:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Name of the team.",
			Optional:            true,
		},
		"slug": schema.StringAttribute{
			MarkdownDescription: "URL friendly unique slug for the team.",
			Optional:            true,
		},
	}
}

func (tr TeamResourceData) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
		"slug": types.StringType,
	}
}

type TeamDetailsResourceData struct {
	TeamResourceData
	Members types.List `tfsdk:"members"`
}

func (pr TeamDetailsResourceData) Attributes() map[string]schema.Attribute {
	trAttributes := TeamResourceData{}.Attributes()
	trAttributes["members"] = schema.ListNestedAttribute{
		Optional:            true,
		MarkdownDescription: "Inline list of team members with roles.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"email": schema.StringAttribute{
					Required: true,
				},
				"name": schema.StringAttribute{
					Required: true,
				},
				"team_lead": schema.BoolAttribute{
					Required: true,
				},
			},
		},
	}

	return trAttributes
}

func (pr TeamDetailsResourceData) AttrTypes() map[string]attr.Type {
	trAttrTypes := TeamResourceData{}.AttrTypes()
	trAttrTypes["members"] = types.ListType{ElemType: types.ObjectType{AttrTypes: TeamMember{}.AttrTypes()}}
	return trAttrTypes
}

func newTeamMembers(ctx context.Context, in []api.TeamMember, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: TeamMember{}.AttrTypes()})
	}

	teamMembers := make([]TeamMember, len(in))
	for i, incoming := range in {
		teamMembers[i].Email = types.StringValue(incoming.Email)
		teamMembers[i].Name = types.StringValue(incoming.Name)
		teamMembers[i].TeamLead = types.BoolValue(incoming.TeamLead)
	}

	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: TeamMember{}.AttrTypes()}, teamMembers)

	diags.Append(d...)

	return result
}

func newTeamResourceData(_ context.Context, in *api.Team) TeamResourceData {
	var data TeamResourceData

	data.ID = types.StringValue(in.ID)
	data.Name = types.StringValue(in.Name)
	data.Slug = types.StringValue(in.Slug)

	return data
}

func newTeamDetailsResourceData(ctx context.Context, in *api.TeamWithMembers) TeamDetailsResourceData {
	var data TeamDetailsResourceData

	// @TODO: Diagnostics handling for unmarshal
	var d diag.Diagnostics

	data.ID = types.StringValue(in.ID)
	data.Name = types.StringValue(in.Name)
	data.Slug = types.StringValue(in.Slug)
	data.Members = newTeamMembers(ctx, in.Members, &d)

	return data
}

func (d *TeamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *TeamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A data source representation for a team within span.",
		Attributes:          TeamDetailsResourceData{}.Attributes(),
	}
}

func (d *TeamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TeamDetailsResourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	teamID := ""

	if !data.Name.IsNull() {
		// Resolve the team by name
		foundTeams, err := d.apiClient.FindTeams(api.FindTeamsRequest{Name: data.Name.ValueString()})
		if err != nil {
			resp.Diagnostics.AddError("Unexpected API error", fmt.Sprintf("Raw: %s\n", err.Error()))
			return
		}

		if len(foundTeams) == 0 {
			resp.Diagnostics.AddError("Missing data source", fmt.Sprintf("Could not load data source for team with name %s", data.Name.ValueString()))
			return
		}

		if len(foundTeams) > 1 {
			resp.Diagnostics.AddError("Multiple matches found where single result expected", fmt.Sprintf("Multiple results for team with name %s", data.Name.ValueString()))
			return
		}

		teamID = foundTeams[0].ID
	}

	if !data.ID.IsNull() && teamID == "" {
		teamID = data.ID.ValueString()
	}

	if teamID == "" {
		resp.Diagnostics.AddError("Missing required parameter for team loading - 'id' or 'name'...", "")
		return
	}

	response, err := d.apiClient.FindTeamByID(teamID)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected API error", fmt.Sprintf("Raw: %s\n", err.Error()))
		return
	}

	if response == nil {
		resp.Diagnostics.AddError("Missing data source", fmt.Sprintf("Could not load data source for team with ID %s", data.ID.ValueString()))
		return
	}

	data = newTeamDetailsResourceData(ctx, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
