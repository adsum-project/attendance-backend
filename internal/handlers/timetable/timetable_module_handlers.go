package timetablehandlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetModules(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	sortBy := strings.TrimSpace(r.URL.Query().Get("sortBy"))
	sortOrder := strings.TrimSpace(r.URL.Query().Get("sortOrder"))
	if sortBy != "" && sortBy != "moduleCode" && sortBy != "moduleName" && sortBy != "startDate" && sortBy != "endDate" {
		sortBy = ""
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = ""
	}
	result, err := p.timetable.GetModules(r.Context(), page, perPage, search, sortBy, sortOrder)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.PaginatedResponseFromResult(w, "", result)
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
