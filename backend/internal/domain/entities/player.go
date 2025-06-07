package entities

import (
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
