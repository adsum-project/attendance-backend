package timetablehandlers

import (
	"encoding/json"
	"net/http"
	"strings"
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

func (p *TimetableProvider) GetNodeTimetable(w http.ResponseWriter, r *http.Request) {
	room := strings.TrimSpace(r.URL.Query().Get("room"))
	if room == "" {
		response.JsonError(w, errs.BadRequest("room query parameter is required"))
		return
	}
	classes, err := p.timetable.GetNodeTimetable(r.Context(), room)
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", classes)
}

func (p *TimetableProvider) GetNodeRoom(w http.ResponseWriter, r *http.Request) {
	room, err := p.timetable.GetNodeRoom(r.Context())
	if err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", map[string]string{"room": room})
}

func (p *TimetableProvider) UpdateNodeRoom(w http.ResponseWriter, r *http.Request) {
	var req NodeRoomAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JsonError(w, errs.BadRequest("invalid request body"))
		return
	}
	room := strings.TrimSpace(req.Room)
	if room == "" {
		response.JsonError(w, errs.BadRequest("room is required"))
		return
	}
	if err := p.timetable.UpdateNodeRoom(r.Context(), room); err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "", map[string]string{"room": room})
}

func (p *TimetableProvider) DeleteNodeRoom(w http.ResponseWriter, r *http.Request) {
	if err := p.timetable.DeleteNodeRoom(r.Context()); err != nil {
		response.JsonError(w, err)
		return
	}
	response.OK(w, "Room assignment removed", nil)
}
