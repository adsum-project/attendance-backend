package timetablehandlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetOwnCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := p.timetable.GetOwnCourses(r.Context())
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", courses)
}

func (p *TimetableProvider) GetCourses(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	sortBy := strings.TrimSpace(r.URL.Query().Get("sortBy"))
	sortOrder := strings.TrimSpace(r.URL.Query().Get("sortOrder"))
	if sortBy != "" && sortBy != "courseCode" && sortBy != "courseName" && sortBy != "campus" {
		sortBy = ""
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = ""
	}
	result, err := p.timetable.GetCourses(r.Context(), page, perPage, search, sortBy, sortOrder)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.PaginatedResponseFromResult(w, "", result)
}

func (p *TimetableProvider) GetCourse(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")

	course, err := p.timetable.GetCourse(r.Context(), courseID)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", course)
}

func (p *TimetableProvider) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var req CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	courseID, err := p.timetable.CreateCourse(r.Context(), req.CourseCode, req.CourseName, req.Campus)
	if err != nil {
		response.JsonError(w, err)
		return
	}

	response.Created(w, "Course created", map[string]any{"courseId": courseID})
}

func (p *TimetableProvider) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")

	var req UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := p.timetable.UpdateCourse(r.Context(), courseID, req.CourseCode, req.CourseName, req.Campus); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}

func (p *TimetableProvider) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	courseID := router.PathParam(r, "course_id")

	if err := p.timetable.DeleteCourse(r.Context(), courseID); err != nil {
		response.JsonError(w, err)
		return
	}
	response.NoContent(w)
}
