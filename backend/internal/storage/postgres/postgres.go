package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"rps/internal/domain/entities"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: pool}, nil
}

func (s *Storage) CreatePlayer(ctx context.Context, email, name, passwordHash string) (entities.Player, error) {
	player := entities.Player{
		Name:      name,
		Email:     email,
		PassHash:  passwordHash,
		CreatedAt: time.Now(),
	}

	query := `INSERT INTO player (name, email, password) VALUES(@name, @email, @password)`
	args := pgx.NamedArgs{
		"name":     name,
		"email":    email,
		"password": passwordHash,
	}

	_, err := s.db.Exec(ctx, query, args)
	if err != nil {
		return entities.Player{}, err
	}

	return player, nil
}

func (s *Storage) GetPlayerByEmail(ctx context.Context, email string) (entities.Player, error) {
	var res entities.Player
	var lastLogin sql.NullTime

	query := `SELECT id, name, email, password, created_at, last_login FROM player WHERE email = @email`
	args := pgx.NamedArgs{
		"email": email,
	}

	err := s.db.QueryRow(ctx, query, args).Scan(
		&res.ID,
		&res.Name,
		&res.Email,
		&res.PassHash,
		&res.CreatedAt,
		&lastLogin,
	)

	if err != nil {
		return res, err
	}

	if lastLogin.Valid {
		res.LastLogin = &lastLogin.Time
	}

	return res, nil
}

func (s *Storage) GetPlayerByID(ctx context.Context, id int) (entities.Player, error) {
	var res entities.Player
	var lastLogin sql.NullTime

	query := `SELECT id, name, email, password, created_at, last_login FROM player WHERE id = @id`
	args := pgx.NamedArgs{
		"id": id,
	}

	err := s.db.QueryRow(ctx, query, args).Scan(
		&res.ID,
		&res.Name,
		&res.Email,
		&res.PassHash,
		&res.CreatedAt,
		&lastLogin,
	)

	if err != nil {
		return res, err
	}

	if lastLogin.Valid {
		res.LastLogin = &lastLogin.Time
	}

	return res, nil
}
