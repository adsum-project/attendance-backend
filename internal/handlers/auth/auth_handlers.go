package authhandlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/adsum-project/attendance-backend/pkg/utils/response"
)

func (p *AuthProvider) Login(w http.ResponseWriter, r *http.Request) {
	state, err := generateRandomString(32)
	if err != nil {
		response.InternalServerError(w, "Failed to generate state: "+err.Error())
		return
	}

	nonce, err := generateRandomString(32)
	if err != nil {
		response.InternalServerError(w, "Failed to generate nonce: "+err.Error())
		return
	}

	setStateCookie(w, state, p.auth.GetCookieDomain())
	setNonceCookie(w, nonce, p.auth.GetCookieDomain())

	authURL := p.auth.AuthCodeURL(state, nonce)

	response.OK(w, "", map[string]interface{}{
		"url": authURL,
	})
}

func (p *AuthProvider) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		response.BadRequest(w, "Missing authorization code", nil)
		return
	}
	state := r.URL.Query().Get("state")
	if state == "" {
		response.BadRequest(w, "Missing state", nil)
		return
	}

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		response.Unauthorized(w, "Invalid state")
		return
	}

	ctx := context.Background()

	token, err := p.auth.ExchangeCode(ctx, code)
	if err != nil {
		response.Unauthorized(w, "Failed to authenticate: "+err.Error())
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		response.Unauthorized(w, "Missing id_token")
		return
	}

	claims, err := p.auth.ValidateToken(ctx, rawIDToken)
	if err != nil {
		response.Unauthorized(w, "Invalid token: "+err.Error())
		return
	}

	nonceCookie, err := r.Cookie("oauth_nonce")
	if err != nil {
		response.Unauthorized(w, "Missing nonce cookie")
		return
	}
	if nonceClaim, ok := claims["nonce"].(string); !ok || nonceClaim != nonceCookie.Value {
		response.Unauthorized(w, "Invalid nonce")
		return
	}

	userID := p.auth.GetUserIDFromClaims(claims)
	if userID == "" {
		response.Unauthorized(w, "Invalid token: missing user ID")
		return
	}

	sessionID, err := p.auth.CreateSession(ctx, userID, claims)
	if err != nil {
		response.InternalServerError(w, "Failed to create session")
		return
	}

	p.auth.SetSessionCookie(w, sessionID)

	clearNonceCookie(w, p.auth.GetCookieDomain())
	clearStateCookie(w, p.auth.GetCookieDomain())

	http.Redirect(w, r, p.frontendURL, http.StatusFound)
}

func (p *AuthProvider) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(map[string]interface{})
	if !ok || claims == nil {
		response.InternalServerError(w, "Failed to retrieve token claims")
		return
	}

	user := make(map[string]interface{})
	if name, ok := claims["name"].(string); ok {
		user["name"] = name
	}
	if email, ok := claims["email"].(string); ok {
		user["email"] = email
	}
	if oid, ok := claims["oid"].(string); ok {
		user["id"] = oid
	}

	response.OK(w, "", map[string]interface{}{
		"user": user,
	})
}

func (p *AuthProvider) Logout(w http.ResponseWriter, r *http.Request) {
	if sessionID, err := p.auth.GetSessionCookie(r); err == nil {
		_ = p.auth.DeleteSession(r.Context(), sessionID)
	}
	p.auth.ClearSessionCookie(w)
	p.auth.ClearOAuthCookies(w)

	response.NoContent(w)
}

// Helper functions
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

func setNonceCookie(w http.ResponseWriter, nonce, domain string) {
	cookie := &http.Cookie{
		Name:     "oauth_nonce",
		Value:    nonce,
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	}
	http.SetCookie(w, cookie)
}

func setStateCookie(w http.ResponseWriter, state, domain string) {
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	}
	http.SetCookie(w, cookie)
}

func clearNonceCookie(w http.ResponseWriter, domain string) {
	cookie := &http.Cookie{
		Name:     "oauth_nonce",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}

func clearStateCookie(w http.ResponseWriter, domain string) {
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
}
