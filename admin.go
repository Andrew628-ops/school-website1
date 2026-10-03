package main

import (
	"crypto/subtle"
	"net/http"
	"os"
	"sync"
	"time"
)

var AdminUserName, AdminPassword string
var sessions = struct {
	sync.Mutex
	values map[string]time.Time
}{values: make(map[string]time.Time)}

func configureAdmin() bool {
	AdminUserName = os.Getenv("SCHOOL_ADMIN_USERNAME")
	if AdminUserName == "" {
		AdminUserName = "school_website"
	}
	AdminPassword = os.Getenv("SCHOOL_ADMIN_PASSWORD")
	if AdminPassword == "" {
		AdminPassword = randomToken()[:20]
		return true
	}
	return false
}
func SetAdminCookie(w http.ResponseWriter, r *http.Request) {
	token := randomToken()
	expires := time.Now().Add(8 * time.Hour)
	sessions.Lock()
	for key, end := range sessions.values {
		if time.Now().After(end) {
			delete(sessions.values, key)
		}
	}
	sessions.values[token] = expires
	sessions.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "admin_session", Value: token, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60, Expires: expires})
}
func IsAdminLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return false
	}
	sessions.Lock()
	defer sessions.Unlock()
	expiry, ok := sessions.values[cookie.Value]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(sessions.values, cookie.Value)
		return false
	}
	return true
}
func adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAdminLoggedIn(r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodPost && !validCSRF(w, r) {
			return
		}
		next(w, r)
	}
}
func csrfToken(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("school_csrf")
	if err == nil && len(c.Value) == 48 {
		return c.Value
	}
	token := randomToken()
	http.SetCookie(w, &http.Cookie{Name: "school_csrf", Value: token, Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: 24 * 60 * 60})
	return token
}
func validCSRF(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "This form is too large or could not be read.", http.StatusBadRequest)
		return false
	}
	c, err := r.Cookie("school_csrf")
	if err != nil || len(c.Value) != 48 || subtle.ConstantTimeCompare([]byte(c.Value), []byte(r.PostForm.Get("csrf"))) != 1 {
		http.Error(w, "Your form session expired. Reload the page and try again.", http.StatusForbidden)
		return false
	}
	return true
}
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("admin_session"); err == nil {
		sessions.Lock()
		delete(sessions.values, c.Value)
		sessions.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: "admin_session", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteStrictMode})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
