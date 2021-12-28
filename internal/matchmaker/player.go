package matchmaker

import (
	"sync"
)

type (
	Player struct {
		DisplayName string
		Email       string
		Password    string

		mu          sync.Mutex
		Rating      float32
		RatingCount int
	}

	PlayerInfo struct {
		DisplayName string
		Email       string
		Rating      float32
		RatingCount int
	}

	Rating struct {
		RaterEmail string
		RateeEmail string

		Rating float32
	}
)

func (u *Player) Rated(rate float32) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.Rating += rate
	u.Rating /= 2
	u.RatingCount++
}
