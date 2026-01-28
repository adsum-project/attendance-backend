package auth

import (
	"context"
	"errors"
	"fmt"
	"os"

	authrepo "github.com/adsum-project/attendance-backend/internal/repo/auth"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Auth struct {
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
	cookieDomain string
	sessionRepo  *authrepo.SessionRepository
}

var errSessionRepoMissing = errors.New("session repository is required")

func NewAuth(sessionRepo *authrepo.SessionRepository) (*Auth, error) {
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

	return &Auth{
		oauth2Config: oauth2Config,
		verifier:     verifier,
		cookieDomain: cookieDomain,
		sessionRepo:  sessionRepo,
	}, nil
}

func (a *Auth) GetCookieDomain() string {
	return a.cookieDomain
}

func (a *Auth) AuthCodeURL(state, nonce string) string {
	return a.oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
}

func (a *Auth) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.oauth2Config.Exchange(ctx, code)
}

func (a *Auth) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	idToken, err := a.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return claims, nil
}
