package authhandlers

import (
	"fmt"
	"os"

	"github.com/adsum-project/attendance-backend/internal/services/auth"
)

type AuthProvider struct {
	auth        *auth.Auth
	frontendURL string
	logoutURL   string
}

func NewAuthProvider(a *auth.Auth) (*AuthProvider, error) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		return nil, fmt.Errorf("FRONTEND_URL environment variable is required")
	}

	tenantID := os.Getenv("ENTRA_TENANT_ID")
	if tenantID == "" {
		return nil, fmt.Errorf("ENTRA_TENANT_ID environment variable is required")
	}

	logoutURL := fmt.Sprintf(
		"https://login.microsoftonline.com/%s/oauth2/v2.0/logout?post_logout_redirect_uri=%s",
		tenantID,
		frontendURL,
	)

	return &AuthProvider{
		auth:        a,
		frontendURL: frontendURL,
		logoutURL:   logoutURL,
	}, nil
}
