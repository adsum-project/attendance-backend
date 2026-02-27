package middleware

import (
	"context"
	"net/http"

	"github.com/adsum-project/attendance-backend/internal/services/auth"
	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils/authorization"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func RequireAuthWithRedirect(a *auth.AuthService, redirectURL string, roles ...string) router.Middleware {
	return func(handler router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request) {
			fail := func(status int, msg string) {
				if redirectURL != "" {
					http.Redirect(w, r, redirectURL, http.StatusFound)
					return
				}
				if status == http.StatusForbidden {
					response.Forbidden(w, msg)
					return
				}
				response.Unauthorized(w, msg)
			}

			sessionID, err := a.GetSessionCookie(r)
			if err != nil {
				fail(http.StatusUnauthorized, "Not authenticated")
				return
			}

			session, err := a.GetSession(r.Context(), sessionID)
			if err != nil {
				a.ClearSessionCookie(w)
				a.ClearOAuthCookies(w)
				fail(http.StatusUnauthorized, "Invalid session")
				return
			}

			if session.UserID == "" {
				a.ClearSessionCookie(w)
				a.ClearOAuthCookies(w)
				fail(http.StatusUnauthorized, "Invalid session: missing user ID")
				return
			}

			ctx := context.WithValue(r.Context(), "sessionID", session.ID)
			ctx = context.WithValue(ctx, "claims", session.Claims)
			ctx = context.WithValue(ctx, "userID", session.UserID)

			if len(roles) > 0 {
				if hasRoles := authorization.HasRoles(ctx, roles...); !hasRoles {
					fail(http.StatusForbidden, "Forbidden")
					return
				}
			}

			handler(w, r.WithContext(ctx))
		}
	}
}

func RequireAuth(a *auth.AuthService, roles ...string) router.Middleware {
	return RequireAuthWithRedirect(a, "", roles...)
}

func RequireNoAuth(a *auth.AuthService) router.Middleware {
	return func(handler router.Handler) router.Handler {
		return func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := a.GetSessionCookie(r)
			if err != nil {
				handler(w, r)
				return
			}

			session, err := a.GetSession(r.Context(), sessionID)
			if err != nil {
				handler(w, r)
				return
			}

			if session.UserID == "" {
				handler(w, r)
				return
			}

			response.BadRequest(w, "Already authenticated", nil)
		}
	}
}
