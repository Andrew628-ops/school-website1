package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func adminHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templ, err := template.ParseFiles("html-file/admin_login.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading server %v", err)
		}
		templ.Execute(w, nil)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		u := r.FormValue("username")
		p := r.FormValue("pwd")

		if u == AdminUserName && p == AdminPassword {
			SetAdminCookie(w)
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error")
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if IsAdminLoggedIn(r) {
		templ, err := template.ParseFiles("html-file/admin_dashboard.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading server %v", err)
			return
		}
		templ.Execute(w, map[string]interface{}{
			"Registrations": GetRegistrations(),
			"Contacts":      GetContacts(),
		})
	} else {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	}
}
