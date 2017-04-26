package server

import (
	"net/http"
	"github.com/djumanoff/matchmaker/lib"
	"log"
	"strings"
	"strconv"
)

type HomePage struct {
	Title string
	Player *lib.Player
	Players []lib.PlayerInfo

	Error error
}

func getEmailFromKey(key string) string {
	parts := strings.Split(key, "rating[")
	if len(parts) < 2 {
		return ""
	}

	parts = strings.Split(parts[1], "]")
	if len(parts) < 2 {
		return ""
	}

	return parts[0]
}

func Rate(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Redirect(res, req, "/home", 301)
		return
	}

	player, err := users.Get(req)
	if err != nil || player == nil {
		log.Println(err)
		users.Delete(req, res)
		http.Redirect(res, req, "/login", 301)
		return
	}

	req.ParseForm()

	for key, values := range req.PostForm {
		email := getEmailFromKey(key)
		rating, _ := strconv.Atoi(values[0])

		playerMaker.RatePlayer(player.Email, email, rating)
	}

	http.Redirect(res, req, "/home", 301)
}

func Home(res http.ResponseWriter, req *http.Request) {
	player, err := users.Get(req)

	if err != nil || player == nil {
		log.Println(err)
		users.Delete(req, res)
		http.Redirect(res, req, "/login", 301)
		return
	}

	players, err := playerMaker.ListPlayersToRate(player)
	if err != nil {
		log.Println(err)
		renderTemplate(res, "home", &HomePage{"Home Page", player, nil, err})
		return
	}

	renderTemplate(res, "home", &HomePage{"Home Page", player, players, nil})
}
