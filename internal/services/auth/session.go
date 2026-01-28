package auth

import (
	"net/http"
	"os"
	"time"
)

const (
	DefaultCookieName   = "session"
	DefaultCookieMaxAge = 60 * 60 * 24 * 7
)

func (a *Auth) SetSessionCookie(w http.ResponseWriter, sessionToken string) {
	cookie := &http.Cookie{
		Name:     DefaultCookieName,
		Value:    sessionToken,
		Path:     "/",
		Domain:   a.cookieDomain,
		HttpOnly: true,
		Secure:   a.isSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   DefaultCookieMaxAge,
		Expires:  time.Now().Add(time.Duration(DefaultCookieMaxAge) * time.Second),
	}

	http.SetCookie(w, cookie)
}

func (a *Auth) GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(DefaultCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (a *Auth) ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     DefaultCookieName,
		Value:    "",
		Path:     "/",
		Domain:   a.cookieDomain,
		HttpOnly: true,
		Secure:   a.isSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}

	http.SetCookie(w, cookie)
}

func (a *Auth) ClearOAuthCookies(w http.ResponseWriter) {
	clearCookie(w, "oauth_nonce", a.cookieDomain, a.isSecure())
	clearCookie(w, "oauth_state", a.cookieDomain, a.isSecure())
}

func clearCookie(w http.ResponseWriter, name, domain string, secure bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}

func (a *Auth) GetUserIDFromClaims(claims map[string]interface{}) string {
	if claims == nil {
		return ""
	}
	// Try to get 'oid' (object ID) first, fallback to 'sub'
	if oid, ok := claims["oid"].(string); ok && oid != "" {
		return oid
	}
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		return sub
	}
	return ""
}

func (a *Auth) isSecure() bool {
	if os.Getenv("ENVIRONMENT") == "production" {
		return true
	}
	return false
}
