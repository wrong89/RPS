package authHandler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"rps/internal/domain/entities"
	"rps/internal/domain/logic/auth"
	"rps/internal/http-server/helpers"
	"rps/internal/lib/logger/sl"
	"rps/internal/utils/validators"
)

type authActions interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (string, error)
	Register(
		ctx context.Context,
		email,
		name,
		password string,
	) (entities.Player, error)
	RefreshAccessToken(
		ctx context.Context,
		refreshTokenString string,
	) (string, error)
}

var (
	errInvalidEmail = errors.New("invalid email")
	errLogin        = errors.New("login error")
	errRegister     = errors.New("register error")
)

type LoginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(logger *slog.Logger, auth authActions) http.HandlerFunc {
	const op = "handlers.authHandler.LoginHandler"

	logger = logger.With(
		slog.String("op", op),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := helpers.GetValidatedBody(r, "email", "password")
		if err != nil {
			logger.Error(err.Error(), sl.Err(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		valid := validators.ValidateEmail(body["email"])
		if !valid {
			logger.Error(errInvalidEmail.Error(), slog.String("email", body["email"]))
			http.Error(w, errInvalidEmail.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("LoginBody", slog.Any("body", body))

		accessToken, err := auth.Login(
			r.Context(),
			body["email"],
			body["password"],
		)
		if err != nil {
			logger.Error(errLogin.Error(), sl.Err(err))
			http.Error(w, errLogin.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug(
			"player successfully login in",
			slog.String("email", body["email"]),
		)

		w.Header().Set("Content-Type", "application/json")

		response := LoginResponse{Token: accessToken}

		json.NewEncoder(w).Encode(response)
	}
}

type RegisterResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func RegisterHandler(logger *slog.Logger, auth authActions) http.HandlerFunc {
	const op = "handlers.authHandler.RegisterHandler"

	logger = logger.With(
		slog.String("op", op),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := helpers.GetValidatedBody(r, "email", "password", "name")
		if err != nil {
			logger.Error(err.Error(), sl.Err(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		valid := validators.ValidateEmail(body["email"])
		if !valid {
			logger.Error(errInvalidEmail.Error(), slog.String("email", body["email"]))
			http.Error(w, errInvalidEmail.Error(), http.StatusBadRequest)
			return
		}

		logger.Debug("RegisterBody", slog.Any("body", body))

		player, err := auth.Register(
			r.Context(),
			body["email"],
			body["name"],
			body["password"],
		)

		if err != nil {
			logger.Error(errRegister.Error(), sl.Err(err))
			http.Error(w, errRegister.Error(), http.StatusBadRequest)
			return
		}

		response := RegisterResponse{
			ID:    player.ID,
			Email: player.Email,
			Name:  player.Name,
		}

		logger.Debug(
			"player successfully registered",
			slog.Int("id", response.ID),
			slog.String("name", response.Name),
			slog.String("email", response.Email),
		)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(response)
	}
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	Token string `json:"token"`
}

// RefreshToken handles access token refresh
func RefreshTokenHandler(logger *slog.Logger, authActions authActions) http.HandlerFunc {
	const op = "handlers.authHandler.RefreshTokenHandler"

	logger = logger.With(
		slog.String("op", op),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := helpers.GetValidatedBody(r, "refresh_token")
		if err != nil {
			logger.Error(err.Error(), sl.Err(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := authActions.RefreshAccessToken(r.Context(), body["refresh_token"])
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) {
				errMsg := "Invalid or expired refresh token"
				logger.Error(errMsg, sl.Err(err))
				http.Error(w, errMsg, http.StatusUnauthorized)
			} else {
				errMsg := "Internal Server Error"
				logger.Error(errMsg, sl.Err(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		logger.Debug("access_token successfully refreshed")

		response := RefreshResponse{Token: token}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
