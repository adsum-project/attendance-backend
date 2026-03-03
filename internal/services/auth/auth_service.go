package auth

import (
	"context"
	"errors"
	"fmt"
	"os"

	authrepo "github.com/adsum-project/attendance-backend/internal/repositories/auth"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type AuthService struct {
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	cookieDomain string
	sessionRepo  *authrepo.SessionRepository
}

var errSessionRepoMissing = errors.New("session repository is required")

// NewAuthService builds the auth service with Entra OIDC and session storage.
func NewAuthService(sessionRepo *authrepo.SessionRepository) (*AuthService, error) {
	if sessionRepo == nil {
		return nil, errSessionRepoMissing
	}

	clientID := os.Getenv("ENTRA_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("ENTRA_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("ENTRA_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("ENTRA_CLIENT_SECRET environment variable is required")
	}

	tenantID := os.Getenv("ENTRA_TENANT_ID")
	if tenantID == "" {
		return nil, fmt.Errorf("ENTRA_TENANT_ID environment variable is required")
	}

	redirectURI := os.Getenv("ENTRA_REDIRECT_URI")
	if redirectURI == "" {
		return nil, fmt.Errorf("ENTRA_REDIRECT_URI environment variable is required")
	}

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	if cookieDomain == "" {
		return nil, fmt.Errorf("COOKIE_DOMAIN environment variable is required")
	}

	issuerURL := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID)
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	return &AuthService{
		oauth2Config: oauth2Config,
		verifier:     verifier,
		cookieDomain: cookieDomain,
		sessionRepo:  sessionRepo,
	}, nil
}

// NewAuthServiceForTesting returns an AuthService with only the session repository configured.
// Use for testing CreateSession, GetSession, DeleteSession, GetUserIDFromClaims. OIDC methods will panic if called.
func NewAuthServiceForTesting(sessionRepo *authrepo.SessionRepository) *AuthService {
	return &AuthService{sessionRepo: sessionRepo}
}
