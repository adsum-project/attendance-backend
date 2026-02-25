package authhandlers

import (
	"fmt"
	"os"

	"github.com/adsum-project/attendance-backend/internal/services/auth"
)

type AuthProvider struct {
	auth        *auth.AuthService
	frontendURL string
}

func NewAuthProvider(a *auth.AuthService) (*AuthProvider, error) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		return nil, fmt.Errorf("FRONTEND_URL environment variable is required")
	}

	return &AuthProvider{
		auth:        a,
		frontendURL: frontendURL,
	}, nil
}
