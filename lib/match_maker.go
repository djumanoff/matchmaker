package lib

import (
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"errors"
)

var (
	ErrMatchPassed = errors.New("Match was already played")
	ErrMatchIsFull = errors.New("Match players already stacked")
	ErrWrongNumberOfPlayers = errors.New("Number of players should be even to number of teams")
)

type (
	MatchMaker struct {
		db *sql.DB
	}

	Match struct {
		MatchID int64

		NumberOfTeams int
		NumberOfPlayersPerTeam int
		TotalPlayers int

		Rating float32

		Date time.Time

		Teams []Team
	}

	MatchTeams struct {
		MatchID int64
		TeamID int64
	}

	MatchPlayers struct {
		MatchID int64
		UserEmail string
	}
)

func (mm *MatchMaker) CreateMatch(date time.Time, numberOfTeams, numberOfPlayersPerTeam int) (*Match, error) {
	if numberOfPlayersPerTeam % numberOfTeams != 0 {
		return nil, ErrWrongNumberOfPlayers
	}

	match := &Match{
		NumberOfPlayersPerTeam: numberOfPlayersPerTeam,
		NumberOfTeams: numberOfTeams,
		Date: date,
	}
	match.TotalPlayers = numberOfTeams * numberOfPlayersPerTeam

	result, err := mm.db.Exec("INSERT INTO matches (numOfPlayersPerTeam, numOfTeams, totalPlayers, date) VALUES (?, ?, ?, ?)",
		numberOfPlayersPerTeam, numberOfTeams, match.TotalPlayers, date)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	match.MatchID = id

	return match, nil
}

func (mm *MatchMaker) GetMatchById(matchID int64) (*Match, error) {
	var numOfPlayersPerTeam, numOfTeams, totalPlayers int
	var date time.Time
	var rating float32

	err := mm.db.QueryRow("SELECT matchId, numOfPlayersPerTeam, numOfTeams, totalPlayers, date, rating FROM matches WHERE matchId = ?", matchID).
		Scan(&matchID, &numOfPlayersPerTeam, &numOfTeams, &totalPlayers, &date, &rating)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &Match{
		MatchID: matchID,
		NumberOfPlayersPerTeam: numOfPlayersPerTeam,
		NumberOfTeams: numOfTeams,
		TotalPlayers: totalPlayers,
		Date: date,
		Rating: rating,
	}, nil
}

func (mm *MatchMaker) GetNumberOfRegisteredPlayers(matchID int64) (int, error) {
	var numOfRegPlayers int

	err := mm.db.QueryRow("SELECT COUNT(playerEmail) FROM matchPlayers WHERE matchId = ?", matchID).Scan(&numOfRegPlayers)
	if err != nil {
		return 0, err
	}

	return numOfRegPlayers, nil
}

func (mm *MatchMaker) GetMatchPlayers(match *Match) ([]PlayerInfo, error) {
	rows, err := mm.db.Query(`SELECT email, rating, ratingCount, displayName FROM players AS p
		LEFT JOIN matchPlayers AS mp ON p.email = mp.playerEmail
		WHERE mp.matchId = ? ORDER BY rating DESC`, match.MatchID)

	if err != nil {
		return nil, err
	}

	players := []PlayerInfo{}

	for rows.Next() {
		var email, displayName string
		var rating float32
		var ratingCount int

		if err := rows.Scan(&email, &rating, &ratingCount, &displayName); err != nil {
			return nil, err
		}

		players = append(players, PlayerInfo{
			Email: email,
			Rating: rating,
			RatingCount: ratingCount,
			DisplayName: displayName,
		})
	}

	return players, nil
}

func (mm *MatchMaker) MakeTeams(match *Match) error {
	players, err := mm.GetMatchPlayers(match)
	if err != nil {
		return err
	}

	teams := make([]Team, match.NumberOfTeams)
	teamRatings := make([]float32, match.NumberOfTeams)

	i := 0
	j := len(players) - 1
	t := 0
	var rating float32 = 0.0

	for i < j {
		for t = 0; t < match.NumberOfTeams; t++ {
			teams[t].AddPlayer(players[i])
			teams[t].AddPlayer(players[j])
			rate := players[i].Rating + players[j].Rating
			rating += rate
			teamRatings[t] += rate
			i++
			i--
		}
	}

	tx, err := mm.db.Begin()
	if err != nil {
		return err
	}

	for t, team := range teams {
		result, err := tx.Exec("INSERT INTO teams (name, rating) VALUES (?, ?)", "", teamRatings[t] / float32(match.NumberOfPlayersPerTeam))
		if err != nil {
			tx.Rollback()
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}
		teams[t].TeamID = id

		for _, player := range team.Players {
			_, err := tx.Exec("INSERT INTO teamPlayers (teamId, email) VALUES (?, ?)", id, player.Email)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		_, err = tx.Exec("INSERT INTO matchTeams (matchId, teamId) VALUES (?, ?)", match.MatchID, id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	rating /= float32(len(players))
	match.Rating = rating
	match.Teams = teams

	_, err = tx.Exec("UPDATE matches SET rating = ? WHERE matchId = ?", rating, match.MatchID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (mm *MatchMaker) RegisterForMatch(matchID int64, playerEmail string) error {
	match, err := mm.GetMatchById(matchID)
	if err != nil {
		return err
	}

	if match.Date.Unix() < time.Now().Unix() {
		return ErrMatchPassed
	}

	numOfParts, err := mm.GetNumberOfRegisteredPlayers(matchID)
	if err != nil {
		return err
	}
	if numOfParts >= match.TotalPlayers {
		return ErrMatchIsFull
	}

	_, err = mm.db.Exec("INSERT INTO matchPlayers (matchId, playerEmail) VALUES (?, ?)", matchID, playerEmail)
	if err != nil {
		return err
	}

	return nil
}

func (mm *MatchMaker) Connect(sqlUri string) error {
	db, err := sql.Open("mysql", sqlUri)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	mm.db = db
	return nil
}

func (mm *MatchMaker) Close() error {
	return mm.db.Close()
}
