package web

import (
	"gopkg.in/kataras/iris.v6"
	"log"
)

type LoginData struct {
	Email string
	Password string
}

func GetLogin(ctx *iris.Context) {
	ctx.Render("login.html", nil, iris.RenderOptions{"gzip": false})
}

func PostLogin(ctx *iris.Context) {
	loginData := &LoginData{}
	err := ctx.ReadForm(loginData)

	log.Println(loginData.Email + " " + loginData.Password)

	if err != nil {
		log.Println("Error when reading form: " + err.Error())
		ctx.Redirect("/login")
		return
	}

	player, err := playerMaker.Login(loginData.Email, loginData.Password)
	if err != nil {
		log.Println("Error when loggin in: " + err.Error())
		ctx.Redirect("/login")
		return
	}

	log.Println("user logged int", player)

	ctx.Session().Set("name", player.DisplayName)
	ctx.Session().Set("email", player.Email)

	ctx.Redirect("/home")
}

func GetLogout(ctx *iris.Context) {
	ctx.Session().Delete("name")
	ctx.Session().Delete("email")

	ctx.Redirect("/")
}
