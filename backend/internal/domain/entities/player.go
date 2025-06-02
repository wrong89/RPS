package entities

type Player struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	PassHash []byte `json:"-"`
}

func CreateNewPlayer(name, email, password string) Player {
	var player Player

	player.Name = name
	player.Email = email
	player.PassHash = []byte(password)

	return player
}
