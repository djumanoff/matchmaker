package web

import "gopkg.in/kataras/iris.v6"

func initRoutes(app *iris.Framework) {
	app.Get("/", GetIndex)
	app.Get("/login", GetLogin)
	app.Post("/login", PostLogin)
}
