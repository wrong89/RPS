package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"rps/internal/domain/entities"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	pool.Reset()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: pool}, nil
}

func (s *Storage) CloseDb() {
	s.db.Close()
}

// CreatePlayerAndStatistic creates records in player and statistic tables and returns id of new player.
func (s *Storage) CreatePlayerAndStatistic(ctx context.Context, email, name, passwordHash string) (entities.Player, error) {
	var newPlayer entities.Player

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return entities.Player{}, err
	}
	defer tx.Rollback(ctx)

	newPlayer, err = s.CreatePlayer(ctx, email, name, passwordHash)
	if err != nil {
		return entities.Player{}, err
	}

	statisticID, err := s.CreateDefaultStatistic(ctx)
	if err != nil {
		return entities.Player{}, err
	}

	err = s.CreatePlayerStatistic(ctx, newPlayer.ID, statisticID)
	if err != nil {
		return entities.Player{}, err
	}

	tx.Commit(ctx)
	return newPlayer, nil
}

// CreatePlayerStatistic inserts data into player_statistic table.
func (s *Storage) CreatePlayerStatistic(ctx context.Context, playerID, statisticID int) error {
	query := `INSERT INTO player_statistic (player_id, statistic_id) VALUES(@playerID, @statisticID)`
	args := pgx.NamedArgs{
		"playerID":    playerID,
		"statisticID": statisticID,
	}

	_, err := s.db.Exec(ctx, query, args)
	return err
}

func (s *Storage) CreatePlayer(ctx context.Context, email, name, passwordHash string) (entities.Player, error) {
	player := entities.Player{
		Name:      name,
		Email:     email,
		PassHash:  passwordHash,
		CreatedAt: time.Now(),
	}

	query := `INSERT INTO player (name, email, password, created_at) VALUES(@name, @email, @password, @created_at) RETURNING id;`
	args := pgx.NamedArgs{
		"name":       name,
		"email":      email,
		"password":   passwordHash,
		"created_at": player.CreatedAt,
	}

	err := s.db.QueryRow(ctx, query, args).Scan(
		&player.ID,
	)
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

func (s *Storage) CreateDefaultStatistic(ctx context.Context) (int, error) {
	var id int

	query := `INSERT INTO statistic DEFAULT VALUES RETURNING id;`

	err := s.db.QueryRow(ctx, query).Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (s *Storage) CreateRefreshToken(ctx context.Context, playerID int, ttl time.Duration) (entities.RefreshToken, error) {
	tokenID := uuid.New()
	expiresAt := time.Now().Add(ttl)

	token := entities.RefreshToken{
		ID:        tokenID,
		PlayerID:  playerID,
		Token:     tokenID.String(),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	query := `
	INSERT INTO refresh_tokens (player_id, token, expires_at, created_at, revoked)
	VALUES (@player_id, @token, @expires_at, @created_at, @revoked);
	`
	args := pgx.NamedArgs{
		"player_id":  playerID,
		"token":      tokenID.String(),
		"expires_at": expiresAt,
		"created_at": time.Now(),
		"revoked":    token.Revoked,
	}

	_, err := s.db.Exec(ctx, query, args)
	fmt.Println("SUCCESS INSERTING NEW REFRESH_TOKEN")
	if err != nil {
		fmt.Println("FAILED")
		return entities.RefreshToken{}, err
	}

	return token, nil
}

func (s *Storage) GetRefreshToken(ctx context.Context, tokenString string) (entities.RefreshToken, error) {
	query := `
		SELECT id, player_id, token, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token = @token;
	`

	args := pgx.NamedArgs{
		"token": tokenString,
	}

	var token entities.RefreshToken
	err := s.db.QueryRow(ctx, query, args).Scan(
		&token.ID,
		&token.PlayerID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.Revoked,
	)

	fmt.Println("TOKEN", token.Token)

	if err != nil {
		return token, err
	}

	return token, nil
}

func (s *Storage) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = true
		WHERE token = @token
	`

	args := pgx.NamedArgs{
		"token": tokenString,
	}

	_, err := s.db.Exec(ctx, query, args)
	return err
}
