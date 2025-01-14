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

var _ datasource.DataSource = &PersonDataSource{}

func NewPersonDataSource() datasource.DataSource {
	return &PersonDataSource{}
}

// PersonDataSource is the concrete implementation
type PersonDataSource struct {
	apiClient api.SpanAPIClient
}

type PersonResourceData struct {
	Email types.String `tfsdk:"email"`
	Name  types.String `tfsdk:"name"`
	Teams types.List   `tfsdk:"teams"`
}

func (pr PersonResourceData) Attributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"email": schema.StringAttribute{
			MarkdownDescription: "The email for the specific person",
			Optional:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Full name of the person.",
			Optional:            true,
		},
		"teams": schema.ListNestedAttribute{
			Computed:            true,
			MarkdownDescription: "The list of teams of which the person is part of.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required: true,
					},
					"name": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func (pr PersonResourceData) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"email": types.StringType,
		"name":  types.StringType,
		"teams": types.ListType{ElemType: types.ObjectType{AttrTypes: PersonTeam{}.AttrTypes()}},
	}
}

type PersonTeam struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (pt PersonTeam) AttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
	}
}

func newPersonTeamList(ctx context.Context, in []api.NamedEntity, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: PersonTeam{}.AttrTypes()})
	}

	personTeams := make([]PersonTeam, len(in))
	for i, incoming := range in {
		personTeams[i].ID = types.StringValue(incoming.ID)
		personTeams[i].Name = types.StringValue(incoming.Name)
	}

	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: PersonTeam{}.AttrTypes()}, personTeams)

	diags.Append(d...)

	return result
}

func newPersonResourceData(ctx context.Context, in *api.PersonWithTeam) PersonResourceData {
	var data PersonResourceData

	// @TODO: Diagnostics handling for unmarshal
	var d diag.Diagnostics

	data.Email = types.StringValue(in.Email)
	data.Name = types.StringValue(in.Name)
	data.Teams = newPersonTeamList(ctx, in.Teams, &d)

	return data
}

func (d *PersonDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_person"
}

func (d *PersonDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A data source representation for people within span.",
		Attributes:          PersonResourceData{}.Attributes(),
	}
}

func (d *PersonDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PersonDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PersonResourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Email.IsNull() {
		resp.Diagnostics.AddError("Missing required parameter for person data source loading...", "")
		return
	}

	response, err := d.apiClient.FindPeople(api.FindPeopleRequest{Email: data.Email.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Unexpected API error", fmt.Sprintf("Raw: %s\n", err.Error()))
		return
	}

	if len(response) == 0 {
		resp.Diagnostics.AddError("Missing data source", fmt.Sprintf("Could not load data source for person with email %s", data.Email.ValueString()))
		return
	}

	if len(response) > 1 {
		resp.Diagnostics.AddError("Multiple matches found where single result expected", fmt.Sprintf("Multiple results for user with e-mail %s", data.Email.ValueString()))
		return
	}

	data = newPersonResourceData(ctx, &response[0])

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
