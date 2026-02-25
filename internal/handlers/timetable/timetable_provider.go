package timetablehandlers

import (
	"fmt"

	"github.com/adsum-project/attendance-backend/internal/services/timetable"
)

type TimetableProvider struct {
	timetable *timetable.TimetableService
}

func NewTimetableProvider(t *timetable.TimetableService) (*TimetableProvider, error) {
	if t == nil {
		return nil, fmt.Errorf("timetable service is required")
	}

	return &TimetableProvider{
		timetable: t,
	}, nil
}
