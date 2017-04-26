package lib

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"github.com/badoux/checkmail"
	"sync"

	"errors"
)

var (
	ErrNotFound = errors.New("Item not found")
	ErrBadRequest = errors.New("Bad request")
	ErrPlayerAlredyExists = errors.New("Player already exists")
)

type (
	PlayerMaker struct {
		db *sql.DB
	}
)

func validatePassword(password string) error {
	if len(password) <= 3 {
		return ErrBadRequest
	}

	return nil
}

func validateEmail(email string) error {
	if err := checkmail.ValidateFormat(email); err != nil {
		return err
	}

	if err := checkmail.ValidateHost(email); err != nil {
		return err
	}

	return nil
}

func validatePlayer(player *Player) error {
	if err := validateEmail(player.Email); err != nil {
		return err
	}
	if player.DisplayName == "" {
		return ErrBadRequest
	}

	return nil
}

func (tb *PlayerMaker) Signup(player *Player, password string) error {
	if err := validatePlayer(player); err != nil {
		return err
	}

	if err := validatePassword(password); err != nil {
		return err
	}

	_, err := tb.getPlayerByEmail(player.Email)
	if err != ErrNotFound {
		return ErrPlayerAlredyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := tb.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO players (email, displayName, password, rating, ratingCount) VALUES " +
		"(?, ?, ?, ?, ?)",
		player.Email, player.DisplayName, hashedPassword, 0.0, 0)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (tb *PlayerMaker) getPlayerByEmail(email string) (*Player, error) {
	var dbEmail, dbDisplayName, dbPassword string
	var dbRating float32
	var dbRatingCount int

	err := tb.db.QueryRow("SELECT email, displayName, password, rating, ratingCount FROM players WHERE email = ?", email).
		Scan(&dbEmail, &dbDisplayName, &dbPassword, &dbRating, &dbRatingCount)

	if err == sql.ErrNoRows || dbEmail != email {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return &Player{dbDisplayName, email, dbPassword, sync.Mutex{}, dbRating, dbRatingCount}, nil
}

func (tb *PlayerMaker) Login(email, password string) (*Player, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if err := validatePassword(password); err != nil {
		return nil, err
	}

	player, err := tb.getPlayerByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(player.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (tb *PlayerMaker) ListPlayersToRate(player *Player) ([]PlayerInfo, error) {
	rows, err := tb.db.Query(`
		SELECT email, displayName, r.rating FROM players AS p
		LEFT JOIN ratings AS r ON p.email = r.rateeEmail AND r.raterEmail = ?
		WHERE p.email <> ?
	`, player.Email, player.Email)

	if err != nil {
		return nil, err
	}

	players := []PlayerInfo{}

	for rows.Next() {
		var email, displayName string
		var rating float32
		var ratingCount int

		rows.Scan(&email, &displayName, &rating);

		players = append(players, PlayerInfo{
			Email: email,
			Rating: rating,
			RatingCount: ratingCount,
			DisplayName: displayName,
		})
	}

	return players, nil
}

func (tb *PlayerMaker) GetAvgRate(rateeEmail string) (int, float32, error) {
	rows, err := tb.db.Query(`SELECT rating FROM ratings WHERE rateeEmail = ?`, rateeEmail)
	if err != nil {
		return 0, 0.0, err
	}
	cnt := 0
	var avg float32 = 0.0

	for rows.Next() {
		var rating float32
		if err := rows.Scan(&rating); err != nil {
			continue
		}
		cnt++
		avg += rating
	}
	avg /= float32(cnt)

	return cnt, avg, nil
}

func (tb *PlayerMaker) RatePlayer(raterEmail, rateeEmail string, rate int) error {
	rateFloat := float32(rate)

	ratee, err := tb.getPlayerByEmail(rateeEmail)
	if err != nil {
		return err
	}

	tx, err := tb.db.Begin()
	if err != nil {
		return err
	}

	var rating float32
	err = tx.QueryRow("SELECT rateeEmail, raterEmail, rating FROM ratings " +
		"WHERE rateeEmail = ? AND raterEmail = ?", rateeEmail, raterEmail).
		Scan(&rateeEmail, &raterEmail, &rating)

	if err == sql.ErrNoRows {
		_, err = tb.db.Exec("INSERT INTO ratings (rateeEmail, raterEmail, rating) VALUES (?, ?, ?)", ratee.Email, raterEmail, rateFloat)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else if err == nil {
		_, err = tb.db.Exec("UPDATE ratings SET rating = ? WHERE rateeEmail = ? AND raterEmail = ?", rateFloat, ratee.Email, raterEmail)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		tx.Rollback()
		return err
	}

	rateCount, rating, err := tb.GetAvgRate(rateeEmail)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE players SET rating = ?, ratingCount = ? WHERE email = ?", rating, rateCount, rateeEmail)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (tb *PlayerMaker) Connect(sqlUri string) error {
	db, err := sql.Open("mysql", sqlUri)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	tb.db = db
	return nil
}

func (tb *PlayerMaker) Close() error {
	return tb.db.Close()
}
