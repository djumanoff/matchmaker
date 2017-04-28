package web

import (
	"gopkg.in/kataras/iris.v6"
	"log"
	"github.com/djumanoff/matchmaker/lib"
)

type SignupData struct {
	Email string
	DisplayName string

	Password string
	ConfirmPassword string
}

func GetSignup(ctx *iris.Context) {
	ctx.Render("signup.html", iris.Map{"Title": "Page Title"}, iris.RenderOptions{"gzip": false})
}

func PostSignup(ctx *iris.Context) {
	signupData := &SignupData{}
	err := ctx.ReadForm(signupData)

	if err != nil {
		log.Println("Error when reading form: " + err.Error())
		ctx.Redirect("/login")
		return
	}

	if signupData.Password != signupData.ConfirmPassword {
		log.Println("Error when reading form: " + err.Error())
		ctx.Redirect("/login")
		return
	}

	player := &lib.Player{
		Email: signupData.Email,
		Password: signupData.Password,
		DisplayName: signupData.DisplayName,
	}

	err = playerMaker.Signup(player, signupData.Password)
	if err != nil {
		log.Println("Error when reading form: " + err.Error())
		ctx.Redirect("/login")
		return
	}

	log.Println("user signed up", player)

	ctx.Session().Set("name", player.DisplayName)
	ctx.Session().Set("email", player.Email)

	ctx.Redirect("/home")
}
