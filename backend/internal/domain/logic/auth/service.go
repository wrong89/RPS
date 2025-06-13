package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"rps/internal/domain/entities"
	"rps/internal/lib/logger/sl"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
	ErrEmailInUse         = errors.New("email already in use")
)

// ? Think about refactoring
// Like using the interface with behavior with for example determine the methods for player
// This dependency inversion easily do low-coupling
type AuthService struct {
	log              *slog.Logger
	playerRepo       entities.PlayerRepository
	refreshTokenRepo entities.RefreshTokenRepository
	jwtSecret        []byte
	accessTokenTTL   time.Duration
}

func NewAuthService(
	log *slog.Logger,
	playerRepo entities.PlayerRepository,
	refreshTokenRepo entities.RefreshTokenRepository,
	jwtSecret string,
	accessTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		log:              log,
		playerRepo:       playerRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        []byte(jwtSecret),
		accessTokenTTL:   accessTokenTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, name, password string) (entities.Player, error) {
	const op = "auth.service.Register"

	s.log = s.log.With(
		slog.String("op", op),
		slog.String("name", name),
		slog.String("email", email),
	)

	var player entities.Player

	_, err := s.playerRepo.GetPlayerByEmail(ctx, email)

	if err == nil {
		s.log.Error("player with this email already exists", sl.Err(err))
		return player, ErrEmailInUse
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return player, err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return player, err
	}

	s.log.Debug("trying to create player")

	player, err = s.playerRepo.CreatePlayer(ctx, email, name, hashedPassword)
	if err != nil {
		return player, err
	}

	s.log.Debug("player successfully created")

	return player, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "auth.service.Login"

	s.log = s.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	player, err := s.playerRepo.GetPlayerByEmail(ctx, email)
	if err != nil {
		s.log.Error(ErrInvalidCredentials.Error(), sl.Err(err))
		return "", ErrInvalidCredentials
	}

	if err := VerifyPassword(player.PassHash, password); err != nil {
		s.log.Error(ErrInvalidCredentials.Error(), sl.Err(err))
		return "", ErrInvalidCredentials
	}

	token, err := s.generateAccessToken(&player)
	if err != nil {
		s.log.Error("err", sl.Err(err))
		return "", err
	}

	return token, err
}

func (s *AuthService) LoginWithRefresh(
	ctx context.Context,
	email,
	password string,
	refreshTokenTTL time.Duration,
) (accessToken, refreshToken string, err error) {
	const op = "auth.service.LoginWithRefresh"
	player, err := s.playerRepo.GetPlayerByEmail(ctx, email)

	s.log = s.log.With(
		slog.String("op", op),
		slog.String("player name", player.Name),
		slog.String("player email", player.Email),
	)

	if err != nil {
		s.log.Error(ErrInvalidCredentials.Error())
		return "", "", ErrInvalidCredentials
	}

	if err := VerifyPassword(player.PassHash, password); err != nil {
		s.log.Error(ErrInvalidCredentials.Error())
		return "", "", ErrInvalidCredentials
	}

	accessToken, err = s.generateAccessToken(&player)
	if err != nil {
		s.log.Error("err", sl.Err(err))
		return "", "", err
	}

	token, err := s.refreshTokenRepo.CreateRefreshToken(ctx, player.ID, refreshTokenTTL)
	if err != nil {
		s.log.Error("err", sl.Err(err))
		return "", "", err
	}

	return accessToken, token.Token, nil
}

func (s *AuthService) generateAccessToken(player *entities.Player) (string, error) {
	const op = "auth.service.generateAccessToken"

	s.log = s.log.With(
		slog.String("op", op),
		// slog.String("email", player.Email),
		slog.Any("player", player),
	)

	expirationTime := time.Now().Add(s.accessTokenTTL)

	claims := jwt.MapClaims{
		"sub":   player.ID,
		"name":  player.Name,
		"email": player.Email,
		"exp":   expirationTime.Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.log.Error("signing token error", sl.Err(err))
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	const op = "auth.service.ValidateToken"

	s.log.With(
		slog.String("op", op),
		slog.String("token", tokenString),
	)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.log.Error(ErrInvalidToken.Error())
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			s.log.Error(ErrExpiredToken.Error())
			return nil, ErrExpiredToken
		}

		s.log.Error(ErrInvalidToken.Error())
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	s.log.Error(ErrInvalidToken.Error())
	return nil, ErrInvalidToken
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshTokenString string) (string, error) {
	const op = "auth.service.RefreshAccessToken"

	s.log = s.log.With(
		slog.String("op", op),
	)

	token, err := s.refreshTokenRepo.GetRefreshToken(ctx, refreshTokenString)
	fmt.Println("[SERVICE] REFRESH_TOKEN", token)
	if err != nil {
		s.log.Error(ErrInvalidToken.Error())
		return "", ErrInvalidToken
	}

	if token.Revoked {
		s.log.Error(ErrInvalidToken.Error())
		return "", ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		s.log.Error(ErrExpiredToken.Error())
		return "", ErrExpiredToken
	}

	player, err := s.playerRepo.GetPlayerByID(ctx, token.PlayerID)
	if err != nil {
		s.log.Error("err", sl.Err(err))
		return "", err
	}

	accessToken, err := s.generateAccessToken(&player)
	if err != nil {
		s.log.Error("err", sl.Err(err))
		return "", err
	}

	return accessToken, nil
}
