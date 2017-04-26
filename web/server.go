package web

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type Config struct {
	StaticDir string
	Endpoint string
	DatabseUri string
}

func Run(cfg Config) {
	err := connectDb(cfg)
	panicOnError(err)

	defer closeDb()

	app := initApp()

	initRoutes(app)

	app.Listen(":8880")
}
