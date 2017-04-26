package web

import "gopkg.in/kataras/iris.v6"

func GetIndex(ctx *iris.Context) {
	ctx.Render("index.html", iris.Map{"Title": "Page Title"}, iris.RenderOptions{"gzip": false})
}
