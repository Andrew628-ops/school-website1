package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type PageData struct {
	Title, Active, AcademicYear, CSRF, Query, School, Category, Error, Notice, Reference, Tab string
	Programs                                                                                  []Program
	Schools                                                                                   []string
	NewsItem                                                                                  []News
	Events                                                                                    []Event
	Teachers                                                                                  []Teacher
	Registrations                                                                             []Registration
	Contacts                                                                                  []Contact
	Program                                                                                   Program
	Schedule                                                                                  []ScheduleRow
	Year                                                                                      int
	Form                                                                                      map[string]string
	Application                                                                               *Registration
	Teacher                                                                                   Teacher
	Article                                                                                   News
	Event                                                                                     Event
	Editing, Admin                                                                            bool
	TotalApplications, PendingApplications, AcceptedApplications, MessageCount, FacultyCount  int
	SearchCount                                                                               int
	SearchPages                                                                               []SearchPage
}
type SearchPage struct{ Title, Description, URL string }

var sitePages = []SearchPage{
	{"About Excellence", "Our mission, values, and university community.", "/about"},
	{"Admissions", "Find your path and learn how to apply.", "/admissions"},
	{"Campus life", "Student societies, sports, wellbeing, and learning spaces.", "/campus-life"},
	{"Academic timetable", "Weekly teaching schedules by programme and year.", "/timetable"},
	{"Contact the university", "Ask a question and connect with our team.", "/contact"},
	{"Track your application", "Use your reference and email to check your admission status.", "/application-status"},
}
var templateFuncs = template.FuncMap{
	"date": func(t time.Time) string {
		if t.IsZero() {
			return "Date unavailable"
		}
		return t.Format("02 Jan 2006")
	},
	"day": func(t time.Time) string { return t.Format("02") }, "month": func(t time.Time) string { return t.Format("Jan") },
	"eventTime": func(t time.Time) string { return t.Format("Mon, 02 Jan · 15:04 MST") },
	"inputDate": func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format("2006-01-02T15:04")
	},
	"excerpt": func(s string) string {
		r := []rune(s)
		if len(r) > 145 {
			return string(r[:142]) + "…"
		}
		return s
	},
	"paragraphs": func(s string) []string { return strings.Split(s, "\n\n") },
	"initials": func(s string) string {
		parts := strings.Fields(s)
		if len(parts) == 0 {
			return "UE"
		}
		start := 0
		if parts[0] == "Dr." || parts[0] == "Prof." {
			start = 1
		}
		out := ""
		for i := start; i < len(parts) && i < start+2; i++ {
			r := []rune(parts[i])
			out += string(r[0])
		}
		return out
	},
	"statusClass": func(s string) string {
		switch s {
		case "Accepted":
			return "accepted"
		case "Under review":
			return "review"
		case "Declined":
			return "declined"
		default:
			return "received"
		}
	},
	"storedDate": func(s string) string {
		t, e := parseDate(s)
		if e != nil {
			return "Legacy record"
		}
		return t.Format("02 Jan 2006")
	},
}

