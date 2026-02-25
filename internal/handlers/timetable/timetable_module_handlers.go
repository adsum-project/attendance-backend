package timetablehandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetModules(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	modules, total, page, perPage, err := p.timetable.GetModules(r.Context(), page, perPage)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.PaginatedResponse(w, "", modules, page, perPage, total)
}

func (p *TimetableProvider) GetModule(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")

	module, err := p.timetable.GetModule(r.Context(), moduleID)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", module)
}

func (p *TimetableProvider) CreateModule(w http.ResponseWriter, r *http.Request) {
	var req CreateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	moduleID, err := p.timetable.CreateModule(r.Context(), req.ModuleCode, req.ModuleName, req.StartDate, req.EndDate)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.Created(w, "Module created", map[string]any{"moduleId": moduleID})
}

func (p *TimetableProvider) UpdateModule(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")

	var req UpdateModuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := p.timetable.UpdateModule(r.Context(), moduleID, req.ModuleCode, req.ModuleName, req.StartDate, req.EndDate); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}

func (p *TimetableProvider) DeleteModule(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")

	if err := p.timetable.DeleteModule(r.Context(), moduleID); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}

func (p *TimetableProvider) CreateModuleEnrollment(w http.ResponseWriter, r *http.Request) {
	actorUserID, _ := r.Context().Value("userID").(string)
	if actorUserID == "" {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	_ = router.PathParam(r, "student_user_id")
	_ = router.PathParam(r, "module_id")

	response.JsonError(w, response.HttpErrorMessage{
		StatusCode: http.StatusNotImplemented,
		Error:      "CreateModuleEnrollment not implemented",
	})
}

func (p *TimetableProvider) DeleteModuleEnrollment(w http.ResponseWriter, r *http.Request) {
	actorUserID, _ := r.Context().Value("userID").(string)
	if actorUserID == "" {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	_ = router.PathParam(r, "student_user_id")
	_ = router.PathParam(r, "module_id")

	response.JsonError(w, response.HttpErrorMessage{
		StatusCode: http.StatusNotImplemented,
		Error:      "DeleteModuleEnrollment not implemented",
	})
}
