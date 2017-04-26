package server

import (
	"net/http"
	"html/template"
	"log"
	"path/filepath"
	"github.com/djumanoff/matchmaker/lib"
)

//var templates = template.Must(template.ParseFiles(
//	"templates/blocks/footer.html",
//	"templates/blocks/header.html",
//	"templates/pages/index.html",
//	"templates/pages/login.html",
//))

var matchMaker = &lib.MatchMaker{}
var playerMaker = &lib.PlayerMaker{}

func renderTemplate(w http.ResponseWriter, name string, p interface{}) {
	lp := filepath.Join("templates", "layouts", "desktop.html")
	fp := filepath.Join("templates", "pages", name + ".html")

	tmpl, err := template.ParseFiles(lp, fp);
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := tmpl.ExecuteTemplate(w, "layout", p); err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type Config struct {
	StaticDir string
	Endpoint string
	DatabseUri string
}

func Run(cfg Config) error {
	err := matchMaker.Connect(cfg.DatabseUri)
	if err != nil {
		return err
	}
	defer matchMaker.Close()

	err = playerMaker.Connect(cfg.DatabseUri)
	if err != nil {
		return err
	}
	defer playerMaker.Close()

	initRoutes()
	initStaticServer(cfg.StaticDir)

	return http.ListenAndServe(cfg.Endpoint, nil)
}

func initStaticServer(staticDir string) {
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

func initRoutes() {
	http.HandleFunc("/", Index)
	http.HandleFunc("/login/", Login)
	http.HandleFunc("/logout/", Logout)
	http.HandleFunc("/signup/", Signup)

	http.HandleFunc("/home/", Home)
	http.HandleFunc("/rate/", Rate)
}
