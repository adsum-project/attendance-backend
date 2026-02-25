package userhandlers

import (
	"net/http"
	"strconv"

	usermodels "github.com/adsum-project/attendance-backend/internal/models/user"
	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	"github.com/adsum-project/attendance-backend/pkg/utils/pagination"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *UserProvider) GetUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	graphUsers, total, page, perPage, err := pagination.Paginate(r.Context(), page, perPage, p.graph.GetUsers, p.graph.GetUsersCount)
	if err != nil {
		response.JsonError(w, errs.BadGateway("Failed to fetch users from directory: "+err.Error()))
		return
	}
	users := make([]usermodels.User, len(graphUsers))
	for i := range graphUsers {
		users[i] = usermodels.User{
			UserID:      graphUsers[i].ID,
			DisplayName: graphUsers[i].DisplayName,
			Email:       graphUsers[i].Mail,
		}
	}
	response.PaginatedResponse(w, "", users, page, perPage, total)
}
