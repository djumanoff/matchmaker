package matchmaker

import (
	"github.com/l00p8/l00p8"
	"github.com/l00p8/log"
	"github.com/l00p8/shield"
	"net/http"
	"strconv"
	"time"
)

func NewApp(mm *MatchMaker, pm *PlayerMaker, errSys l00p8.ErrorSystem, log log.Logger, iss shield.Issuer) *App {
	return &App{mm, pm, errSys, log, iss}
}

type App struct {
	mm     *MatchMaker
	pm     *PlayerMaker
	errSys l00p8.ErrorSystem
	log    log.Logger
	iss    shield.Issuer
}

func (api *App) Signup(r *http.Request) l00p8.Response {
	type SignupData struct {
		Email           string `json:"email"`
		DisplayName     string `json:"display_name"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	data := &SignupData{}
	err := l00p8.ParseBodyFromRequest(r, data)
	if err != nil {
		return api.errSys.BadRequest(400, err.Error())
	}
	if data.Password != data.ConfirmPassword {
		return api.errSys.BadRequest(410, "password and confirm password don't match.")
	}

	player := &Player{
		Email:       data.Email,
		Password:    data.Password,
		DisplayName: data.DisplayName,
	}

	err = api.pm.Signup(player)
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	player.Password = ""
	tkn, err := api.iss.Issue(&shield.JWTClaims{Email: player.Email, Name: player.DisplayName})
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	type response struct {
		Token string `json:"token"`
	}
	resp := &response{Token: tkn}

	return l00p8.Created(resp)
}

func (api *App) Login(r *http.Request) l00p8.Response {
	type LoginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	loginData := &LoginData{}
	err := l00p8.ParseBodyFromRequest(r, loginData)
	if err != nil {
		return api.errSys.BadRequest(400, err.Error())
	}
	player, err := api.pm.Login(loginData.Email, loginData.Password)
	if err == ErrNotFound {
		return api.errSys.NotFound(414, err.Error())
	} else if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}

	tkn, err := api.iss.Issue(&shield.JWTClaims{Email: player.Email, Name: player.DisplayName})
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	type response struct {
		Token string `json:"token"`
	}
	resp := &response{Token: tkn}

	return l00p8.OK(resp)
}

func (api *App) Rate(r *http.Request) l00p8.Response {
	type Rates struct {
		Rating map[string]int `json:"rates"`
	}
	rates := &Rates{}
	err := l00p8.ParseBodyFromRequest(r, rates)
	if err != nil {
		return api.errSys.BadRequest(400, err.Error())
	}
	claims := r.Context().Value("claims").(*shield.JWTClaims)
	for email, rating := range rates.Rating {
		err = api.pm.RatePlayer(claims.Email, email, rating)
		if err != nil {
			api.log.Warn(err.Error())
		}
	}
	return l00p8.OK("{}")
}

func (api *App) GetMatchTeams(r *http.Request) l00p8.Response {
	matchId := l00p8.URLParam(r, "matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	match, err := api.mm.GetMatchById(mId)
	if err == ErrNotFound {
		return api.errSys.NotFound(414, err.Error())
	} else if err != nil {
		return api.errSys.BadRequest(420, err.Error())
	}
	teams, err := api.mm.GetTeams(match)
	if err != nil {
		return api.errSys.InternalServerError(520, err.Error())
	}
	return l00p8.OK(teams)
}

func (api *App) GetMatches(r *http.Request) l00p8.Response {
	matches, err := api.mm.GetUpcomingMatches()
	if err != nil {
		return api.errSys.BadRequest(400, err.Error())
	}
	return l00p8.OK(matches)
}

func (api *App) CreateMatches(r *http.Request) l00p8.Response {
	type MatchFormData struct {
		Date                   string `json:"date"`
		NumberOfTeams          int    `json:"number_of_teams"`
		NumberOfPlayersPerTeam int    `json:"number_of_players_per_team"`
	}
	matchData := &MatchFormData{}
	err := l00p8.ParseBodyFromRequest(r, matchData)
	if err != nil {
		return api.errSys.BadRequest(400, err.Error())
	}
	date, err := time.Parse("2006-01-02 15:04:05", matchData.Date)
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	match, err := api.mm.CreateMatch(date, matchData.NumberOfTeams, matchData.NumberOfPlayersPerTeam)
	if err != nil {
		return api.errSys.InternalServerError(500, err.Error())
	}
	return l00p8.Created(match)
}

func (api *App) MakeTeams(r *http.Request) l00p8.Response {
	matchId := l00p8.URLParam(r, "matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	match, err := api.mm.GetMatchById(mId)
	if err == ErrNotFound {
		return api.errSys.NotFound(414, err.Error())
	} else if err != nil {
		return api.errSys.BadRequest(420, err.Error())
	}
	err = api.mm.MakeTeams(match)
	if err != nil {
		return api.errSys.InternalServerError(520, err.Error())
	}
	return l00p8.OK("{}")
}

func (api *App) GetMatchParticipants(r *http.Request) l00p8.Response {
	matchId := l00p8.URLParam(r, "matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	players, err := api.mm.GetRegisteredPlayers(mId)
	if err != nil {
		return api.errSys.InternalServerError(510, err.Error())
	}
	return l00p8.OK(players)
}

func (api *App) RegisterInMatch(r *http.Request) l00p8.Response {
	matchId := l00p8.URLParam(r, "matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	claims := r.Context().Value("claims").(*shield.JWTClaims)
	err = api.mm.RegisterForMatch(mId, claims.Email)
	if err != nil {
		return api.errSys.InternalServerError(510, err.Error())
	}
	return l00p8.OK("{}")
}

func (api *App) UnregisterInMatch(r *http.Request) l00p8.Response {
	matchId := l00p8.URLParam(r, "matchId")
	mId, err := strconv.ParseInt(matchId, 10, 64)
	if err != nil {
		return api.errSys.BadRequest(410, err.Error())
	}
	claims := r.Context().Value("claims").(*shield.JWTClaims)
	err = api.mm.UnregisterForMatch(mId, claims.Email)
	if err != nil {
		return api.errSys.InternalServerError(510, err.Error())
	}
	return l00p8.OK("{}")
}
