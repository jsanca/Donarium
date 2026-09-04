package http

import (
	"net/http"
	"time"
)

type SessionCookieWriter interface {
	Write(w http.ResponseWriter, token string)
	Clear(w http.ResponseWriter)
}

type SessionCookieReader interface {
	Read(r *http.Request) (string, error)
}

type CookieSessionHandler struct {
	name     string
	path     string
	domain   string
	secure   bool
	sameSite http.SameSite
	ttl      time.Duration
}

func NewCookieSessionHandler(name, path string, secure bool, ttl time.Duration) *CookieSessionHandler {
	return &CookieSessionHandler{
		name:     name,
		path:     path,
		secure:   secure,
		sameSite: http.SameSiteLaxMode,
		ttl:      ttl,
	}
}

func (h *CookieSessionHandler) Write(w http.ResponseWriter, token string) {
	expires := time.Now().UTC().Add(h.ttl)
	http.SetCookie(w, &http.Cookie{
		Name:     h.name,
		Value:    token,
		Path:     h.path,
		Domain:   h.domain,
		Expires:  expires,
		MaxAge:   int(h.ttl.Seconds()),
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: h.sameSite,
	})
}

func (h *CookieSessionHandler) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.name,
		Value:    "",
		Path:     h.path,
		Domain:   h.domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: h.sameSite,
	})
}

func (h *CookieSessionHandler) Read(r *http.Request) (string, error) {
	cookie, err := r.Cookie(h.name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
