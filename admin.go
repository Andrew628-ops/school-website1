package main

import (
	"fmt"
	"net/http"
)

var AdminUserName = "school_website"
var AdminPassword = "234571$"

func SetAdminCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:  "admin_session",
		Value: "logged_in",
	})

}

func IsAdminLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		fmt.Println("Error: ", cookie)
		return false
	}

	if cookie.Value != "logged_in" {
		return false
	}
	return true
}
