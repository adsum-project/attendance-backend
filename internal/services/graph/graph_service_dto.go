package graph

import graphmodels "github.com/adsum-project/attendance-backend/internal/models/graph"

type usersResponse struct {
	Value    []graphmodels.GraphUser `json:"value"`
	NextLink string                  `json:"@odata.nextLink"`
	Count    int                     `json:"@odata.count"`
}

type getByIdsRequest struct {
	IDs   []string `json:"ids"`
	Types []string `json:"types"`
}

type getByIdsResponse struct {
	Value []graphmodels.GraphUser `json:"value"`
}

type servicePrincipalResponse struct {
	ID       string `json:"id"`
	AppRoles []struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"appRoles"`
}

type appRoleAssignmentsResponse struct {
	Value []struct {
		AppRoleId string `json:"appRoleId"`
	} `json:"value"`
}
