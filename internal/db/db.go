package db

import (
	"fmt"
	"os"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
)

func OpenFromEnv() (*sqlx.DB, error) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_DSN environment variable is required")
	}

	db, err := sqlx.Open("sqlserver", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	var pingErr error
	for attempt := 0; attempt < 5; attempt++ {
		if pingErr = db.Ping(); pingErr == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to ping database: %w", pingErr)
	}

	return db, nil
}
