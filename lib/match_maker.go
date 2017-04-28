package lib

import (
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"errors"
	"log"
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
		RegCount int

		Rating float32

		Date time.Time

		Teams []Team
	}

	MatchPlayers struct {
		MatchID int64
		UserEmail string
	}
)

func (mm *MatchMaker) GetUpcomingMatches() ([]Match, error) {
	rows, err := mm.db.Query(`SELECT m.matchId, numOfTeams, numOfPlayersPerTeam, totalPlayers, date, rating, COUNT(mp.playerEmail) AS regCount
		FROM matches AS m
		LEFT JOIN matchPlayers AS mp ON m.matchId = mp.matchId
		WHERE m.date >= CURDATE()
		GROUP BY m.matchId
		ORDER BY date ASC`)

	if err != nil {
		return nil, err
	}

	matches := []Match{}

	for rows.Next() {
		var matchId int64
		var numOfTeams, numOfPlayersPerTeam, totalPlayers, regCount int
		var date string
		var rating float32

		if err := rows.Scan(&matchId, &numOfTeams, &numOfPlayersPerTeam, &totalPlayers, &date, &rating, &regCount); err != nil {
			return nil, err
		}

		timeVal, err := time.Parse("2006-01-02 15:04:05", date)
		if err != nil {
			return nil, err
		}

		matches = append(matches, Match{
			MatchID: matchId,
			Rating: rating,
			NumberOfTeams: numOfTeams,
			NumberOfPlayersPerTeam: numOfPlayersPerTeam,
			TotalPlayers: totalPlayers,
			Date: timeVal,
			RegCount: regCount,
		})
	}

	return matches, nil
}

