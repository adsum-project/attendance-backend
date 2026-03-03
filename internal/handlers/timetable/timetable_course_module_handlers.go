package timetablehandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetModuleCourses(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))

	result, err := p.timetable.GetModuleCourses(r.Context(), moduleID, page, perPage)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.PaginatedResponseFromResult(w, "", result)
}

func (p *TimetableProvider) GetCourseModules(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))

	result, err := p.timetable.GetCourseModules(r.Context(), courseID, page, perPage)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.PaginatedResponseFromResult(w, "", result)
}

func (p *TimetableProvider) AssignModuleToCourse(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")
	moduleID := router.PathParam(r, "module_id")

	var req AssignModuleToCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := p.timetable.AssignModuleToCourse(r.Context(), courseID, moduleID, req.YearOfStudy); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}

func (p *TimetableProvider) UnassignModuleFromCourse(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")
	moduleID := router.PathParam(r, "module_id")

	if err := p.timetable.UnassignModuleFromCourse(r.Context(), courseID, moduleID); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}
