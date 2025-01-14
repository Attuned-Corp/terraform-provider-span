package api

type FindPeopleRequest struct {
	Email   string
	TeamIDs []string
}

type FindTeamsRequest struct {
	Name string
}
