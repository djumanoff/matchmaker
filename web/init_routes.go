package web

import "gopkg.in/kataras/iris.v6"

func initRoutes(app *iris.Framework) {
	app.Get("/", GetIndex)

	app.Get("/login", GetLogin)
	app.Post("/login", PostLogin)
	app.Get("/logout", GetLogout)

	app.Get("/home", GetHome)

	app.Get("/signup", GetSignup)
	app.Post("/signup", PostSignup)

	app.Post("/rate", PostRate)

	app.Post("/matches", PostMatches)

	app.Get("/matches/:matchId/register", GetMatchesRegister)
	app.Get("/matches/:matchId/unregister", GetMatchesUnregister)
	app.Get("/matches/:matchId/registered", GetMatchesRegistered)

	app.Get("/matches/:matchId/make-teams", GetMatchesMakeTeams)
	app.Get("/matches/:matchId/teams", GetMatchTeams)
}
