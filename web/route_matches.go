package web

import (
	"gopkg.in/kataras/iris.v6"
	"log"
	"time"
	"strconv"
)

type MatchFormData struct {
	Date string
	NumberOfTeams int
	NumberOfPlayersPerTeam int
}

func GetMatchTeams(ctx *iris.Context) {
	matchId := ctx.Param("matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		log.Println("Wrong match id " + matchId)
		ctx.Redirect("/home")
		return
	}

	match, err := matchMaker.GetMatchById(mId)
	if err != nil {
		log.Println("Database error " + err.Error())
		ctx.Redirect("/home")
		return
	}

	teams, err := matchMaker.GetTeams(match)

	if err != nil {
		log.Println("Database error " + err.Error())
		ctx.Redirect("/home")
		return
	}

	ctx.Render("match_teams.html", iris.Map{
		"Teams": teams,
	}, iris.RenderOptions{"gzip": false})
}

func GetMatchesMakeTeams(ctx *iris.Context) {
	name := ctx.Session().GetString("name")
	email := ctx.Session().GetString("email")

	if name == "" || email == "" {
		log.Println("User not authorized")
		ctx.Redirect("/login")
		return
	}

	matchId := ctx.Param("matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		log.Println("Wrong match id " + matchId)
		ctx.Redirect("/home")
		return
	}

	match, err := matchMaker.GetMatchById(mId)
	if err != nil {
		log.Println("Database error " + err.Error())
		ctx.Redirect("/home")
		return
	}

	err = matchMaker.MakeTeams(match)
	if err != nil {
		log.Println("Database error " + err.Error())
		ctx.Redirect("/home")
		return
	}

	ctx.Redirect("/matches/" + matchId + "/teams")
}

func GetMatchesRegister(ctx *iris.Context) {
	name := ctx.Session().GetString("name")
	email := ctx.Session().GetString("email")

	if name == "" || email == "" {
		log.Println("User not authorized")
		ctx.Redirect("/login")
		return
	}

	matchId := ctx.Param("matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		log.Println("Wrong match id " + matchId)
		ctx.Redirect("/home")
		return
	}

	err = matchMaker.RegisterForMatch(mId, email)
	if err != nil {
		log.Println("Database error " + err.Error())
		ctx.Redirect("/home")
		return
	}

	ctx.Redirect("/matches/" + matchId + "/registered")
}

func GetMatchesRegistered(ctx *iris.Context) {
	matchId := ctx.Param("matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		log.Println("Wrong match id " + matchId)
		return
	}

	players, err := matchMaker.GetRegisteredPlayers(mId)
	if err != nil {
		log.Println("Database error " + err.Error())
		return
	}

	ctx.Render("match_registered.html", iris.Map{
		"Players": players,
	}, iris.RenderOptions{"gzip": false})
}

func PostMatches(ctx *iris.Context) {
	name := ctx.Session().GetString("name")
	email := ctx.Session().GetString("email")

	if name == "" || email == "" {
		log.Println("User not authorized")
		ctx.Redirect("/login")
		return
	}

	defer ctx.Redirect("/home")

	matchData := &MatchFormData{}
	err := ctx.ReadForm(matchData)

	if err != nil {
		log.Println("Error when reading form: " + err.Error())
		return
	}

	log.Println(matchData.Date)

	date, err := time.Parse("2006-01-02 15:04:05", matchData.Date)
	if err != nil {
		log.Println("Error when reading form: " + err.Error())
		return
	}

	_, err = matchMaker.CreateMatch(date, matchData.NumberOfTeams, matchData.NumberOfPlayersPerTeam)
	if err != nil {
		log.Println("Error creating match: " + err.Error())
		return
	}
}
