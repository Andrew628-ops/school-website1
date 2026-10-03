package main

import (
	"crypto/subtle"
	"database/sql"
	"encoding/csv"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var loginAttempts = struct {
	sync.Mutex
	entries map[string]struct {
		Count int
		Since time.Time
	}
}{entries: make(map[string]struct {
	Count int
	Since time.Time
})}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if IsAdminLoggedIn(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	render(w, r, "admin_login.html", PageData{Title: "Staff sign in"}, 200)
}
func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !validCSRF(w, r) {
		return
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	loginAttempts.Lock()
	for key, a := range loginAttempts.entries {
		if time.Since(a.Since) > 5*time.Minute {
			delete(loginAttempts.entries, key)
		}
	}
	attempt := loginAttempts.entries[ip]
	limited := attempt.Count >= 8
	loginAttempts.Unlock()
	if limited {
		render(w, r, "admin_login.html", PageData{Title: "Staff sign in", Error: "Too many sign-in attempts. Please try again in five minutes."}, 429)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.PostForm.Get("username")), []byte(AdminUserName)) == 1 && subtle.ConstantTimeCompare([]byte(r.PostForm.Get("pwd")), []byte(AdminPassword)) == 1 {
		loginAttempts.Lock()
		delete(loginAttempts.entries, ip)
		loginAttempts.Unlock()
		SetAdminCookie(w, r)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	loginAttempts.Lock()
	a := loginAttempts.entries[ip]
	if a.Since.IsZero() {
		a.Since = time.Now()
	}
	a.Count++
	loginAttempts.entries[ip] = a
	loginAttempts.Unlock()
	render(w, r, "admin_login.html", PageData{Title: "Staff sign in", Error: "The username or password is incorrect.", Form: map[string]string{"username": r.PostForm.Get("username")}}, 401)
}
func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	apps, err := GetRegistrations()
	if err != nil {
		serverError(w, err)
		return
	}
	contacts, err := GetContacts()
	if err != nil {
		serverError(w, err)
		return
	}
	teachers, err := GetTeachers()
	if err != nil {
		serverError(w, err)
		return
	}
	news, err := GetNews()
	if err != nil {
		serverError(w, err)
		return
	}
	events, err := GetEvents()
	if err != nil {
		serverError(w, err)
		return
	}
	tab := r.URL.Query().Get("tab")
	switch tab {
	case "applications", "messages", "faculty", "news", "events":
	default:
		tab = "overview"
	}
	data := PageData{Title: "University dashboard", Active: "admin", Tab: tab, Registrations: apps, Contacts: contacts, Teachers: teachers, NewsItem: news, Events: events, TotalApplications: len(apps), MessageCount: len(contacts), FacultyCount: len(teachers)}
	for _, a := range apps {
		if a.Status == "Received" || a.Status == "Under review" {
			data.PendingApplications++
		}
		if a.Status == "Accepted" {
			data.AcceptedApplications++
		}
	}
	switch r.URL.Query().Get("done") {
	case "saved":
		data.Notice = "Your changes have been saved."
	case "deleted":
		data.Notice = "The record has been deleted."
	case "status":
		data.Notice = "The application status has been updated."
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	data.Query = q
	if q != "" {
		data.Registrations = []Registration{}
		for _, a := range apps {
			if contains(a.FirstName+" "+a.LastName+" "+a.Email+" "+a.CourseOfStudy+" "+a.Reference, q) {
				data.Registrations = append(data.Registrations, a)
			}
		}
	}
	render(w, r, "admin_dashboard.html", data, 200)
}
func adminRedirect(w http.ResponseWriter, r *http.Request, tab, done string) {
	http.Redirect(w, r, "/admin?tab="+tab+"&done="+done, http.StatusSeeOther)
}
func recordError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		notFoundHandler(w, r)
	} else {
		serverError(w, err)
	}
}
func postID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PostForm.Get("id"))
	if err != nil || id < 1 {
		return 0, sql.ErrNoRows
	}
	return id, nil
}
func deleteRecord(w http.ResponseWriter, r *http.Request, tab string, fn func(int) error) {
	id, err := postID(r)
	if err == nil {
		err = fn(id)
	}
	if err != nil {
		recordError(w, r, err)
		return
	}
	adminRedirect(w, r, tab, "deleted")
}
func deleteRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	deleteRecord(w, r, "applications", DeleteRegistration)
}
func deleteContactHandler(w http.ResponseWriter, r *http.Request) {
	deleteRecord(w, r, "messages", DeleteContact)
}
func deleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	deleteRecord(w, r, "faculty", DeleteTeacher)
}
func deleteNewsHandler(w http.ResponseWriter, r *http.Request) {
	deleteRecord(w, r, "news", DeleteNews)
}
func deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	deleteRecord(w, r, "events", DeleteEvent)
}
func updateStatusHandler(w http.ResponseWriter, r *http.Request) {
	status := r.PostForm.Get("status")
	if status != "Received" && status != "Under review" && status != "Accepted" && status != "Declined" {
		http.Error(w, "Choose a valid application status.", 422)
		return
	}
	id, err := postID(r)
	if err == nil {
		err = UpdateRegistrationStatus(id, status)
	}
	if err != nil {
		recordError(w, r, err)
		return
	}
	adminRedirect(w, r, "applications", "status")
}
func safeImage(s string) bool {
	if s == "" {
		return true
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	return (u.Scheme == "https" && u.Host != "") || (u.Scheme == "" && u.Host == "" && strings.HasPrefix(u.Path, "/static/"))
}
func teacherHandler(w http.ResponseWriter, r *http.Request) {
	editing := r.URL.Path == "/admin/edit-teacher"
	data := PageData{Title: "Add faculty member", Active: "admin", Editing: editing, Form: map[string]string{}}
	if editing {
		data.Title = "Edit faculty member"
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		if r.Method == http.MethodPost {
			id, _ = postID(r)
		}
		t, err := GetTeacherById(id)
		if err != nil {
			recordError(w, r, err)
			return
		}
		data.Teacher = t
		data.Form = map[string]string{"name": t.Name, "subject": t.Subject, "department": t.Department, "image": t.Image}
	}
	if r.Method == http.MethodGet {
		render(w, r, "teacher_form.html", data, 200)
		return
	}
	f := formValues(r)
	data.Form = f
	if !shortFields(f) || !required(f, "name", "subject", "department") || !safeImage(f["image"]) {
		data.Error = "Enter a name, subject, and department. Use a /static/ path or HTTPS address for the photo."
		render(w, r, "teacher_form.html", data, 422)
		return
	}
	t := Teacher{ID: data.Teacher.ID, Name: f["name"], Subject: f["subject"], Department: f["department"], Image: f["image"]}
	var err error
	if editing {
		err = UpdateTeacher(t)
	} else {
		err = AddProfile(t)
	}
	if err != nil {
		recordError(w, r, err)
		return
	}
	adminRedirect(w, r, "faculty", "saved")
}
func newsEditorHandler(w http.ResponseWriter, r *http.Request) {
	editing := r.URL.Path == "/admin/edit-news"
	data := PageData{Title: "Publish campus news", Active: "admin", Editing: editing, Form: map[string]string{}}
	if editing {
		data.Title = "Edit campus news"
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		if r.Method == http.MethodPost {
			id, _ = postID(r)
		}
		n, err := GetNewsByID(id)
		if err != nil {
			recordError(w, r, err)
			return
		}
		data.Article = n
		data.Form = map[string]string{"headline": n.Headline, "article": n.Article, "category": n.Category, "image": n.Image}
	}
	if r.Method == http.MethodGet {
		render(w, r, "news_form.html", data, 200)
		return
	}
	f := formValues(r)
	data.Form = f
	if !shortFields(f) || !required(f, "headline", "article", "category") || !safeImage(f["image"]) {
		data.Error = "Add a headline, category, and article. Images must use a /static/ path or HTTPS address."
		render(w, r, "news_form.html", data, 422)
		return
	}
	n := News{ID: data.Article.ID, Headline: f["headline"], Article: f["article"], Category: f["category"], Image: f["image"], Date: time.Now()}
	if err := SaveNews(n); err != nil {
		recordError(w, r, err)
		return
	}
	adminRedirect(w, r, "news", "saved")
}
func eventEditorHandler(w http.ResponseWriter, r *http.Request) {
	editing := r.URL.Path == "/admin/edit-event"
	data := PageData{Title: "Create a campus event", Active: "admin", Editing: editing, Form: map[string]string{}}
	if editing {
		data.Title = "Edit campus event"
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		if r.Method == http.MethodPost {
			id, _ = postID(r)
		}
		e, err := GetEventByID(id)
		if err != nil {
			recordError(w, r, err)
			return
		}
		data.Event = e
		data.Form = map[string]string{"title": e.Title, "description": e.Description, "location": e.Location, "category": e.Category, "starts_at": e.Start.In(time.Local).Format("2006-01-02T15:04")}
	}
	if r.Method == http.MethodGet {
		render(w, r, "event_form.html", data, 200)
		return
	}
	f := formValues(r)
	data.Form = f
	start, err := time.ParseInLocation("2006-01-02T15:04", f["starts_at"], time.Local)
	if err != nil || !shortFields(f) || !required(f, "title", "description", "location", "category") {
		data.Error = "Enter the event details and a valid date and time."
		render(w, r, "event_form.html", data, 422)
		return
	}
	e := Event{ID: data.Event.ID, Title: f["title"], Description: f["description"], Location: f["location"], Category: f["category"], Start: start}
	if err := SaveEvent(e); err != nil {
		recordError(w, r, err)
		return
	}
	adminRedirect(w, r, "events", "saved")
}
func csvSafe(s string) string {
	if len(s) > 0 && strings.ContainsRune("=+-@\t\r\n", rune(s[0])) {
		return "'" + s
	}
	return s
}
func exportApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	apps, err := GetRegistrations()
	if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="applications.csv"`)
	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"Reference", "First name", "Middle name", "Last name", "Email", "Phone", "Programme", "Nationality", "Region", "Local area", "Status", "Submitted"})
	for _, a := range apps {
		row := []string{a.Reference, a.FirstName, a.MiddleName, a.LastName, a.Email, a.PhoneNo, a.CourseOfStudy, a.Nationality, a.StateOfOrigin, a.LGAOfOrigin, a.Status, a.CreatedAt}
		for i := range row {
			row[i] = csvSafe(row[i])
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}
	writer.Flush()
}
