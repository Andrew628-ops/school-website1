package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /about", aboutHandler)
	mux.HandleFunc("GET /programs", programsHandler)
	mux.HandleFunc("GET /programs/{slug}", programHandler)
	mux.HandleFunc("GET /academics", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/programs", http.StatusSeeOther) })
	mux.HandleFunc("GET /admissions", admissionsHandler)
	mux.HandleFunc("GET /campus-life", campusHandler)
	mux.HandleFunc("GET /timetable", timeTableHandler)
	mux.HandleFunc("GET /news", newsHandler)
	mux.HandleFunc("GET /news/{id}", articleHandler)
	mux.HandleFunc("GET /events", eventsHandler)
	mux.HandleFunc("GET /events/{id}/calendar", calendarHandler)
	mux.HandleFunc("GET /search", searchHandler)
	mux.HandleFunc("GET /profile", profileHandler)
	mux.HandleFunc("GET /contact", contactHandler)
	mux.HandleFunc("POST /contact", contactHandler)
	mux.HandleFunc("GET /registration", registrationHandler)
	mux.HandleFunc("POST /registration", registrationHandler)
	mux.HandleFunc("GET /application-status", trackingHandler)
	mux.HandleFunc("POST /application-status", trackingHandler)
	mux.HandleFunc("GET /admin/login", adminHandler)
	mux.HandleFunc("POST /admin/login", adminLoginHandler)
	mux.HandleFunc("GET /admin", adminOnly(adminDashboardHandler))
	mux.HandleFunc("GET /admin/dashboard", adminOnly(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/admin", http.StatusSeeOther) }))
	mux.HandleFunc("POST /admin/logout", adminOnly(logoutHandler))
	mux.HandleFunc("POST /admin/delete-registration", adminOnly(deleteRegistrationHandler))
	mux.HandleFunc("POST /admin/delete-contact", adminOnly(deleteContactHandler))
	mux.HandleFunc("POST /admin/update-status", adminOnly(updateStatusHandler))
	mux.HandleFunc("GET /admin/export", adminOnly(exportApplicationsHandler))
	for _, path := range []string{"/admin/add-teacher", "/admin/edit-teacher"} {
		mux.HandleFunc("GET "+path, adminOnly(teacherHandler))
		mux.HandleFunc("POST "+path, adminOnly(teacherHandler))
	}
	mux.HandleFunc("POST /admin/delete-teacher", adminOnly(deleteTeacherHandler))
	for _, path := range []string{"/admin/add-news", "/admin/edit-news"} {
		mux.HandleFunc("GET "+path, adminOnly(newsEditorHandler))
		mux.HandleFunc("POST "+path, adminOnly(newsEditorHandler))
	}
	mux.HandleFunc("POST /admin/delete-news", adminOnly(deleteNewsHandler))
	for _, path := range []string{"/admin/add-event", "/admin/edit-event"} {
		mux.HandleFunc("GET "+path, adminOnly(eventEditorHandler))
		mux.HandleFunc("POST "+path, adminOnly(eventEditorHandler))
	}
	mux.HandleFunc("POST /admin/delete-event", adminOnly(deleteEventHandler))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("GET /", notFoundHandler)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https:; style-src 'self'; script-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		mux.ServeHTTP(w, r)
	})
}
func main() {
	if zone, err := time.LoadLocation("Africa/Casablanca"); err == nil {
		time.Local = zone
	}
	generated := configureAdmin()
	if err := InitDB(); err != nil {
		log.Fatal("Open school database: ", err)
	}
	defer db.Close()
	if os.Getenv("SCHOOL_SEED_DEMO") != "false" {
		if err := SeedDemoContent(); err != nil {
			log.Fatal("Prepare campus content: ", err)
		}
	}
	addr := os.Getenv("SCHOOL_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	log.Printf("University of Excellence: http://%s", addr)
	if generated {
		log.Printf("Temporary admin login: %s / %s (set SCHOOL_ADMIN_PASSWORD to keep your own password)", AdminUserName, AdminPassword)
	}
	server := &http.Server{Addr: addr, Handler: routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Fatal(server.ListenAndServe())
}
