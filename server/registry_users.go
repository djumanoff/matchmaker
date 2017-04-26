package server

import (
	"github.com/djumanoff/matchmaker/lib"
	"net/http"
	"time"
	"github.com/satori/go.uuid"
	"errors"
)

var users = &UserRegistry{map[string]*lib.Player{}}
var expiration = time.Now().Add(365 * 24 * time.Hour)

var ErrNoSession = errors.New("No session")

type UserRegistry struct {
	data map[string]*lib.Player
}

func (reg *UserRegistry) Save(player *lib.Player, w http.ResponseWriter) {
	token := uuid.NewV4().String()

	cookie := http.Cookie{
		Name: "token",
		Value: token,
		Path: "/",
		HttpOnly: true,
		Expires: expiration,
		MaxAge: int(365 * 24 * time.Hour),
	}

	http.SetCookie(w, &cookie)
	reg.data[token] = player
}

func (reg *UserRegistry) Get(r *http.Request) (*lib.Player, error) {
	cookie, err := r.Cookie("token")

	if err != nil {
		return nil, err
	}

	if cookie == nil {
		return nil, ErrNoSession
	}

	return reg.data[cookie.Value], nil
}

func (reg *UserRegistry) Delete(r *http.Request, w http.ResponseWriter) error {
	cookie, err := r.Cookie("token")
	if err != nil {
		return err
	}

	delete(reg.data, cookie.Value)
	http.SetCookie(w, nil)

	return nil
}
