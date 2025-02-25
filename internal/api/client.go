package api

import (
	"github.com/imroc/req/v3"
)

const (
	DefaultEndpoint = "https://span.app/api/external/v1"
)

type SpanAPIClient interface {
	FindPeople(r FindPeopleRequest) ([]PersonWithTeam, error)
	FindTeams(r FindTeamsRequest) ([]Team, error)
	FindTeamByID(teamID string) (*TeamWithMembers, error)
	FindTeamManifestByTeamID(teamID string) (*TeamManifest, error)
}

type client struct {
	endpoint   string
	token      string
	httpClient *req.Client
}

func (c *client) FindPeople(r FindPeopleRequest) ([]PersonWithTeam, error) {
	var resp FindPeopleResponse

	request := c.httpClient.Get("/catalog/people")

	if len(r.TeamIDs) > 0 {
		request.AddQueryParams("teamIds", r.TeamIDs...)
	}

	if r.Email != "" {
		request.AddQueryParam("email", r.Email)
	}

	err := request.
		Do().
		Into(&resp)

	if err != nil {
		return nil, NewUnknownError()
	}

	return resp.Data, nil
}

func (c *client) FindTeams(r FindTeamsRequest) ([]Team, error) {
	var resp FindTeamsResponse

	request := c.httpClient.Get("/catalog/teams")

	if r.Name != "" {
		request.AddQueryParam("name", r.Name)
	}

	err := request.
		Do().
		Into(&resp)

	if err != nil {
		return nil, NewUnknownError()
	}

	return resp.Data, nil
}

func (c *client) FindTeamByID(teamID string) (*TeamWithMembers, error) {
	var resp FindTeamResponse

	err := c.httpClient.Get("/catalog/teams/{teamID}").
		SetPathParam("teamID", teamID).
		Do().
		Into(&resp)

	if err != nil {
		return nil, NewUnknownError()
	}

	return &resp.Data, nil
}

func (c *client) FindTeamManifestByTeamID(teamID string) (*TeamManifest, error) {
	var resp FindTeamManifestResponse

	err := c.httpClient.Get("/catalog/teams/{teamID}/manifest").
		SetPathParam("teamID", teamID).
		Do().
		Into(&resp)

	if err != nil {
		return nil, NewUnknownError()
	}

	var manifest *TeamManifest
	for k, m := range resp.Data {
		manifest = &m
		manifest.TeamReference = k
		manifest.TeamID = teamID
	}

	return manifest, nil

}

type clientOptions struct {
	endpoint string
	token    string
}

type ClientOption func(*clientOptions) *clientOptions

func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) *clientOptions {
		o.endpoint = endpoint
		return o
	}
}

func WithToken(token string) ClientOption {
	return func(o *clientOptions) *clientOptions {
		o.token = token
		return o
	}
}

// NewSpanAPIClient instantiates a new client able to connect to the SPAN api
func NewSpanAPIClient(opt ...ClientOption) (SpanAPIClient, error) {
	opts := &clientOptions{
		endpoint: DefaultEndpoint,
	}

	for _, funcOpt := range opt {
		opts = funcOpt(opts)
	}

	return &client{
		endpoint:   opts.endpoint,
		token:      opts.token,
		httpClient: req.C().SetBaseURL(opts.endpoint).SetCommonBearerAuthToken(opts.token),
	}, nil
}
