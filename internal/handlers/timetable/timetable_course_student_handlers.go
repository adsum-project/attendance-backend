package timetablehandlers

import (
	"encoding/json"
	"net/http"

	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetCourseStudents(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")

	students, err := p.timetable.GetCourseStudents(r.Context(), courseID)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", students)
}

func (p *TimetableProvider) AssignStudentToCourse(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")
	userID := router.PathParam(r, "user_id")

	var req AssignStudentToCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}
	if req.YearOfStudy == 0 {
		req.YearOfStudy = 1
	}

	if err := p.timetable.AssignStudentToCourse(r.Context(), courseID, userID, req.YearOfStudy); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}

func (p *TimetableProvider) UnassignStudentFromCourse(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")
	userID := router.PathParam(r, "user_id")

	if err := p.timetable.UnassignStudentFromCourse(r.Context(), courseID, userID); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}
