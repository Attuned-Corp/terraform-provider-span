package api

type FindPeopleRequest struct {
	Email   string
	TeamIDs []string
}

type FindTeamsRequest struct {
	Name string
}

type SetTeamManifestRequest struct {
	Reference string         `json:"externalReference"`
	Vendors   map[string]any `json:"vendors"`
}
