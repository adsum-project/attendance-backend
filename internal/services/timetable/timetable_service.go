package timetable

import (
	"errors"

	graphservice "github.com/adsum-project/attendance-backend/internal/services/graph"
	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
)

var errTimetableRepoMissing = errors.New("timetable repository is required")

type TimetableService struct {
	repo  *timetablerepo.TimetableRepository
	graph *graphservice.GraphService
}

// NewTimetableService creates the timetable service with repo and optional Graph client for owner resolution.
func NewTimetableService(repo *timetablerepo.TimetableRepository, graph *graphservice.GraphService) (*TimetableService, error) {
	if repo == nil {
		return nil, errTimetableRepoMissing
	}

	return &TimetableService{
		repo:  repo,
		graph: graph,
	}, nil
}
