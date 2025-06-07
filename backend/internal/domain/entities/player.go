package entities

import (
	"context"
	"time"
)

type Player struct {
	ID        int
	Name      string
	Email     string
	PassHash  string
	CreatedAt time.Time
	LastLogin *time.Time
}

type PlayerRepository interface {
	CreatePlayer(
		ctx context.Context,
		email,
		name,
		passwordHash string,
	) (Player, error)
	GetPlayerByEmail(
		ctx context.Context,
		email string,
	) (Player, error)
	GetPlayerByID(
		ctx context.Context,
		id int,
	) (Player, error)
}
