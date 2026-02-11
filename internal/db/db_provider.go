package db

import "github.com/jmoiron/sqlx"

type DbProvider struct {
	DB *sqlx.DB
}

func NewDbProvider(dbConn *sqlx.DB) *DbProvider {
	return &DbProvider{DB: dbConn}
}
