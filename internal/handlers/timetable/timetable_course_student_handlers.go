package timetablehandlers

import (
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

	if err := p.timetable.AssignStudentToCourse(r.Context(), courseID, userID); err != nil {
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
