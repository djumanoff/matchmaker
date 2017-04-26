package main

import (
	"github.com/djumanoff/matchmaker/web"
)

func main() {
	web.Run(web.Config{
		StaticDir: "public",
		Endpoint: ":8800",
		DatabseUri: "root:root@tcp(33.33.33.1:3306)/team_builder?charset=utf8",
	})
}
