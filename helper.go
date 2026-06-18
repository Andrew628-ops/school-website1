package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Teacher struct {
	ID         int
	Name       string
	Subject    string
	Department string
	Image      string
}

type Registration struct {
	ID            int
	FirstName     string
	MiddleName    string
	LastName      string
	Email         string
	PhoneNo       string
	CourseOfStudy string
	Nationality   string
	StateOfOrigin string
	LGAOfOrigin   string
}

type Contact struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	PhoneNo   string
	Address   string
	Message   string
}

type News struct {
	Headline string
	Article  string
	Date     time.Time
}

func AddArticle(article News) {
	db.Exec(`INSERT INTO news (headline, article, date) VALUES (?, ?, ?)`, article.Headline, article.Article, article.Date)
}

func SaveContact(s Contact) {
	db.Exec(`INSERT INTO contacts (firstname, lastname, email, phoneno, address, message) VALUES (?, ?, ?, ?, ?, ?)`, s.FirstName, s.LastName, s.Email, s.PhoneNo, s.Address, s.Message)
}

func SaveRegistration(r Registration) {
	db.Exec(`INSERT INTO registrations (firstname, middlename, lastname, email, phoneno, course, nationality, stateoforigin, lgaorigin) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, r.FirstName, r.MiddleName, r.LastName, r.Email, r.PhoneNo, r.CourseOfStudy, r.Nationality, r.StateOfOrigin, r.LGAOfOrigin)
}

func AddProfile(profile Teacher) {
	db.Exec(`INSERT INTO teachers (name, subject, department, image) VALUES (?, ?, ?, ?)`, profile.Name, profile.Subject, profile.Department, profile.Image)
}

func InitDB() error {
	var err error
	db, err = sql.Open("sqlite3", "school.db")
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS teachers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		subject TEXT,
		department TEXT,
		image TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS registrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		firstname TEXT,
		middlename TEXT,
		lastname TEXT,
		email TEXT,
		phoneno TEXT,
		course TEXT,
		nationality TEXT,
		stateoforigin TEXT,
		lgaorigin TEXT 
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		firstname TEXT,
		lastname  TEXT,
		email     TEXT,
		phoneno   TEXT,
		address   TEXT,
		message   TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS news (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		headline TEXT,
		article  TEXT,
		date     TEXT
	)`)
	return nil
}

func GetNews() []News {
	rows, err := db.Query(`SELECT headline, article, date FROM news`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []News

	for rows.Next() {
		var n News
		rows.Scan(&n.Headline, &n.Article, &n.Date)
		results = append(results, n)
	}

	return results
}

func GetRegistrations() []Registration {
	rows, err := db.Query(`SELECT id, firstname, middlename, lastname, email, phoneno, course, nationality, stateoforigin, lgaorigin FROM registrations`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []Registration

	for rows.Next() {
		var r Registration
		rows.Scan(&r.ID, &r.FirstName, &r.MiddleName, &r.LastName, &r.Email, &r.PhoneNo, &r.CourseOfStudy, &r.Nationality, &r.StateOfOrigin, &r.LGAOfOrigin)
		results = append(results, r)
	}

	return results
}

func GetContacts() []Contact {
	rows, err := db.Query(`SELECT id, firstname, lastname, email, phoneno, address, message FROM contacts`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []Contact

	for rows.Next() {
		var c Contact
		rows.Scan(&c.ID, &c.FirstName, &c.LastName, &c.Email, &c.PhoneNo, &c.Address, &c.Message)
		results = append(results, c)
	}

	return results
}

func GetTeachers() []Teacher {
	rows, err := db.Query(`SELECT id, name, subject, department, image FROM teachers`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []Teacher

	for rows.Next() {
		var t Teacher
		rows.Scan(&t.ID, &t.Name, &t.Subject, &t.Department, &t.Image)
		results = append(results, t)

	}

	return results
}

func DeleteRegistration(id int) {
	db.Exec(`DELETE FROM registrations WHERE id = ?`, id)
}

func DeleteContact(id int) {
	db.Exec(`DELETE FROM contacts WHERE id = ?`, id)
}

func GetTeacherById(id int) Teacher {
	rows := db.QueryRow(`SELECT id, name, subject, department, image FROM teachers WHERE id = ?`, id)

	var t Teacher
	rows.Scan(&t.ID, &t.Name, &t.Subject, &t.Department, &t.Image)
	return t
}

func UpdateTeacher(t Teacher) {
	db.Exec(`UPDATE teachers SET name = ?, subject = ?, department = ?, image = ? WHERE id = ?`, t.Name, t.Subject, t.Department, t.Image, t.ID)
}
