package web

import "gopkg.in/kataras/iris.v6"

type LoginData struct {
	Email string
	Password string
}

func GetLogin(ctx *iris.Context) {
	ctx.Render("login.html", nil, iris.RenderOptions{"gzip": false})
}

func PostLogin(ctx *iris.Context) {
	loginData := LoginData{}
	err := ctx.ReadForm(&loginData)
	if err != nil {

		ctx.Log(iris.DevMode, "Error when reading form: " + err.Error())
		ctx.RedirectTo("/login", iris.Map{"err": err.Error()})
		return
	}

	player, err := playerMaker.Login(loginData.Email, loginData.Password)
	if err != nil {

	}
}
