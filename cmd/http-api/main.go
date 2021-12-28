package main

import (
	"github.com/djumanoff/matchmaker/internal/matchmaker"
	"github.com/l00p8/cfg"
	"github.com/l00p8/l00p8"
	"github.com/l00p8/log"
	"github.com/l00p8/shield"
	"github.com/l00p8/xserver"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var (
	// flags of the cli
	flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Load configuration from .env file",
			Value:   "",
			EnvVars: []string{"CONFIG", "CFG"},
		},
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "run http server on specified address",
			Value:   "",
			EnvVars: []string{"ADDR"},
		},
	}

	version     = "0.1.0"
	serviceName = "matchmaker-http-api"
)

// initialize application
func main() {
	app := &cli.App{}
	app.Name = serviceName
	app.Usage = serviceName
	app.Flags = flags
	app.Action = run
	app.Version = version
	err := app.Run(os.Args)
	must(err)
}

// Config struct for server
type Config struct {
	Addr        string `envconfig:"addr" mapstructure:"addr" default:":8080"`
	RateLimit   int64  `envconfig:"rate_limit" mapstructure:"rate_limit" default:"10000"`
	LogLevel    string `envconfig:"log_level" mapstructure:"log_level" default:"debug"`
	System      string `envconfig:"system" mapstructure:"system" default:"matchmaker-http-api"`
	Hostname    string `envconfig:"hostname" mapstructure:"hostname" default:"localhost"`
	Environment string `envconfig:"env" mapstructure:"env" default:"local"`
	JaegerUrl   string `envconfig:"jaeger_url" mapstructure:"jaeger_url" default:"http://localhost:14268/api/traces"`
	PubKey      string `envconfig:"pub_key" mapstructure:"pub_key" default:""`
	PrivKey     string `envconfig:"priv_key" mapstructure:"priv_key" default:""`
	KeyAlg      string `envconfig:"key_alg" mapstructure:"key_alg" default:""`
	DBURL       string `envconfig:"db_url" mapstructure:"db_url" default:""`
}

func (c *Config) load(ctx *cli.Context) {
	// load config from file if config file provided
	configPath := ctx.String("config")
	if configPath != "" {
		_ = cfg.LoadFromFile("", c, configPath)
	} else {
		_ = cfg.Load("", c)
	}
	// load config from command line arguments
	addr := ctx.String("address")
	if addr != "" {
		c.Addr = addr
	}
}

func getLogger(c *Config) (*log.Factory, error) {
	lg, err := log.NewLogger(
		c.LogLevel,
		//zap.String("system", c.System),
		//zap.String("hostname", c.Hostname),
	)
	if err != nil {
		return nil, err
	}
	return log.NewFactory(lg), nil
}

// run func runs http server, returns error if was not able to run the server
// panics when passed invalid configuration
func run(c *cli.Context) error {
	config := &Config{}

	// init config
	config.load(c)
	logger, err := getLogger(config)
	must(err)

	// init config for http server
	hhCfg := xserver.Config{
		GracefulTimeout: 3 * time.Second,
		ShutdownTimeout: 3 * time.Second,
		Addr:            config.Addr,
		RateLimit:       config.RateLimit,
		Logger:          logger,
		Timeout:         60 * time.Second,
	}

	// init error system
	errSys := l00p8.NewErrorSystem(config.System)

	//err = tracer.InitProvider(&tracer.Config{
	//	ServiceName:    serviceName,
	//	ServiceVersion: version,
	//	Environment:    config.Environment,
	//	JaegerUrl:      config.JaegerUrl,
	//})
	//must(err)

	router := l00p8.NewHandlerRouter(
		//xserver.NewRouterWithTracing(
		xserver.NewRouter(hhCfg),
		//),
		l00p8.JSON,
	)

	mm := &matchmaker.MatchMaker{}
	err = mm.Connect(config.DBURL)
	must(err)

	pm := &matchmaker.PlayerMaker{}
	err = pm.Connect(config.DBURL)
	must(err)

	vld, err := shield.NewValidator([]byte(config.PubKey), config.KeyAlg)
	must(err)

	iss, err := shield.NewIssuer([]byte(config.PrivKey), config.KeyAlg)
	must(err)

	app := matchmaker.NewApp(mm, pm, errSys, logger, iss)

	mw := l00p8.NewMiddlewareFactory(vld, errSys)

	// init routes
	router.Post("/login", app.Login)
	router.Post("/signup", app.Signup)

	router.Get("/matches", app.GetMatches)
	router.Get("/matches/{matchId}/teams", app.GetMatchTeams)

	router.Post("/matches", mw.Auth(app.CreateMatches))
	router.Post("/matches/{matchId}/participants", mw.Auth(app.GetMatchParticipants))
	router.Post("/matches/{matchId}/register", mw.Auth(app.RegisterInMatch))
	router.Delete("/matches/{matchId}/unregister", mw.Auth(app.UnregisterInMatch))
	router.Post("/matches/{matchId}/make-teams", mw.Auth(app.MakeTeams))

	router.Healthers(mm, pm)

	// start http server with cleanup function
	// to close db connections, files, queues etc.
	return xserver.Listen(hhCfg, router, func() {
		logger.Info("cleanup finished")
	})
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
