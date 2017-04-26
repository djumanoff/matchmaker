package lib

type (
	Team struct {
		TeamID int64
		Name string
		Rating float32

		Players []PlayerInfo
	}

	TeamPlayers struct {
		TeamID int64
		PlayerEmail string
	}
)

func (t *Team) AddPlayer(player PlayerInfo) {
	t.Players = append(t.Players, player)
}
