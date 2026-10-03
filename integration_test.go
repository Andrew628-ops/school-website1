package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTest(t *testing.T, seed bool) http.Handler {
	t.Helper()
	if err := initDatabase(filepath.Join(t.TempDir(), "school.db")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	AdminUserName = "test_staff"
	AdminPassword = "a-private-test-password"
	sessions.values = make(map[string]time.Time)
	loginAttempts.entries = make(map[string]struct {
		Count int
		Since time.Time
	})
	if seed {
		if err := SeedDemoContent(); err != nil {
			t.Fatal(err)
		}
	}
	return routes()
}

type testBrowser struct {
	handler http.Handler
	cookies map[string]*http.Cookie
}

func browser(h http.Handler) *testBrowser { return &testBrowser{h, map[string]*http.Cookie{}} }
func (b *testBrowser) request(method, path string, form url.Values) *httptest.ResponseRecorder {
	body := ""
	if form != nil {
		body = form.Encode()
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:12345"
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, c := range b.cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	b.handler.ServeHTTP(w, req)
	for _, c := range w.Result().Cookies() {
		if c.MaxAge < 0 {
			delete(b.cookies, c.Name)
		} else {
			b.cookies[c.Name] = c
		}
	}
	return w
}
func (b *testBrowser) post(path string, form url.Values) *httptest.ResponseRecorder {
	if form == nil {
		form = url.Values{}
	}
	if c := b.cookies["school_csrf"]; c != nil {
		form.Set("csrf", c.Value)
	}
	return b.request("POST", path, form)
}
func assertCode(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("HTTP %d, want %d: %s", w.Code, want, w.Body.String())
	}
}
func (b *testBrowser) login(t *testing.T) {
	t.Helper()
	assertCode(t, b.request("GET", "/admin/login", nil), 200)
	assertCode(t, b.post("/admin/login", url.Values{"username": {AdminUserName}, "pwd": {AdminPassword}}), 303)
}

func TestPublicPagesAndFiltering(t *testing.T) {
	b := browser(setupTest(t, true))
	for _, path := range []string{"/", "/about", "/programs", "/programs/computer-science", "/admissions", "/registration", "/contact", "/profile", "/campus-life", "/news", "/news/1", "/events", "/timetable", "/application-status", "/search", "/search?q=computing", "/admin/login"} {
		t.Run(path, func(t *testing.T) {
			w := b.request("GET", path, nil)
			assertCode(t, w, 200)
			if !strings.Contains(w.Body.String(), "<title>") {
				t.Fatal("page is not rendered")
			}
		})
	}
	w := b.request("GET", "/programs?q=computer", nil)
	assertCode(t, w, 200)
	if !strings.Contains(w.Body.String(), "1 programmes") {
		t.Fatal("programme search count incorrect")
	}
	if strings.Contains(w.Body.String(), "Business Administration") {
		t.Fatal("programme filter included an unrelated programme")
	}
	w = b.request("GET", "/programs?q=zzzzzzz", nil)
	if !strings.Contains(w.Body.String(), "No programmes found.") {
		t.Fatal("empty programme results missing")
	}
	w = b.request("GET", "/profile?school=Computing+%26+Technology", nil)
	if !strings.Contains(w.Body.String(), "2 faculty profiles") {
		t.Fatal("faculty school filter did not work")
	}
	w = b.request("GET", "/timetable?program=law&year=5", nil)
	if !strings.Contains(w.Body.String(), "Year 5 · Sample") || !strings.Contains(w.Body.String(), "Constitutional law") {
		t.Fatal("programme/year schedule missing")
	}
	for _, path := range []string{"/missing", "/programs/missing", "/news/999", "/events/999/calendar"} {
		assertCode(t, b.request("GET", path, nil), 404)
	}
}
func TestApplicationSubmissionReviewAndTracking(t *testing.T) {
	h := setupTest(t, false)
	applicant := browser(h)
	assertCode(t, applicant.request("GET", "/registration", nil), 200)
	invalid := url.Values{"fname": {"Ada"}, "lname": {"Student"}, "em": {"wrong"}, "phone": {"123"}, "course": {"Computer Science"}, "nationality": {"Nigerian"}}
	assertCode(t, applicant.post("/registration", invalid), 422)
	apps, err := GetRegistrations()
	if err != nil || len(apps) != 0 {
		t.Fatal("invalid application persisted")
	}
	valid := url.Values{"fname": {"Ada"}, "lname": {"Student"}, "em": {"ADA@example.com"}, "phone": {"1234567"}, "course": {"Computer Science"}, "nationality": {"Nigerian"}}
	w := applicant.post("/registration", valid)
	assertCode(t, w, 303)
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "/registration?submitted=UE-") {
		t.Fatal("application reference missing")
	}
	assertCode(t, applicant.request("GET", loc, nil), 200)
	apps, err = GetRegistrations()
	if err != nil || len(apps) != 1 {
		t.Fatalf("application not saved: %v", err)
	}
	app := apps[0]
	assertCode(t, applicant.request("GET", loc, nil), 200)
	apps, _ = GetRegistrations()
	if len(apps) != 1 {
		t.Fatal("refresh duplicated the application")
	}
	tracker := url.Values{"reference": {strings.ToLower(app.Reference)}, "em": {"ada@example.com"}}
	w = applicant.post("/application-status", tracker)
	assertCode(t, w, 200)
	if !strings.Contains(w.Body.String(), "Hello, Ada.") || !strings.Contains(w.Body.String(), ">Received</span>") {
		t.Fatal("tracking did not display the application")
	}
	tracker.Set("em", "someone-else@example.com")
	assertCode(t, applicant.post("/application-status", tracker), 422)
	staff := browser(h)
	staff.login(t)
	assertCode(t, staff.post("/admin/update-status", url.Values{"id": {fmt.Sprint(app.ID)}, "status": {"Accepted"}}), 303)
	tracker.Set("em", "ada@example.com")
	w = applicant.post("/application-status", tracker)
	assertCode(t, w, 200)
	if !strings.Contains(w.Body.String(), "Your application has been accepted.") {
		t.Fatal("review status did not reach the applicant")
	}
	assertCode(t, staff.post("/admin/update-status", url.Values{"id": {fmt.Sprint(app.ID)}, "status": {"Made up"}}), 422)
	w = staff.request("GET", "/admin/export", nil)
	assertCode(t, w, 200)
	if !strings.Contains(w.Body.String(), app.Reference) || !strings.Contains(w.Body.String(), "Accepted") {
		t.Fatal("application missing from export")
	}
	for _, tab := range []string{"overview", "applications", "messages", "faculty", "news", "events"} {
		assertCode(t, staff.request("GET", "/admin?tab="+tab, nil), 200)
	}
	assertCode(t, staff.post("/admin/delete-registration", url.Values{"id": {fmt.Sprint(app.ID)}}), 303)
	assertCode(t, applicant.post("/application-status", tracker), 422)
}
func TestAdminAccessAndFormProtection(t *testing.T) {
	h := setupTest(t, true)
	anonymous := browser(h)
	for _, path := range []string{"/admin", "/admin/export", "/admin/add-teacher", "/admin/edit-teacher?id=1", "/admin/add-news", "/admin/edit-news?id=1", "/admin/add-event", "/admin/edit-event?id=1"} {
		w := anonymous.request("GET", path, nil)
		assertCode(t, w, 303)
		if w.Header().Get("Location") != "/admin/login" {
			t.Fatal("private page accessible")
		}
	}
	anonymous.cookies["admin_session"] = &http.Cookie{Name: "admin_session", Value: "logged_in"}
	assertCode(t, anonymous.request("GET", "/admin", nil), 303)
	for _, path := range []string{"/admin/delete-registration", "/admin/delete-contact", "/admin/edit-teacher", "/admin/delete-teacher", "/admin/update-status", "/admin/add-news", "/admin/delete-news", "/admin/add-event", "/admin/delete-event"} {
		assertCode(t, anonymous.post(path, url.Values{"id": {"1"}}), 303)
	}
	for _, path := range []string{"/registration", "/contact", "/admin/login", "/application-status"} {
		assertCode(t, anonymous.request("POST", path, url.Values{}), 403)
	}
	staff := browser(h)
	staff.login(t)
	assertCode(t, staff.request("POST", "/admin/delete-teacher", url.Values{"id": {"1"}, "csrf": {"forged"}}), 403)
	teachers, _ := GetTeachers()
	if len(teachers) != 8 {
		t.Fatal("CSRF request changed records")
	}
	cookie := staff.cookies["admin_session"]
	if cookie.Value == "logged_in" || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
		t.Fatal("session cookie protections missing")
	}
	assertCode(t, staff.post("/admin/logout", nil), 303)
	assertCode(t, staff.request("GET", "/admin", nil), 303)
}
func TestContactsAndCampusContentManagement(t *testing.T) {
	h := setupTest(t, false)
	b := browser(h)
	b.request("GET", "/contact", nil)
	assertCode(t, b.post("/contact", url.Values{"fname": {"Alex"}, "lname": {"Student"}, "em": {"alex@example.com"}, "subject": {"Admissions"}, "message": {"Can I visit?"}}), 303)
	contacts, err := GetContacts()
	if err != nil || len(contacts) != 1 || contacts[0].Subject != "Admissions" {
		t.Fatal("message was not saved")
	}
	staff := browser(h)
	staff.login(t)
	for _, path := range []string{"/admin/add-teacher", "/admin/add-news", "/admin/add-event"} {
		assertCode(t, staff.request("GET", path, nil), 200)
	}
	assertCode(t, staff.post("/admin/add-teacher", url.Values{"name": {"Dr. Test"}, "subject": {"AI"}, "department": {"Computing & Technology"}, "image": {"javascript:alert(1)"}}), 422)
	assertCode(t, staff.post("/admin/add-teacher", url.Values{"name": {"Dr. Test"}, "subject": {"AI"}, "department": {"Computing & Technology"}}), 303)
	teachers, _ := GetTeachers()
	id := teachers[0].ID
	assertCode(t, staff.request("GET", fmt.Sprintf("/admin/edit-teacher?id=%d", id), nil), 200)
	assertCode(t, staff.post("/admin/edit-teacher", url.Values{"id": {fmt.Sprint(id)}, "name": {"Dr. Updated"}, "subject": {"Databases"}, "department": {"Computing & Technology"}, "image": {"/static/image1.jpeg"}}), 303)
	teacher, _ := GetTeacherById(id)
	if teacher.Name != "Dr. Updated" {
		t.Fatal("faculty update lost")
	}
	assertCode(t, staff.post("/admin/add-news", url.Values{"headline": {"A <script>new</script> story"}, "article": {"First paragraph.\n\nSecond paragraph."}, "category": {"Research"}}), 303)
	news, _ := GetNews()
	nid := news[0].ID
	w := b.request("GET", fmt.Sprintf("/news/%d", nid), nil)
	assertCode(t, w, 200)
	if strings.Contains(w.Body.String(), "<script>new</script>") {
		t.Fatal("unescaped news content")
	}
	assertCode(t, staff.request("GET", fmt.Sprintf("/admin/edit-news?id=%d", nid), nil), 200)
	assertCode(t, staff.post("/admin/edit-news", url.Values{"id": {fmt.Sprint(nid)}, "headline": {"Updated story"}, "article": {"Updated content."}, "category": {"Campus life"}}), 303)
	start := time.Now().AddDate(0, 0, 3).Format("2006-01-02T15:04")
	assertCode(t, staff.post("/admin/add-event", url.Values{"title": {"Open day, welcome"}, "description": {"Learn; meet\nExplore"}, "category": {"Open day"}, "location": {"Hall A"}, "starts_at": {start}}), 303)
	events, _ := GetEvents()
	eid := events[0].ID
	w = b.request("GET", fmt.Sprintf("/events/%d/calendar", eid), nil)
	assertCode(t, w, 200)
	if !strings.Contains(w.Body.String(), "BEGIN:VCALENDAR") || !strings.Contains(w.Body.String(), "SUMMARY:Open day\\, welcome") || !strings.Contains(w.Body.String(), "DESCRIPTION:Learn\\; meet\\nExplore") {
		t.Fatal("calendar escaping invalid")
	}
	assertCode(t, staff.request("GET", fmt.Sprintf("/admin/edit-event?id=%d", eid), nil), 200)
	assertCode(t, staff.post("/admin/edit-event", url.Values{"id": {fmt.Sprint(eid)}, "title": {"Updated event"}, "description": {"Event details"}, "category": {"Research"}, "location": {"Hall B"}, "starts_at": {start}}), 303)
	assertCode(t, staff.post("/admin/delete-news", url.Values{"id": {fmt.Sprint(nid)}}), 303)
	assertCode(t, staff.post("/admin/delete-event", url.Values{"id": {fmt.Sprint(eid)}}), 303)
	assertCode(t, staff.post("/admin/delete-teacher", url.Values{"id": {fmt.Sprint(id)}}), 303)
	assertCode(t, staff.post("/admin/delete-contact", url.Values{"id": {fmt.Sprint(contacts[0].ID)}}), 303)
	assertCode(t, staff.request("GET", "/admin/edit-teacher?id=999", nil), 404)
}
func TestLegacyMigrationPreservesData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = legacy.Exec(`CREATE TABLE registrations (id INTEGER PRIMARY KEY AUTOINCREMENT,firstname TEXT,middlename TEXT,lastname TEXT,email TEXT,phoneno TEXT,course TEXT,nationality TEXT,stateoforigin TEXT,lgaorigin TEXT); INSERT INTO registrations (firstname,lastname,email,course) VALUES ('Existing','Applicant','existing@example.com','Computer Science')`)
	if err != nil {
		t.Fatal(err)
	}
	legacy.Close()
	if err := initDatabase(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	apps, err := GetRegistrations()
	if err != nil || len(apps) != 1 {
		t.Fatalf("legacy record lost: %v", err)
	}
	if apps[0].FirstName != "Existing" || apps[0].Reference == "" || apps[0].Status != "Received" {
		t.Fatal("legacy fields or migration defaults incorrect")
	}
	ref := apps[0].Reference
	db.Close()
	if err := initDatabase(path); err != nil {
		t.Fatal(err)
	}
	apps, _ = GetRegistrations()
	if apps[0].Reference != ref {
		t.Fatal("migration changed existing reference")
	}
}
func TestDemoSeedingDoesNotDuplicateContent(t *testing.T) {
	setupTest(t, true)
	if err := SeedDemoContent(); err != nil {
		t.Fatal(err)
	}
	teachers, _ := GetTeachers()
	news, _ := GetNews()
	events, _ := GetEvents()
	if len(teachers) != 8 || len(news) != 3 || len(events) != 3 {
		t.Fatal("repeat startup duplicated demo content")
	}
}

func TestCalendarFolding(t *testing.T) {
	original := "SUMMARY:" + strings.Repeat("Welcome é", 30) + "\r\n"
	folded := foldCalendar(original)
	for _, line := range strings.Split(strings.TrimSuffix(folded, "\r\n"), "\r\n") {
		if len(line) > 75 {
			t.Fatalf("Calendar line has %d octets", len(line))
		}
	}
	if strings.ReplaceAll(folded, "\r\n ", "") != original {
		t.Fatal("Calendar folding changed the content")
	}
}
func TestPublicNewsCannotBePublishedWithoutStaffRoute(t *testing.T) {
	b := browser(setupTest(t, false))
	b.request("GET", "/news", nil)
	assertCode(t, b.post("/news", url.Values{"headline": {"Unapproved"}, "article": {"Unapproved content"}}), http.StatusMethodNotAllowed)
	news, err := GetNews()
	if err != nil || len(news) != 0 {
		t.Fatal("Public endpoint published news")
	}
}
