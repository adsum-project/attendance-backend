package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	graphmodels "github.com/adsum-project/attendance-backend/internal/models/graph"
	"github.com/adsum-project/attendance-backend/pkg/utils"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
)

func (g *GraphService) GetUser(ctx context.Context, userID string) (*graphmodels.GraphUser, error) {
	path := "/users/" + url.PathEscape(userID)
	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, graphBaseURL, path, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("graph user: %w", err)
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graph user: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	var user graphmodels.GraphUser
	if err := json.Unmarshal(res.Body, &user); err != nil {
		return nil, fmt.Errorf("graph user decode: %w", err)
	}
	return &user, nil
}

func (g *GraphService) GetUsersCount(ctx context.Context) (int, error) {
	path := "/users?$top=1&$count=true&$orderby=displayName"
	headers := http.Header{}
	headers.Set("ConsistencyLevel", "eventual")

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, graphBaseURL, path, nil, headers, nil)
	if err != nil {
		return 0, fmt.Errorf("graph users count: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("graph users count: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	var data usersResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return 0, fmt.Errorf("graph users count decode: %w", err)
	}
	return data.Count, nil
}

func (g *GraphService) GetUsers(ctx context.Context, page, perPage int) ([]graphmodels.GraphUser, error) {
	path := fmt.Sprintf("/users?$top=%d&$count=true&$orderby=displayName", perPage)
	headers := http.Header{}
	headers.Set("ConsistencyLevel", "eventual")

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, graphBaseURL, path, nil, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("graph users: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graph users: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	var data usersResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return nil, fmt.Errorf("graph users decode: %w", err)
	}

	users := data.Value
	nextLink := data.NextLink

	for i := 1; i < page; i++ {
		if nextLink == "" {
			return []graphmodels.GraphUser{}, nil
		}
		pageUsers, next, err := g.getUsersFromURL(ctx, nextLink)
		if err != nil {
			return nil, err
		}
		users = pageUsers
		nextLink = next
	}

	return users, nil
}

func (g *GraphService) getUsersFromURL(ctx context.Context, pageURL string) ([]graphmodels.GraphUser, string, error) {
	// nextLink is a full URL; pass empty baseURL so path is used as-is
	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, "", pageURL, nil, nil, nil)
	if err != nil {
		return nil, "", err
	}
	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("graph users page: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}
	var data usersResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return nil, "", err
	}
	return data.Value, data.NextLink, nil
}

func (g *GraphService) GetUsersByIDs(ctx context.Context, ids []string) ([]graphmodels.GraphUser, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	body := getByIdsRequest{
		IDs:   ids,
		Types: []string{"user"},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodPost, graphBaseURL, "/directoryObjects/getByIds", nil, headers, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, errs.BadGateway("graph getByIds: " + err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return nil, errs.BadGateway("graph getByIds: " + fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))
	}
	var data getByIdsResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return nil, errs.BadGateway("graph getByIds decode: " + err.Error())
	}
	return data.Value, nil
}
