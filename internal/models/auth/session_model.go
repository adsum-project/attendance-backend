package authmodels

import "time"

type Session struct {
	ID        string
	UserID    string
	Claims    map[string]interface{}
	ExpiresAt time.Time
}