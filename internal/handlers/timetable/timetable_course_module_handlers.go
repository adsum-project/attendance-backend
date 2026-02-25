package timetablehandlers

import (
	"encoding/json"
	"net/http"

	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetCourseModules(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")

	modules, err := p.timetable.GetCourseModules(r.Context(), courseID)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", modules)
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