func render(w http.ResponseWriter, r *http.Request, file string, data PageData, status int) {
	data.AcademicYear = AcademicYear()
	data.CSRF = csrfToken(w, r)
	data.Admin = IsAdminLoggedIn(r)
	if data.Programs == nil {
		data.Programs = Programs
	}
	if data.Schools == nil {
		data.Schools = Schools()
	}
	if data.Form == nil {
		data.Form = map[string]string{}
	}
	t, err := template.New(file).Funcs(templateFuncs).ParseFiles("html-file/partials.html", "html-file/"+file)
	if err != nil {
		serverError(w, err)
		return
	}
	var buffer bytes.Buffer
	if err = t.ExecuteTemplate(&buffer, file, data); err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(buffer.Bytes())
}
func serverError(w http.ResponseWriter, err error) {
	log.Printf("School request: %v", err)
	http.Error(w, "We couldn't complete this request. Please try again.", http.StatusInternalServerError)
}
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "404.html", PageData{Title: "Page not found"}, http.StatusNotFound)
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	news, err := GetNews()
	if err != nil {
		serverError(w, err)
		return
	}
	if len(news) > 3 {
		news = news[:3]
	}
	events, err := GetEvents()
	if err != nil {
		serverError(w, err)
		return
	}
	events = upcoming(events)
	if len(events) > 3 {
		events = events[:3]
	}
	render(w, r, "index.html", PageData{Title: "A brighter future starts here", Active: "home", NewsItem: news, Events: events}, 200)
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "about.html", PageData{Title: "About Excellence", Active: "about"}, 200)
}
func admissionsHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admissions.html", PageData{Title: "Your next chapter", Active: "admissions"}, 200)
}
func campusHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "campus.html", PageData{Title: "Life at Excellence", Active: "campus"}, 200)
}
func programsHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	school := r.URL.Query().Get("school")
	found := []Program{}
	for _, p := range Programs {
		if (school == "" || p.School == school) && (q == "" || contains(p.Name+" "+p.Summary+" "+p.School, q)) {
			found = append(found, p)
		}
	}
	render(w, r, "programs.html", PageData{Title: "Find your path", Active: "programs", Programs: found, Query: q, School: school}, 200)
}
func programHandler(w http.ResponseWriter, r *http.Request) {
	p, ok := FindProgram(r.PathValue("slug"))
	if !ok {
		notFoundHandler(w, r)
		return
	}
	render(w, r, "program.html", PageData{Title: p.Name, Active: "programs", Program: p}, 200)
}
func timeTableHandler(w http.ResponseWriter, r *http.Request) {
	p, ok := FindProgram(r.URL.Query().Get("program"))
	if !ok {
		p = Programs[0]
	}
	y, _ := strconv.Atoi(r.URL.Query().Get("year"))
	maxYear := 4
	if p.Duration == "5 years" {
		maxYear = 5
	}
	if y < 1 || y > maxYear {
		y = 1
	}
	render(w, r, "timeTable.html", PageData{Title: "Your week, at a glance", Active: "timetable", Program: p, Year: y, Schedule: ProgramSchedule(p, y)}, 200)
}
func newsHandler(w http.ResponseWriter, r *http.Request) {
	items, err := GetNews()
	if err != nil {
		serverError(w, err)
		return
	}
	category := r.URL.Query().Get("category")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	filtered := []News{}
	for _, n := range items {
		if (category == "" || n.Category == category) && (q == "" || contains(n.Headline+" "+n.Article, q)) {
			filtered = append(filtered, n)
		}
	}
	render(w, r, "news.html", PageData{Title: "What's happening at Excellence", Active: "news", NewsItem: filtered, Category: category, Query: q}, 200)
}
func articleHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	n, err := GetNewsByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		notFoundHandler(w, r)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	render(w, r, "article.html", PageData{Title: n.Headline, Active: "news", Article: n}, 200)
}
func upcoming(events []Event) []Event {
	found := []Event{}
	for _, e := range events {
		if e.Start.Add(2 * time.Hour).After(time.Now()) {
			found = append(found, e)
		}
	}
	return found
}
func eventsHandler(w http.ResponseWriter, r *http.Request) {
	events, err := GetEvents()
	if err != nil {
		serverError(w, err)
		return
	}
	render(w, r, "events.html", PageData{Title: "Make room for something new", Active: "events", Events: upcoming(events)}, 200)
}
func calendarEscape(s string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\r", "", "\n", "\\n", ",", "\\,", ";", "\\;")
	return replacer.Replace(s)
}

