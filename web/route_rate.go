package web

import (
	"gopkg.in/kataras/iris.v6"
	"github.com/djumanoff/matchmaker/lib"
	"log"
)

type Rates struct {
	Rating map[string]int
}

func PostRate(ctx *iris.Context) {
	name := ctx.Session().GetString("name")
	email := ctx.Session().GetString("email")

	if name == "" || email == "" {
		log.Println("User not authorized")
		ctx.Redirect("/login")
		return
	}

	player := &lib.Player{Email:email, DisplayName: name}

	rates := &Rates{}
	err := ctx.ReadForm(rates)

	if err != nil {
		log.Println("Error when reading form: " + err.Error())
		ctx.Redirect("/login")
		return
	}

	for email, rating := range rates.Rating {
		playerMaker.RatePlayer(player.Email, email, rating)
	}

	ctx.Redirect("/home")
}
