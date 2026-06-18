package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
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
			"Teachers":      GetTeachers(),
		})
	} else {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	}
}

func deleteRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		b, _ := strconv.Atoi(r.FormValue("id"))
		DeleteRegistration(b)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func deleteContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		b, _ := strconv.Atoi(r.FormValue("id"))
		DeleteContact(b)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func editTeacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templ, err := template.ParseFiles("html-file/admin_edit_teacher.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error: loading server %v", err)
			return
		}
		Y, _ := strconv.Atoi(r.URL.Query().Get("id"))
		Result := GetTeacherById(Y)
		templ.Execute(w, Result)
	case "POST":
		Name := r.FormValue("name")
		Subject := r.FormValue("subject")
		Dept := r.FormValue("dept")
		img := r.FormValue("image")
		id, _ := strconv.Atoi(r.FormValue("id"))

		Result := Teacher{
			Name:       Name,
			Subject:    Subject,
			Department: Dept,
			Image:      img,
			ID:         id,
		}
		UpdateTeacher(Result)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Error: method not allowed")
		return
	}
}

func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	if IsAdminLoggedIn(r) {
		switch r.Method {
		case "GET":
			templ, err := template.ParseFiles("html-file/admin_add_teacher.html")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Error: loading server %v", err)
				return
			}
			templ.Execute(w, nil)
		case "POST":
			name := r.FormValue("name")
			subject := r.FormValue("sub")
			dept := r.FormValue("department")
			image := r.FormValue("imageurl")

			result := Teacher{
				Name:       name,
				Subject:    subject,
				Department: dept,
				Image:      image,
			}
			AddProfile(result)
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Error: method not allowed")
			return
		}
	} else {
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	}
}
