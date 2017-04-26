package web

import (
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/view"
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/middleware/logger"
	"gopkg.in/kataras/iris.v6/adaptors/sessions"
	"github.com/gorilla/securecookie"
)

func initApp() *iris.Framework {
	app := iris.New()

	// logger init
	customLogger := logger.New(logger.Config{
		// Status displays status code
		Status: true,
		// IP displays request's remote address
		IP: true,
		// Method displays the http method
		Method: true,
		// Path displays the request path
		Path: true,
	})
	app.Use(customLogger)

	// init sessions
	cookieName := "mycustomsessionid"
	// AES only supports key sizes of 16, 24 or 32 bytes.
	// You either need to provide exactly that amount or you derive the key from what you type in.
	hashKey := []byte("the-big-and-secret-fash-key-here")
	blockKey := []byte("lot-secret-of-characters-big-too")
	secureCookie := securecookie.New(hashKey, blockKey)
	mySessions := sessions.New(sessions.Config{
		Cookie: cookieName,
		Encode: secureCookie.Encode,
		Decode: secureCookie.Decode,
	})
	app.Adapt(mySessions)

	// init dev logger
	app.Adapt(iris.DevLogger())

	// init router
	app.Adapt(httprouter.New())

	// init view engine
	tmpl := view.HTML("./templates", ".html")
	tmpl.Layout("desktop.html")
	app.Adapt(tmpl.Reload(true))

	// init static server
	app.StaticWeb("/static", "./public")

	// init not found handler
	errorLogger := logger.New()
	app.OnError(iris.StatusNotFound, func(ctx *iris.Context) {
		errorLogger.Serve(ctx)
		ctx.Writef("404")
	})

	return app
}
