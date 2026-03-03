package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	graphmodels "github.com/adsum-project/attendance-backend/internal/models/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGraphServiceWithServer(t *testing.T, handler http.HandlerFunc) (*GraphService, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	svc := NewGraphServiceWithHTTPClient(srv.Client(), srv.URL+"/v1.0")
	return svc, srv
}

func TestGetUser(t *testing.T) {
	t.Run("returns user when found", func(t *testing.T) {
		want := graphmodels.GraphUser{
			ID:          "user-1",
			DisplayName: "Test User",
			Mail:        "test@example.com",
		}
		svc, srv := newGraphServiceWithServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.True(t, strings.HasSuffix(r.URL.Path, "/users/user-1"), "path: %s", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(want)
		})
		defer srv.Close()

		user, err := svc.GetUser(context.Background(), "user-1")
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, "user-1", user.ID)
		assert.Equal(t, "Test User", user.DisplayName)
		assert.Equal(t, "test@example.com", user.Mail)
	})

	t.Run("returns nil when not found", func(t *testing.T) {
		svc, srv := newGraphServiceWithServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		defer srv.Close()

		user, err := svc.GetUser(context.Background(), "missing")
		require.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		svc, srv := newGraphServiceWithServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer srv.Close()

		user, err := svc.GetUser(context.Background(), "user-1")
		require.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "500")
	})
}

func TestGetUsersByIDs(t *testing.T) {
	t.Run("returns empty when ids empty", func(t *testing.T) {
		svc := NewGraphServiceWithHTTPClient(http.DefaultClient, "")
		users, err := svc.GetUsersByIDs(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("returns users from getByIds response", func(t *testing.T) {
		want := []graphmodels.GraphUser{
			{ID: "u1", DisplayName: "User One"},
			{ID: "u2", DisplayName: "User Two"},
		}
		svc, srv := newGraphServiceWithServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.True(t, strings.HasSuffix(r.URL.Path, "/directoryObjects/getByIds"), "path: %s", r.URL.Path)
			assert.Equal(t, http.MethodPost, r.Method)
			var req struct {
				IDs []string `json:"ids"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.ElementsMatch(t, []string{"u1", "u2"}, req.IDs)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"value": want})
		})
		defer srv.Close()

		users, err := svc.GetUsersByIDs(context.Background(), []string{"u1", "u2"})
		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, "User One", users[0].DisplayName)
		assert.Equal(t, "User Two", users[1].DisplayName)
	})
}

func TestGetUserRoles(t *testing.T) {
	t.Run("returns roles from app role assignments", func(t *testing.T) {
		os.Setenv("ENTRA_CLIENT_ID", "test-app-id")
		defer os.Unsetenv("ENTRA_CLIENT_ID")

		svc, srv := newGraphServiceWithServer(t, func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.Contains(path, "servicePrincipals"):
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"id": "sp-123",
					"appRoles": []map[string]interface{}{
						{"id": "role-admin-id", "value": "Admin"},
						{"id": "role-user-id", "value": "User"},
					},
				})
			case strings.Contains(path, "appRoleAssignments"):
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"value": []map[string]interface{}{
						{"appRoleId": "role-admin-id"},
						{"appRoleId": "role-user-id"},
					},
				})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})
		defer srv.Close()

		roles, err := svc.GetUserRoles(context.Background(), "user-1")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"Admin", "User"}, roles)
	})
}
