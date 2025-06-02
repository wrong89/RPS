package postgres

import (
	"context"
	"fmt"
	"os"
	"rps/internal/domain/entities"

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

func (s *Storage) SavePlayer(ctx context.Context, name, email string, password []byte) error {
	const op = "storage.postgres.SavePlayer"

	query := `INSERT INTO player (name, email, password) VALUES(@name, @email, @password)`
	args := pgx.NamedArgs{
		"name":     name,
		"email":    email,
		"password": password,
	}

	_, err := s.db.Exec(ctx, query, args)

	if err != nil {
		return fmt.Errorf("%s: unable to insert row: %w", op, err)
	}

	return nil
}

func (s *Storage) GetPlayer(ctx context.Context, email string) (entities.Player, error) {
	const op = "storage.postgres.GetPlayer"

	var res entities.Player

	query := `SELECT * FROM player WHERE email = @email`
	args := pgx.NamedArgs{
		"email": email,
	}

	row := s.db.QueryRow(ctx, query, args)
	err := row.Scan(
		&res.ID,
		&res.Name,
		&res.Email,
		&res.PassHash,
	)
	if err != nil {
		return res, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil

}
