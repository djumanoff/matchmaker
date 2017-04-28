package main

import (
	"github.com/djumanoff/matchmaker/web"
	"github.com/urfave/cli"
	"os"
	"fmt"
	"errors"
	"github.com/joho/godotenv"
)

var (
	DB_URL string = ""
	ENDPOINT string = ""
	CONFIG_PATH string = ""
	VERSION string = "1.0.0"
)

var flags []cli.Flag = []cli.Flag{
	cli.StringFlag{
		Name:        "config, c",
		Value:       "",
		Usage:       "path to .env config file",
		Destination: &CONFIG_PATH,
	},
	cli.StringFlag{
		Name:        "db_url, dbu",
		Value:       os.Getenv("DB_URL"),
		Usage:       "database uri",
		Destination: &DB_URL,
	},
	cli.StringFlag{
		Name:        "endpoint, e",
		Value:       os.Getenv("ENDPOINT"),
		Usage:       "mongodb host",
		Destination: &ENDPOINT,
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "Match Maker"
	app.Usage = "match maker service"
	app.UsageText = "matchmaker [global options]"
	app.Version = VERSION
	app.Flags = flags
	app.Action = runWeb

	fmt.Println(app.Run(os.Args))
}

func runWeb(*cli.Context) error {
	parseEnvFile()

	if ENDPOINT == "" {
		return errors.New("Endpoint to mount not provided")
	}

	if DB_URL == "" {
		return errors.New("No database url")
	}

	web.Run(web.Config{
		StaticDir: "public",
		Endpoint: ENDPOINT,
		DatabseUri: DB_URL,
	})

	return nil
}

func parseEnvFile() {
	if CONFIG_PATH == "" {
		return
	}
	_ = godotenv.Load(CONFIG_PATH)

	ENDPOINT = os.Getenv("ENDPOINT")
	DB_URL = os.Getenv("DB_URL")
}
