package timetablerepo

import "github.com/jmoiron/sqlx"

type TimetableRepository struct {
	db *sqlx.DB
}

// NewTimetableRepository creates the timetable data access layer.
func NewTimetableRepository(db *sqlx.DB) *TimetableRepository {
	return &TimetableRepository{
		db: db,
	}
}
