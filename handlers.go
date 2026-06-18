package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templ, err := template.ParseFiles("html-file/about.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Server Error: %v", err)
			return
		}
		templ.Execute(w, nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func timeTableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templ, err := template.ParseFiles("html-file/timeTable.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Server Error: %v", err)
			return
		}
		templ.Execute(w, nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templ, err := template.ParseFiles("html-file/index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Server Error: %v", err)
			return
		}
		templ.Execute(w, nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templ, err := template.ParseFiles("html-file/news.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading server%v", err)
			return
		}
		templ.Execute(w, map[string]interface{}{
			"NewsItem": GetNews(),
		})
	case "POST":
		templ, err := template.ParseFiles("html-file/news.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading server%v", err)
			return
		}

		b := r.FormValue("headline")
		s := r.FormValue("article")

		result := News{
			Headline: b,
			Article:  s,
			Date:     time.Now(),
		}
		AddArticle(result)
		templ.Execute(w, map[string]interface{}{
			"NewsItem": GetNews(),
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templ, err := template.ParseFiles("html-file/contact.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading.. %v", err)
			return
		}
		templ.Execute(w, nil)
	case "POST":
		templ, err := template.ParseFiles("html-file/contact.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading.. %v", err)
			return
		}

		name := r.FormValue("fname")
		lastname := r.FormValue("lname")
		email := r.FormValue("em")
		number := r.FormValue("phone")
		address := r.FormValue("address")
		message := r.FormValue("message")
		result := Contact{
			FirstName: name,
			LastName:  lastname,
			Email:     email,
			PhoneNo:   number,
			Address:   address,
			Message:   message,
		}

		SaveContact(result)
		templ.Execute(w, "Thank You, You Have Successfully Submitted your contact")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func registrationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templ, err := template.ParseFiles("html-file/registration.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading... %v", err)
			return
		}
		templ.Execute(w, nil)
	case "POST":
		templ, err := template.ParseFiles("html-file/registration.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading... %v", err)
			return
		}

		firstname := r.FormValue("fname")
		middlename := r.FormValue("mname")
		lastname := r.FormValue("lname")
		email := r.FormValue("em")
		number := r.FormValue("phone")
		course := r.FormValue("course")
		nationality := r.FormValue("nationality")
		origin := r.FormValue("sorigin")
		LGA := r.FormValue("lorigin")

		result := Registration{
			FirstName:     firstname,
			MiddleName:    middlename,
			LastName:      lastname,
			Email:         email,
			PhoneNo:       number,
			CourseOfStudy: course,
			Nationality:   nationality,
			StateOfOrigin: origin,
			LGAOfOrigin:   LGA,
		}
		SaveRegistration(result)

		templ.Execute(w, "Thanks for Registering we will get back to you shortly")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templ, err := template.ParseFiles("html-file/profile.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading server.. %v", err)
			return
		}
		templ.Execute(w, map[string]interface{}{
			"Teachers": GetTeachers(),
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}
