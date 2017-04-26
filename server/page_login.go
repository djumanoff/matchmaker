package server

import (
	"net/http"
	"log"
)

type LoginPage struct {
	Email string
	Password string

	Error error
}

func Logout(res http.ResponseWriter, req *http.Request) {
	users.Delete(req, res)
	http.Redirect(res, req, "/", 301)
}

func Login(res http.ResponseWriter, req *http.Request) {
	player, err := users.Get(req)
	log.Println(player, err)

	if err == nil && player != nil {
		http.Redirect(res, req, "/home", 301)
		return
	}


	if req.Method == "GET" {
		renderTemplate(res, "login", &LoginPage{})
		return
	} else if req.Method == "POST" {
		req.ParseForm()

		email := req.FormValue("email")
		password := req.FormValue("password")

		player, err := playerMaker.Login(email, password)
		if err != nil {
			log.Println(err)
			renderTemplate(res, "login", &LoginPage{email, "", err})
			return
		}

		users.Save(player, res)

		log.Println(users)

		http.Redirect(res, req, "/home", 301)
		return
	}

	http.Redirect(res, req, "/", 301)
}
