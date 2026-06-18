package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	err := InitDB()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/timetable", timeTableHandler)
	http.HandleFunc("/news", newsHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/registration", registrationHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("GET /admin/login", adminHandler)
	http.HandleFunc("POST /admin/login", adminLoginHandler)
	http.HandleFunc("/admin", adminDashboardHandler)
	http.HandleFunc("/admin/delete-registration", deleteRegistrationHandler)
	http.HandleFunc("/admin/delete-contact", deleteContactHandler)
	http.HandleFunc("/admin/edit-teacher", editTeacherHandler)
	http.HandleFunc("/admin/add-teacher", addTeacherHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Loading Server At: http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
