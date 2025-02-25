package api

import "time"

type NamedEntity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Person struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type PersonWithTeam struct {
	Person
	Teams []NamedEntity `json:"teams"`
}

type TeamMember struct {
	Person
	TeamLead bool `json:"teamLead"`
}

type Team struct {
	NamedEntity
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"createdAt"`
}

type TeamWithMembers struct {
	Team
	Members []TeamMember `json:"members"`
}

type TeamManifest struct {
	TeamID        string
	TeamName      string         `json:"pretty_name"`
	TeamReference string         `json:"external_reference"`
	TechLead      string         `json:"tech_lead"`
	Vendors       map[string]any `json:"vendors"`
}

type Meta struct {
}

type ResponseWithMeta struct {
	Meta Meta `json:"meta"`
}

type FindPeopleResponse struct {
	ResponseWithMeta
	Data []PersonWithTeam `json:"data"`
}

type FindTeamsResponse struct {
	ResponseWithMeta
	Data []Team `json:"data"`
}

type FindTeamResponse struct {
	ResponseWithMeta
	Data TeamWithMembers `json:"data"`
}

type FindTeamManifestResponse struct {
	ResponseWithMeta
	Data map[string]TeamManifest `json:"data"`
}
