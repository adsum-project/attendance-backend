package userhandlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	graphmodels "github.com/adsum-project/attendance-backend/internal/models/graph"
	usermodels "github.com/adsum-project/attendance-backend/internal/models/user"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	"github.com/adsum-project/attendance-backend/pkg/utils/pagination"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *UserProvider) GetUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	sortBy := strings.TrimSpace(r.URL.Query().Get("sortBy"))
	sortOrder := strings.TrimSpace(r.URL.Query().Get("sortOrder"))
	if sortBy != "" && sortBy != "displayName" && sortBy != "mail" {
		sortBy = ""
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = ""
	}
	fetch := func(ctx context.Context, pg, perPg int) ([]graphmodels.GraphUser, error) {
		return p.graph.GetUsers(ctx, pg, perPg, search, sortBy, sortOrder)
	}
	count := func(ctx context.Context) (int, error) {
		return p.graph.GetUsersCount(ctx, search)
	}
	result, err := pagination.Paginate(r.Context(), page, perPage, fetch, count)
	if err != nil {
		response.JsonError(w, errs.BadGateway("Failed to fetch users from directory: "+err.Error()))
		return
	}
	users := make([]usermodels.User, len(result.Data))
	for i := range result.Data {
		role := "default"
		if roles, err := p.graph.GetUserRoles(r.Context(), result.Data[i].ID); err == nil {
			for _, r := range roles {
				if r == "admin" {
					role = "admin"
					break
				}
				if r == "staff" {
					role = "staff"
				}
			}
		}
		users[i] = usermodels.User{
			UserID:      result.Data[i].ID,
			DisplayName: result.Data[i].DisplayName,
			Email:       result.Data[i].Mail,
			Role:        role,
		}
	}
	response.PaginatedResponse(w, "", users, result.Page, result.PerPage, result.Total)
}
