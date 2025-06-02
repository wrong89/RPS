package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"rps/internal/domain/entities"
	jwt "rps/internal/lib/jwt/newToken"
	"rps/internal/lib/logger/sl"
	"rps/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log            *slog.Logger
	playerSaver    PlayerSaver
	playerProvider PlayerProvider
	tokenTTL       time.Duration
}

type PlayerSaver interface {
	SavePlayer(
		ctx context.Context,
		name string,
		email string,
		password []byte,
	) (err error)
}

type PlayerProvider interface {
	GetPlayer(ctx context.Context, email string) (entities.Player, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrPlayerExists       = errors.New("player already exists")
)

func New(
	log *slog.Logger,
	playerSaver PlayerSaver,
	playerProvider PlayerProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:            log,
		playerSaver:    playerSaver,
		playerProvider: playerProvider,
		tokenTTL:       tokenTTL,
	}
}

// Login checks if player with given credentials exists in the system
//
// If player exists, but password is incorrect, returns error.
// If player doesn't exist, returns error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login player")

	player, err := a.playerProvider.GetPlayer(ctx, email)
	if err != nil {
		if errors.Is(err, ErrPlayerExists) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(player.PassHash, []byte(password)); err != nil {
		a.log.Info(ErrInvalidCredentials.Error(), sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("player successfully logged in")

	token, err := jwt.NewToken(player, a.tokenTTL)
	if err != nil {
		a.log.Info("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewPlayer registers new player in the system and returns playerID.
// If the player with given email exists, returns error.
func (a *Auth) RegisterNewPlayer(ctx context.Context, name, email, pass string) error {
	const op = "auth.RegisterNewPlayer"

	log := a.log.With(
		slog.String("op", op),
		slog.String("username", name),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	err = a.playerSaver.SavePlayer(ctx, name, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrPlayerExists) {
			return fmt.Errorf("%s: %w", op, ErrPlayerExists)
		}

		log.Error("failed to save player", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return nil
}
