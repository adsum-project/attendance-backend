package auth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

func (a *AuthService) AuthCodeURL(state, nonce string) string {
	return a.oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
}

func (a *AuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.oauth2Config.Exchange(ctx, code)
}

func (a *AuthService) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
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

func (a *AuthService) GetUserIDFromClaims(claims map[string]interface{}) string {
	if claims == nil {
		return ""
	}

	if oid, ok := claims["oid"].(string); ok && oid != "" {
		return oid
	}
	
	return ""
}
