package auth

import (
	"net/http"
	"os"
	"time"
)

const (
	DefaultCookieName    = "session"
	DefaultCookieMaxAge  = 60 * 60 * 24 * 7
	OAuthStateCookieName = "oauth_state"
	OAuthNonceCookieName = "oauth_nonce"
	OAuthCookieMaxAge    = 600
)

func (a *AuthService) GetCookieDomain() string {
	return a.cookieDomain
}

func (a *AuthService) SetSessionCookie(w http.ResponseWriter, sessionToken string) {
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

func (a *AuthService) GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(DefaultCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (a *AuthService) ClearSessionCookie(w http.ResponseWriter) {
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

func (a *AuthService) ClearOAuthCookies(w http.ResponseWriter) {
	clearCookie(w, OAuthNonceCookieName, a.cookieDomain, a.isSecure())
	clearCookie(w, OAuthStateCookieName, a.cookieDomain, a.isSecure())
}

func (a *AuthService) SetOAuthStateCookie(w http.ResponseWriter, state string) {
	a.setOAuthCookie(w, OAuthStateCookieName, state)
}

func (a *AuthService) SetOAuthNonceCookie(w http.ResponseWriter, nonce string) {
	a.setOAuthCookie(w, OAuthNonceCookieName, nonce)
}

func (a *AuthService) ClearOAuthStateCookie(w http.ResponseWriter) {
	clearCookie(w, OAuthStateCookieName, a.cookieDomain, a.isSecure())
}

func (a *AuthService) ClearOAuthNonceCookie(w http.ResponseWriter) {
	clearCookie(w, OAuthNonceCookieName, a.cookieDomain, a.isSecure())
}

func (a *AuthService) setOAuthCookie(w http.ResponseWriter, name, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   a.cookieDomain,
		HttpOnly: true,
		Secure:   a.isSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   OAuthCookieMaxAge,
	}
	http.SetCookie(w, cookie)
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

func (a *AuthService) isSecure() bool {
	return os.Getenv("ENVIRONMENT") == "production"
}
