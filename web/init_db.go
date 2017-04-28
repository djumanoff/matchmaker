package web

import "github.com/djumanoff/matchmaker/lib"

var matchMaker = &lib.MatchMaker{}
var playerMaker = &lib.PlayerMaker{}

func connectDb(cfg Config) error {
	err := matchMaker.Connect(cfg.DatabseUri)
	if err != nil {
		return err
	}

	err = playerMaker.Connect(cfg.DatabseUri)
	if err != nil {
		return err
	}

	return nil
}

func closeDb() {
	matchMaker.Close()
	playerMaker.Close()
}
