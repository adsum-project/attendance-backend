package timetablehandlers

import (
	"net/http"
	"time"

	"github.com/adsum-project/attendance-backend/pkg/utils/errs"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *TimetableProvider) GetOwnTimetable(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("weekStart")
	if q == "" {
		response.JsonError(w, errs.BadRequest("weekStart query parameter is required"))
		return
	}
	weekStart, parseErr := time.Parse("2006-01-02", q)
	if parseErr != nil {
		response.JsonError(w, errs.BadRequest("weekStart must be a valid date (YYYY-MM-DD)"))
		return
	}
	classes, err := p.timetable.GetOwnTimetable(r.Context(), weekStart)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", classes)
}
