package auth

import (
	"context"

	authmodels "github.com/adsum-project/attendance-backend/internal/models/auth"
)

// CreateSession persists a session for the user and returns a session token.
func (a *AuthService) CreateSession(ctx context.Context, userID string, claims map[string]interface{}) (string, error) {
	if a.sessionRepo == nil {
		return "", errSessionRepoMissing
	}
	return a.sessionRepo.CreateSession(ctx, userID, claims)
}

// GetSession returns the session for the given token, or ErrSessionNotFound.
func (a *AuthService) GetSession(ctx context.Context, token string) (*authmodels.Session, error) {
	if a.sessionRepo == nil {
		return nil, errSessionRepoMissing
	}
	return a.sessionRepo.GetSession(ctx, token)
}

// DeleteSession removes the session for the given token.
func (a *AuthService) DeleteSession(ctx context.Context, token string) error {
	if a.sessionRepo == nil {
		return errSessionRepoMissing
	}
	return a.sessionRepo.DeleteSession(ctx, token)
}
