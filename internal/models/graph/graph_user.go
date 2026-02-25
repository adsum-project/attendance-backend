package graphmodels

type GraphUser struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
	JobTitle          string `json:"jobTitle"`
}