// Calendar content lines are folded at 75 octets, preserving UTF-8 characters.
func foldCalendar(document string) string {
	var out strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(document, "\r\n"), "\r\n") {
		limit := 75
		for len(line) > limit {
			cut := limit
			for cut > 0 && !utf8.ValidString(line[:cut]) {
				cut--
			}
			out.WriteString(line[:cut])
			out.WriteString("\r\n ")
			line = line[cut:]
			limit = 74
		}
		out.WriteString(line)
		out.WriteString("\r\n")
	}
	return out.String()
}
func calendarHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	e, err := GetEventByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		notFoundHandler(w, r)
		return
	}
	if err != nil {
		serverError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="excellence-event.ics"`)
	document := fmt.Sprintf("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//University of Excellence//Campus Events//EN\r\nBEGIN:VEVENT\r\nUID:event-%d@university-of-excellence.local\r\nDTSTAMP:%s\r\nDTSTART:%s\r\nDTEND:%s\r\nSUMMARY:%s\r\nDESCRIPTION:%s\r\nLOCATION:%s\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n", e.ID, time.Now().UTC().Format("20060102T150405Z"), e.Start.UTC().Format("20060102T150405Z"), e.Start.Add(2*time.Hour).UTC().Format("20060102T150405Z"), calendarEscape(e.Title), calendarEscape(e.Description), calendarEscape(e.Location))
	_, _ = w.Write([]byte(foldCalendar(document)))
}
func profileHandler(w http.ResponseWriter, r *http.Request) {
	teachers, err := GetTeachers()
	if err != nil {
		serverError(w, err)
		return
	}
	school := r.URL.Query().Get("school")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	found := []Teacher{}
	for _, t := range teachers {
		if (school == "" || t.Department == school) && (q == "" || contains(t.Name+" "+t.Subject+" "+t.Department, q)) {
			found = append(found, t)
		}
	}
	render(w, r, "profile.html", PageData{Title: "Meet the minds behind your future", Active: "faculty", Teachers: found, School: school, Query: q}, 200)
}
func formValues(r *http.Request) map[string]string {
	result := map[string]string{}
	for key, values := range r.PostForm {
		if key != "csrf" && len(values) > 0 {
			result[key] = strings.TrimSpace(values[0])
		}
	}
	return result
}
func validEmail(s string) bool {
	a, e := mail.ParseAddress(s)
	return e == nil && a.Address == s && len(s) <= 254
}
func shortFields(form map[string]string) bool {
	for key, v := range form {
		limit := 150
		if key == "message" || key == "article" || key == "description" {
			limit = 10000
		}
		if len([]rune(v)) > limit {
			return false
		}
	}
	return true
}
func required(f map[string]string, keys ...string) bool {
	for _, k := range keys {
		if f[k] == "" {
			return false
		}
	}
	return true
}
func registrationHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Start something extraordinary", Active: "admissions", Form: map[string]string{}}
	if p, ok := FindProgram(r.URL.Query().Get("program")); ok {
		data.Form["course"] = p.Name
	}
	if r.Method == http.MethodGet {
		ref := r.URL.Query().Get("submitted")
		if strings.HasPrefix(ref, "UE-") && len(ref) == 15 {
			data.Reference = ref
			data.Notice = "Your application has been received."
		}
		render(w, r, "registration.html", data, 200)
		return
	}
	if !validCSRF(w, r) {
		return
	}
	f := formValues(r)
	data.Form = f
	if !shortFields(f) || !required(f, "fname", "lname", "em", "phone", "course", "nationality") || !validEmail(f["em"]) {
		data.Error = "Enter your name, a valid email, phone number, programme, and nationality."
		render(w, r, "registration.html", data, 422)
		return
	}
	p, ok := FindProgram(f["course"])
	if !ok {
		data.Error = "Choose a programme from the list."
		render(w, r, "registration.html", data, 422)
		return
	}
	ref, err := SaveRegistration(Registration{FirstName: f["fname"], MiddleName: f["mname"], LastName: f["lname"], Email: f["em"], PhoneNo: f["phone"], CourseOfStudy: p.Name, Nationality: f["nationality"], StateOfOrigin: f["sorigin"], LGAOfOrigin: f["lorigin"]})
	if err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, "/registration?submitted="+url.QueryEscape(ref), http.StatusSeeOther)
}
func contactHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Let's start a conversation", Active: "contact"}
	if r.Method == http.MethodGet {
		if r.URL.Query().Get("sent") == "1" {
			data.Notice = "Thank you. Your message has been delivered to the university team."
		}
		render(w, r, "contact.html", data, 200)
		return
	}
	if !validCSRF(w, r) {
		return
	}
	f := formValues(r)
	data.Form = f
	if !shortFields(f) || !required(f, "fname", "lname", "em", "message", "subject") || !validEmail(f["em"]) {
		data.Error = "Enter your name, a valid email, a subject, and a message."
		render(w, r, "contact.html", data, 422)
		return
	}
	err := SaveContact(Contact{FirstName: f["fname"], LastName: f["lname"], Email: f["em"], PhoneNo: f["phone"], Address: f["address"], Message: f["message"], Subject: f["subject"]})
	if err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, "/contact?sent=1", http.StatusSeeOther)
}
func trackingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data := PageData{Title: "Follow your next chapter", Active: "admissions"}
	if r.Method == http.MethodPost {
		if !validCSRF(w, r) {
			return
		}
		f := formValues(r)
		data.Form = f
		if !shortFields(f) || !required(f, "reference", "em") || !validEmail(f["em"]) {
			data.Error = "Enter your application reference and the email used to apply."
			render(w, r, "tracking.html", data, 422)
			return
		}
		app, err := FindRegistration(f["reference"], f["em"])
		if errors.Is(err, sql.ErrNoRows) {
			data.Error = "We couldn't find a matching application. Check your reference and email."
			render(w, r, "tracking.html", data, 422)
			return
		}
		if err != nil {
			serverError(w, err)
			return
		}
		data.Application = &app
	}
	render(w, r, "tracking.html", data, 200)
}
func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	data := PageData{Title: "Find what you're looking for", Query: q, Programs: []Program{}, NewsItem: []News{}, Teachers: []Teacher{}, Events: []Event{}}
	if q != "" {
		for _, p := range Programs {
			if contains(p.Name+" "+p.Summary+" "+p.School, q) {
				data.Programs = append(data.Programs, p)
			}
		}
		for _, p := range sitePages {
			if contains(p.Title+" "+p.Description, q) {
				data.SearchPages = append(data.SearchPages, p)
			}
		}
		news, err := GetNews()
		if err != nil {
			serverError(w, err)
			return
		}
		for _, n := range news {
			if contains(n.Headline+" "+n.Article, q) {
				data.NewsItem = append(data.NewsItem, n)
			}
		}
		teachers, err := GetTeachers()
		if err != nil {
			serverError(w, err)
			return
		}
		for _, t := range teachers {
			if contains(t.Name+" "+t.Subject+" "+t.Department, q) {
				data.Teachers = append(data.Teachers, t)
			}
		}
		events, err := GetEvents()
		if err != nil {
			serverError(w, err)
			return
		}
		for _, e := range upcoming(events) {
			if contains(e.Title+" "+e.Description, q) {
				data.Events = append(data.Events, e)
			}
		}
	}
	data.SearchCount = len(data.Programs) + len(data.SearchPages) + len(data.NewsItem) + len(data.Teachers) + len(data.Events)
	render(w, r, "search.html", data, 200)
}
