package timetablerepo

import "github.com/jmoiron/sqlx"

type TimetableRepository struct {
	db *sqlx.DB
}

func NewTimetableRepository(db *sqlx.DB) *TimetableRepository {
	return &TimetableRepository{
		db: db,
	}
}
