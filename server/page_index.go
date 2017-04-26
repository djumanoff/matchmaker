package server

import (
	"net/http"
)

func Index(res http.ResponseWriter, req *http.Request) {
	player, err := users.Get(req)

	if err == nil && player != nil {
		http.Redirect(res, req, "/home", 301)
		return
	}

	renderTemplate(res, "index", nil)
}