func (mm *MatchMaker) CreateMatch(date time.Time, numberOfTeams, numberOfPlayersPerTeam int) (*Match, error) {
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
	var date string
	var rating float32

	err := mm.db.QueryRow("SELECT matchId, numOfPlayersPerTeam, numOfTeams, totalPlayers, date, rating FROM matches WHERE matchId = ?", matchID).
		Scan(&matchID, &numOfPlayersPerTeam, &numOfTeams, &totalPlayers, &date, &rating)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	timeVal, err := time.Parse("2006-01-02 15:04:05", date)
	if err != nil {
		return nil, err
	}

	return &Match{
		MatchID: matchID,
		NumberOfPlayersPerTeam: numOfPlayersPerTeam,
		NumberOfTeams: numOfTeams,
		TotalPlayers: totalPlayers,
		Date: timeVal,
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

func (mm *MatchMaker) FormTeams(players []PlayerInfo, numOfTeams int) ([]Team) {
	teams := make([]Team, numOfTeams)

	i := 0
	j := len(players) - 1
	t := 0

	evenNumPlayers := len(players) % 2 == 0

	for i < j {
		for t = 0; t < numOfTeams; t++ {
			if j - i > numOfTeams - 1 || evenNumPlayers {
				teams[t].AddPlayer(players[i])
				teams[t].AddPlayer(players[j])
				rate := players[i].Rating + players[j].Rating
				teams[t].Rating += rate
				j--
			} else {
				teams[t].AddPlayer(players[i])
				teams[t].Rating += players[i].Rating
			}
			i++
		}
	}

	for t = range teams {
		teams[t].Rating /= float32(len(teams[t].Players))
	}

	return teams
}

func (mm *MatchMaker) GetTeams(match *Match) (map[int64]*Team, error) {
	rows, err := mm.db.Query(`
		SELECT displayName, email, p.rating AS rating, tp.teamId AS teamId, t.rating AS teamRating FROM players AS p
		LEFT JOIN teamPlayers AS tp
		ON p.email = tp.playerEmail
		LEFT JOIN teams AS t
		ON t.teamId = tp.teamId
		WHERE t.matchId = ?
	`, match.MatchID)

	if err != nil {
		return nil, err
	}

	teams := map[int64]*Team{}

	for rows.Next() {
		var email, displayName string
		var rating, teamRating float32
		var teamId int64

		if err := rows.Scan(&displayName, &email, &rating, &teamId, &teamRating); err != nil {
			return nil, err
		}

		player := PlayerInfo{
			Email: email,
			Rating: rating,
			DisplayName: displayName,
		}
		team, ok := teams[teamId]

		if !ok {
			team = &Team{TeamID: teamId, Rating: teamRating}
			teams[teamId] = team
		}

		team.AddPlayer(player)
	}

	return teams, nil
}

func (mm *MatchMaker) MakeTeams(match *Match) error {
	players, err := mm.GetMatchPlayers(match)
	if err != nil {
		return err
	}
	log.Println("players = ", players)

	teams := mm.FormTeams(players, match.NumberOfTeams)
	log.Println("teams = ", teams)

	tx, err := mm.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM teams WHERE matchId = ?", match.MatchID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM teamPlayers WHERE matchId = ?", match.MatchID)
	if err != nil {
		tx.Rollback()
		return err
	}

	var rating float32 = 0.0

	for t, team := range teams {
		rating += teams[t].Rating

		log.Println("INSERT INTO teams (name, rating) VALUES (?, ?, ?)", "", teams[t].Rating, match.MatchID)
		result, err := tx.Exec("INSERT INTO teams (name, rating, matchId) VALUES (?, ?, ?)", "", teams[t].Rating, match.MatchID)
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
			log.Println("INSERT INTO teamPlayers (teamId, playerEmail, matchId) VALUES (?, ?, ?)", id, player.Email, match.MatchID)

			_, err := tx.Exec("INSERT INTO teamPlayers (teamId, playerEmail, matchId) VALUES (?, ?, ?)", id, player.Email, match.MatchID)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
		//log.Println("INSERT INTO matchTeams (matchId, teamId) VALUES (?, ?)", match.MatchID, id)
		//_, err = tx.Exec("INSERT INTO matchTeams (matchId, teamId) VALUES (?, ?)", match.MatchID, id)
		//if err != nil {
		//	tx.Rollback()
		//	return err
		//}
	}

	rating /= float32(len(teams))
	match.Rating = rating
	match.Teams = teams

	log.Println("UPDATE matches SET rating = ? WHERE matchId = ?", rating, match.MatchID)
	_, err = tx.Exec("UPDATE matches SET rating = ? WHERE matchId = ?", rating, match.MatchID)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (mm *MatchMaker) GetRegisteredPlayers(matchID int64) ([]PlayerInfo, error) {
	rows, err := mm.db.Query("SELECT p.email, p.displayName, p.rating FROM matchPlayers AS mp LEFT JOIN players AS p ON mp.playerEmail = p.Email WHERE mp.matchId = ?", matchID)

	if err != nil {
		return nil, err
	}

	players := []PlayerInfo{}

	for rows.Next() {
		var email, displayName string
		var rating float32

		if err := rows.Scan(&email, &displayName, &rating); err != nil {
			return nil, err
		}

		players = append(players, PlayerInfo{
			Email: email,
			DisplayName: displayName,
			Rating: rating,
		})
	}

	return players, nil
}

func (mm *MatchMaker) GetPlayerRegisteredMatchIds(email string) (map[int64]bool, error) {
	rows, err := mm.db.Query("SELECT matchId FROM matchPlayers WHERE playerEmail = ?", email)

	if err != nil {
		return nil, err
	}

	ids := map[int64]bool{}

	for rows.Next() {
		var matchId int64

		if err := rows.Scan(&matchId); err != nil {
			return nil, err
		}

		ids[matchId] = true
	}

	return ids, nil
}

func (mm *MatchMaker) UnregisterForMatch(matchID int64, playerEmail string) error {
	match, err := mm.GetMatchById(matchID)
	if err != nil {
		return err
	}

	if match.Date.Unix() < time.Now().Unix() {
		return ErrMatchPassed
	}

	_, err = mm.db.Exec("DELETE FROM matchPlayers WHERE matchId = ? AND playerEmail = ? LIMIT 1", matchID, playerEmail)
	if err != nil {
		return err
	}

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
