package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Teacher struct {
	ID                               int
	Name, Subject, Department, Image string
}
type Registration struct {
	ID                                                                                                      int
	FirstName, MiddleName, LastName, Email, PhoneNo, CourseOfStudy, Nationality, StateOfOrigin, LGAOfOrigin string
	Reference, Status, CreatedAt                                                                            string
}
type Contact struct {
	ID                                                                        int
	FirstName, LastName, Email, PhoneNo, Address, Message, Subject, CreatedAt string
}
type News struct {
	ID                                 int
	Headline, Article, Category, Image string
	Date                               time.Time
}
type Event struct {
	ID                                     int
	Title, Description, Location, Category string
	Start                                  time.Time
}

func randomToken() string {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("cannot generate a secure token")
	}
	return hex.EncodeToString(b[:])
}
func newReference() string { return "UE-" + strings.ToUpper(randomToken()[:12]) }

func InitDB() error {
	path := os.Getenv("SCHOOL_DB_PATH")
	if path == "" {
		path = "school.db"
	}
	return initDatabase(path)
}
func initDatabase(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	if err = db.Ping(); err != nil {
		return err
	}
	statements := []string{
		`PRAGMA busy_timeout = 5000`,
		`CREATE TABLE IF NOT EXISTS teachers (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, subject TEXT, department TEXT, image TEXT)`,
		`CREATE TABLE IF NOT EXISTS registrations (id INTEGER PRIMARY KEY AUTOINCREMENT, firstname TEXT, middlename TEXT, lastname TEXT, email TEXT, phoneno TEXT, course TEXT, nationality TEXT, stateoforigin TEXT, lgaorigin TEXT)`,
		`CREATE TABLE IF NOT EXISTS contacts (id INTEGER PRIMARY KEY AUTOINCREMENT, firstname TEXT, lastname TEXT, email TEXT, phoneno TEXT, address TEXT, message TEXT)`,
		`CREATE TABLE IF NOT EXISTS news (id INTEGER PRIMARY KEY AUTOINCREMENT, headline TEXT, article TEXT, date TEXT)`,
		`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, description TEXT NOT NULL, location TEXT NOT NULL, category TEXT NOT NULL, starts_at TEXT NOT NULL)`,
	}
	for _, stmt := range statements {
		if _, err = db.Exec(stmt); err != nil {
			return err
		}
	}
	for _, c := range []struct{ table, name, definition string }{
		{"registrations", "reference", "TEXT NOT NULL DEFAULT ''"}, {"registrations", "status", "TEXT NOT NULL DEFAULT 'Received'"}, {"registrations", "created_at", "TEXT NOT NULL DEFAULT ''"},
		{"contacts", "subject", "TEXT NOT NULL DEFAULT 'General enquiry'"}, {"contacts", "created_at", "TEXT NOT NULL DEFAULT ''"},
		{"news", "category", "TEXT NOT NULL DEFAULT 'Campus'"}, {"news", "image", "TEXT NOT NULL DEFAULT ''"},
	} {
		if err = addColumn(c.table, c.name, c.definition); err != nil {
			return err
		}
	}
	rows, err := db.Query(`SELECT id FROM registrations WHERE reference = ''`)
	if err != nil {
		return err
	}
	var ids []int
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, id := range ids {
		if _, err = db.Exec(`UPDATE registrations SET reference = ? WHERE id = ?`, newReference(), id); err != nil {
			return err
		}
	}
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS registrations_reference ON registrations(reference)`)
	return err
}
func addColumn(table, name, definition string) error {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, notnull, pk int
		var column, kind string
		var defaultVal sql.NullString
		if err = rows.Scan(&cid, &column, &kind, &notnull, &defaultVal, &pk); err != nil {
			rows.Close()
			return err
		}
		if column == name {
			found = true
		}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if found {
		return nil
	}
	_, err = db.Exec("ALTER TABLE " + table + " ADD COLUMN " + name + " " + definition)
	return err
}

func SaveRegistration(r Registration) (string, error) {
	ref := newReference()
	_, err := db.Exec(`INSERT INTO registrations (firstname,middlename,lastname,email,phoneno,course,nationality,stateoforigin,lgaorigin,reference,status,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, r.FirstName, r.MiddleName, r.LastName, strings.ToLower(r.Email), r.PhoneNo, r.CourseOfStudy, r.Nationality, r.StateOfOrigin, r.LGAOfOrigin, ref, "Received", time.Now().Format(time.RFC3339))
	return ref, err
}
func SaveContact(c Contact) error {
	_, err := db.Exec(`INSERT INTO contacts (firstname,lastname,email,phoneno,address,message,subject,created_at) VALUES (?,?,?,?,?,?,?,?)`, c.FirstName, c.LastName, c.Email, c.PhoneNo, c.Address, c.Message, c.Subject, time.Now().Format(time.RFC3339))
	return err
}
func AddProfile(t Teacher) error {
	_, err := db.Exec(`INSERT INTO teachers (name,subject,department,image) VALUES (?,?,?,?)`, t.Name, t.Subject, t.Department, t.Image)
	return err
}
func UpdateTeacher(t Teacher) error {
	res, err := db.Exec(`UPDATE teachers SET name=?,subject=?,department=?,image=? WHERE id=?`, t.Name, t.Subject, t.Department, t.Image, t.ID)
	return changed(res, err)
}
func DeleteTeacher(id int) error {
	res, err := db.Exec(`DELETE FROM teachers WHERE id=?`, id)
	return changed(res, err)
}
func SaveNews(n News) error {
	if n.ID == 0 {
		_, err := db.Exec(`INSERT INTO news (headline,article,date,category,image) VALUES (?,?,?,?,?)`, n.Headline, n.Article, n.Date.Format(time.RFC3339), n.Category, n.Image)
		return err
	}
	res, err := db.Exec(`UPDATE news SET headline=?,article=?,category=?,image=? WHERE id=?`, n.Headline, n.Article, n.Category, n.Image, n.ID)
	return changed(res, err)
}
func DeleteNews(id int) error {
	res, err := db.Exec(`DELETE FROM news WHERE id=?`, id)
	return changed(res, err)
}
func SaveEvent(e Event) error {
	if e.ID == 0 {
		_, err := db.Exec(`INSERT INTO events (title,description,location,category,starts_at) VALUES (?,?,?,?,?)`, e.Title, e.Description, e.Location, e.Category, e.Start.Format(time.RFC3339))
		return err
	}
	res, err := db.Exec(`UPDATE events SET title=?,description=?,location=?,category=?,starts_at=? WHERE id=?`, e.Title, e.Description, e.Location, e.Category, e.Start.Format(time.RFC3339), e.ID)
	return changed(res, err)
}
func DeleteEvent(id int) error {
	res, err := db.Exec(`DELETE FROM events WHERE id=?`, id)
	return changed(res, err)
}
func changed(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
func DeleteRegistration(id int) error {
	res, err := db.Exec(`DELETE FROM registrations WHERE id=?`, id)
	return changed(res, err)
}
func DeleteContact(id int) error {
	res, err := db.Exec(`DELETE FROM contacts WHERE id=?`, id)
	return changed(res, err)
}
func UpdateRegistrationStatus(id int, status string) error {
	res, err := db.Exec(`UPDATE registrations SET status=? WHERE id=?`, status, id)
	return changed(res, err)
}

const registrationColumns = `id,COALESCE(firstname,''),COALESCE(middlename,''),COALESCE(lastname,''),COALESCE(email,''),COALESCE(phoneno,''),COALESCE(course,''),COALESCE(nationality,''),COALESCE(stateoforigin,''),COALESCE(lgaorigin,''),reference,status,created_at`

func scanRegistration(row interface{ Scan(...any) error }) (Registration, error) {
	var r Registration
	err := row.Scan(&r.ID, &r.FirstName, &r.MiddleName, &r.LastName, &r.Email, &r.PhoneNo, &r.CourseOfStudy, &r.Nationality, &r.StateOfOrigin, &r.LGAOfOrigin, &r.Reference, &r.Status, &r.CreatedAt)
	return r, err
}
func GetRegistrations() ([]Registration, error) {
	rows, err := db.Query(`SELECT ` + registrationColumns + ` FROM registrations ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Registration{}
	for rows.Next() {
		r, err := scanRegistration(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
func FindRegistration(ref, email string) (Registration, error) {
	return scanRegistration(db.QueryRow(`SELECT `+registrationColumns+` FROM registrations WHERE reference=? AND lower(email)=?`, strings.ToUpper(ref), strings.ToLower(email)))
}
func GetContacts() ([]Contact, error) {
	rows, err := db.Query(`SELECT id,COALESCE(firstname,''),COALESCE(lastname,''),COALESCE(email,''),COALESCE(phoneno,''),COALESCE(address,''),COALESCE(message,''),subject,created_at FROM contacts ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []Contact{}
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.PhoneNo, &c.Address, &c.Message, &c.Subject, &c.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, rows.Err()
}
func GetTeachers() ([]Teacher, error) {
	rows, err := db.Query(`SELECT id,COALESCE(name,''),COALESCE(subject,''),COALESCE(department,''),COALESCE(image,'') FROM teachers ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := []Teacher{}
	for rows.Next() {
		var t Teacher
		if err := rows.Scan(&t.ID, &t.Name, &t.Subject, &t.Department, &t.Image); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, rows.Err()
}
func GetTeacherById(id int) (Teacher, error) {
	var t Teacher
	err := db.QueryRow(`SELECT id,name,subject,department,image FROM teachers WHERE id=?`, id).Scan(&t.ID, &t.Name, &t.Subject, &t.Department, &t.Image)
	return t, err
}
func parseDate(raw string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05.999999999 -0700 MST", "2006-01-02 15:04:05", "2006-01-02"} {
		t, err := time.Parse(layout, raw)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid stored date: %s", raw)
}
func GetNews() ([]News, error) {
	rows, err := db.Query(`SELECT id,COALESCE(headline,''),COALESCE(article,''),COALESCE(date,''),category,image FROM news ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []News{}
	for rows.Next() {
		var n News
		var raw string
		if err := rows.Scan(&n.ID, &n.Headline, &n.Article, &raw, &n.Category, &n.Image); err != nil {
			return nil, err
		}
		n.Date, _ = parseDate(raw)
		result = append(result, n)
	}
	return result, rows.Err()
}
func GetNewsByID(id int) (News, error) {
	var n News
	var raw string
	err := db.QueryRow(`SELECT id,headline,article,date,category,image FROM news WHERE id=?`, id).Scan(&n.ID, &n.Headline, &n.Article, &raw, &n.Category, &n.Image)
	if err != nil {
		return n, err
	}
	n.Date, _ = parseDate(raw)
	return n, nil
}
func GetEvents() ([]Event, error) {
	rows, err := db.Query(`SELECT id,title,description,location,category,starts_at FROM events ORDER BY starts_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Event{}
	for rows.Next() {
		var e Event
		var raw string
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Location, &e.Category, &raw); err != nil {
			return nil, err
		}
		e.Start, err = parseDate(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
func GetEventByID(id int) (Event, error) {
	var e Event
	var raw string
	err := db.QueryRow(`SELECT id,title,description,location,category,starts_at FROM events WHERE id=?`, id).Scan(&e.ID, &e.Title, &e.Description, &e.Location, &e.Category, &raw)
	if err != nil {
		return e, err
	}
	e.Start, err = parseDate(raw)
	return e, err
}

// Demo content fills empty catalogues only. Existing school records are retained.
func SeedDemoContent() error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM teachers`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		for _, t := range []Teacher{
			{Name: "Dr. Amina Bello", Subject: "Software Engineering", Department: "Computing & Technology", Image: "/static/image1.jpeg"},
			{Name: "Prof. David Okafor", Subject: "Public Health", Department: "Health & Life Sciences", Image: "/static/image2.jpeg"},
			{Name: "Dr. Sarah Williams", Subject: "Entrepreneurship", Department: "Business & Management", Image: "/static/image3.jpg"},
			{Name: "Dr. Grace Adeyemi", Subject: "Media & Communication", Department: "Arts & Humanities", Image: "/static/image4.jpg"},
			{Name: "Dr. James Morgan", Subject: "Civil Engineering", Department: "Engineering & Design", Image: "/static/image5.jpg"},
			{Name: "Dr. Daniel Mensah", Subject: "International Relations", Department: "Law & Social Sciences", Image: "/static/image6.jpg"},
			{Name: "Dr. Samuel Ibrahim", Subject: "Artificial Intelligence", Department: "Computing & Technology", Image: "/static/image.jpg"},
			{Name: "Dr. Michael Johnson", Subject: "Electrical Systems", Department: "Engineering & Design", Image: "/static/image7.jpg"},
		} {
			if err := AddProfile(t); err != nil {
				return err
			}
		}
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM news`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		for i, n := range []News{
			{Headline: "A new chapter for student innovation", Category: "Research", Article: "Our student innovation programme brings together curious minds from computing, engineering, and business.\n\nThrough guided team projects, students develop ideas that respond to challenges in their communities. The programme includes mentoring, practical workshops, and opportunities to share early prototypes.\n\nInterested students can contact their faculty office to learn how to join the next project cycle."},
			{Headline: "More than a degree: finding your community", Category: "Campus life", Article: "A university experience is made of the people you meet and the things you try. Our student societies create space for creativity, collaboration, and friendship.\n\nFrom debate and creative writing to coding and community service, there are different ways to get involved. The student services team can help you discover a group that matches your interests.\n\nVisit the campus life page to explore the opportunities available in this sample university community."},
			{Headline: "Your next chapter starts with an application", Category: "Admissions", Article: "Explore our undergraduate programmes and find the path that fits your interests. Each programme page introduces the subjects you will study and the careers they can support.\n\nApplications can be submitted through our admissions form. After submission, keep your application reference: you will need it alongside your email address to check your progress.\n\nOur admissions team reviews submissions and updates their status in the university portal."},
		} {
			n.Date = time.Now().AddDate(0, 0, -(8 - i*3))
			if err := SaveNews(n); err != nil {
				return err
			}
		}
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM events`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		for i, e := range []Event{
			{Title: "Discover Excellence: campus open day", Description: "Meet faculty, explore our learning spaces, and ask your questions about university life. A welcoming introduction for prospective students and their families.", Location: "Main campus · Welcome Centre", Category: "Open day"},
			{Title: "Ideas into action: innovation showcase", Description: "See student projects, meet their creators, and discover what happens when different disciplines work together.", Location: "Innovation Hub", Category: "Research"},
			{Title: "Find your people: societies fair", Description: "Explore student-led clubs, meet society representatives, and discover a new interest. All current students are welcome.", Location: "Student Centre courtyard", Category: "Campus life"},
		} {
			day := time.Now().AddDate(0, 0, 7+i*7)
			e.Start = time.Date(day.Year(), day.Month(), day.Day(), 10+i, 0, 0, 0, day.Location())
			if err := SaveEvent(e); err != nil {
				return err
			}
		}
	}
	return nil
}
