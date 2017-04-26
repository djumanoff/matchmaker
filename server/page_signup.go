package server

import (
	"net/http"
	"github.com/djumanoff/matchmaker/lib"
	"errors"
	"log"
)

type SignupPage struct {
	Email string

	DisplayName string
	Password string
	ConfirmPassword string

	Error error
}

func Signup(res http.ResponseWriter, req *http.Request) {
	player, err := users.Get(req)
	log.Println(player, err)

	if err == nil && player != nil {
		http.Redirect(res, req, "/home", 301)
		return
	}

	if req.Method == "GET" {
		renderTemplate(res, "signup", nil)
		return
	} else if req.Method == "POST" {
		req.ParseForm()

		email := req.FormValue("email")
		displayName := req.FormValue("displayName")
		password := req.FormValue("password")
		confirmPassword := req.FormValue("confirmPassword")

		if password != confirmPassword {
			log.Println(errors.New("Passwords do not match"))

			renderTemplate(res, "signup", &SignupPage{email, displayName, "", "", errors.New("Passwords do not match")})
			return
		}

		player := &lib.Player{
			Email: email,
			DisplayName: displayName,
		}

		err := playerMaker.Signup(player, password)
		if err != nil {
			renderTemplate(res, "signup", &SignupPage{email, displayName, "", "", err})
			return
		}

		users.Save(player, res)

		http.Redirect(res, req, "/home", 301)
		return
	}

	http.Redirect(res, req, "/", 301)
}
