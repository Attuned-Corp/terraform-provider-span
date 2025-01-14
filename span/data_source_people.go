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

var _ datasource.DataSource = &PersonDataSource{}

func NewPeopleDataSource() datasource.DataSource {
	return &PeopleDataSource{}
}

type PeopleResourceData struct {
	People types.List `tfsdk:"people"`
}

// PersonDataSource is the concrete implementation
type PeopleDataSource struct {
	apiClient api.SpanAPIClient
}

func (d *PeopleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_people"
}

func (d *PeopleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "A data source representation of multiple people within Span.",
		Attributes: map[string]schema.Attribute{
			"people": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Complete list of people within Span.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: PersonResourceData{}.Attributes(),
				},
			},
		},
	}
}

func (d *PeopleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func newPeopleResourceData(ctx context.Context, in []api.PersonWithTeam, diags *diag.Diagnostics) PeopleResourceData {
	var data PeopleResourceData

	if len(in) == 0 {
		return data
	}

	people := make([]PersonResourceData, len(in))
	for i, incoming := range in {
		people[i] = newPersonResourceData(ctx, &incoming)
	}

	result, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: PersonResourceData{}.AttrTypes()}, people)

	diags.Append(d...)

	data.People = result

	return data
}

func (d *PeopleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PeopleResourceData

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	response, err := d.apiClient.FindPeople(api.FindPeopleRequest{})
	if err != nil {
		resp.Diagnostics.AddError("Unexpected API error", fmt.Sprintf("Raw: %s\n", err.Error()))
		return
	}

	data = newPeopleResourceData(ctx, response, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
