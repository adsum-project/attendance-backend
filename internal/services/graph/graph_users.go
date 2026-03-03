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

// GetUser fetches a user by ID from Microsoft Graph. Returns nil if not found.
func (g *GraphService) GetUser(ctx context.Context, userID string) (*graphmodels.GraphUser, error) {
	path := "/users/" + url.PathEscape(userID)
	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, g.baseURL, path, nil, nil, nil)
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

func (g *GraphService) GetUsersCount(ctx context.Context, search string) (int, error) {
	path := "/users?$top=1&$count=true&$orderby=displayName"
	if search != "" {
		searchClause := fmt.Sprintf("\"displayName:%s\" OR \"mail:%s\"", search, search)
		path = "/users?$search=" + url.QueryEscape(searchClause) + "&$top=1&$count=true&$orderby=displayName"
	}
	headers := http.Header{}
	headers.Set("ConsistencyLevel", "eventual")

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, g.baseURL, path, nil, headers, nil)
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

// GetUsers fetches a page of users from Graph. Supports search (displayName/mail) and sort (displayName/mail).
func (g *GraphService) GetUsers(ctx context.Context, page, perPage int, search, sortBy, sortOrder string) ([]graphmodels.GraphUser, error) {
	orderBy := "displayName"
	if sortBy == "mail" {
		orderBy = "mail"
	}
	if sortOrder == "desc" {
		orderBy += " desc"
	}
	path := fmt.Sprintf("/users?$top=%d&$count=true&$orderby=%s", perPage, url.QueryEscape(orderBy))
	if search != "" {
		searchClause := fmt.Sprintf("\"displayName:%s\" OR \"mail:%s\"", search, search)
		path = fmt.Sprintf("/users?$search=%s&$top=%d&$count=true&$orderby=%s",
			url.QueryEscape(searchClause), perPage, url.QueryEscape(orderBy))
	}
	headers := http.Header{}
	headers.Set("ConsistencyLevel", "eventual")

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, g.baseURL, path, nil, headers, nil)
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

// GetUsersByIDs fetches multiple users by ID in a single Graph API call.
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

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodPost, g.baseURL, "/directoryObjects/getByIds", nil, headers, bytes.NewReader(jsonBody))
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
