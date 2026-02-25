package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/adsum-project/attendance-backend/pkg/utils"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
)

func (g *GraphService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	roleMap, spID, err := g.getAppRoleMap(ctx)
	if err != nil {
		return nil, err
	}

	appRoleIds, err := g.getAppRoleAssignments(ctx, userID, spID)
	if err != nil {
		return nil, err
	}

	var roles []string
	for _, id := range appRoleIds {
		if value, ok := roleMap[id]; ok && value != "" {
			roles = append(roles, value)
		}
	}
	return roles, nil
}

func (g *GraphService) getAppRoleMap(ctx context.Context) (map[string]string, string, error) {
	clientID := os.Getenv("ENTRA_CLIENT_ID")
	path := fmt.Sprintf("/servicePrincipals(appId='%s')?$select=id,appRoles", clientID)

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, graphBaseURL, path, nil, nil, nil)
	if err != nil {
		return nil, "", errs.BadGateway("graph service principal: " + err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return nil, "", errs.BadGateway("graph service principal: " + fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))
	}

	var sp servicePrincipalResponse
	if err := json.Unmarshal(res.Body, &sp); err != nil {
		return nil, "", errs.BadGateway("graph service principal decode: " + err.Error())
	}
	if sp.ID == "" {
		return nil, "", errs.BadGateway("service principal not found for app")
	}

	roleMap := make(map[string]string)
	for _, r := range sp.AppRoles {
		roleMap[r.ID] = r.Value
	}
	return roleMap, sp.ID, nil
}

func (g *GraphService) getAppRoleAssignments(ctx context.Context, userID, resourceID string) ([]string, error) {
	path := fmt.Sprintf("/users/%s/appRoleAssignments", url.PathEscape(userID))
	query := url.Values{}
	query.Set("$filter", fmt.Sprintf("resourceId eq %s", resourceID))

	res, err := utils.RequestUpstream(ctx, g.httpClient, http.MethodGet, graphBaseURL, path, query, nil, nil)
	if err != nil {
		return nil, errs.BadGateway("graph app role assignments: " + err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return nil, errs.BadGateway("graph app role assignments: " + fmt.Sprintf("%d %s", res.StatusCode, http.StatusText(res.StatusCode)))
	}

	var data appRoleAssignmentsResponse
	if err := json.Unmarshal(res.Body, &data); err != nil {
		return nil, errs.BadGateway("graph app role assignments decode: " + err.Error())
	}

	var appRoleIds []string
	for _, a := range data.Value {
		if a.AppRoleId != "00000000-0000-0000-0000-000000000000" {
			appRoleIds = append(appRoleIds, a.AppRoleId)
		}
	}
	return appRoleIds, nil
}
