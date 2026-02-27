package timetablehandlers

import (
	"encoding/json"
	"net/http"

	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetClasses(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")

	classes, err := p.timetable.GetClasses(r.Context(), moduleID)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", classes)
}

func (p *TimetableProvider) GetClass(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")
	classID := router.PathParam(r, "class_id")

	class, err := p.timetable.GetClass(r.Context(), moduleID, classID)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", class)
}

func (p *TimetableProvider) CreateClass(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")

	var req CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	classID, err := p.timetable.CreateClass(r.Context(), moduleID, req.ClassName, req.Room, req.DayOfWeek, req.StartsAt, req.EndsAt, req.Recurrence)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.Created(w, "Class created", map[string]any{"classId": classID})
}

func (p *TimetableProvider) UpdateClass(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")
	classID := router.PathParam(r, "class_id")

	var req UpdateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := p.timetable.UpdateClass(r.Context(), moduleID, classID, req.ClassName, req.Room, req.DayOfWeek, req.StartsAt, req.EndsAt, req.Recurrence); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}

func (p *TimetableProvider) DeleteClass(w http.ResponseWriter, r *http.Request) {
	moduleID := router.PathParam(r, "module_id")
	classID := router.PathParam(r, "class_id")

	if err := p.timetable.DeleteClass(r.Context(), moduleID, classID); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}
