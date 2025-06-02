package entities

type Lobby struct {
	LobbyId int
	LobbySettings
}

type LobbySettings struct {
	Title        string
	Password     []byte
	Players      []Player
	PlayersCount int
}
