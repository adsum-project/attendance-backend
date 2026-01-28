package db

import "github.com/jmoiron/sqlx"

// DbProvider holds the shared database connection pool.
type DbProvider struct {
	DB *sqlx.DB
}

func NewDbProvider(dbConn *sqlx.DB) *DbProvider {
	return &DbProvider{DB: dbConn}
}
