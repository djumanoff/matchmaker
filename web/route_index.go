package web

import (
	"gopkg.in/kataras/iris.v6"
	"github.com/djumanoff/matchmaker/lib"
	"log"
)

func GetIndex(ctx *iris.Context) {
	ctx.Render("index.html", iris.Map{"Title": "Page Title"}, iris.RenderOptions{"gzip": false})
}

func GetHome(ctx *iris.Context) {
	name := ctx.Session().GetString("name")
	email := ctx.Session().GetString("email")

	if name == "" || email == "" {
		log.Println("User not authorized")
		ctx.Redirect("/login")
		return
	}

	player := &lib.Player{Email:email, DisplayName: name}
	players, err := playerMaker.ListPlayersToRate(player)
	if err != nil {
		log.Println(err.Error())
		ctx.Redirect("/home")
		return
	}

	matches, err := matchMaker.GetUpcomingMatches()
	if err != nil {
		log.Println(err.Error())
		ctx.Redirect("/home")
		return
	}

	regIds, err := matchMaker.GetPlayerRegisteredMatchIds(email)
	if err != nil {
		log.Println(err.Error())
		ctx.Redirect("/home")
		return
	}

	ctx.Render("home.html", iris.Map{
		"Name": name,
		"Email": email,
		"Players": players,
		"Matches": matches,
		"RegIds": regIds,
	}, iris.RenderOptions{"gzip": false})
}
