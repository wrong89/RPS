package authHandler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
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
	RegisterNewPlayer(
		ctx context.Context,
		name,
		email,
		pass string,
	) error
}

var (
	errInvalidEmail = errors.New("invalid email")
	errLogin        = errors.New("login error")
	errRegister     = errors.New("register error")
)

func LoginHandler(logger *slog.Logger, auth authActions) http.HandlerFunc {
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

		token, err := auth.Login(r.Context(), body["email"], body["password"])
		if err != nil {
			logger.Error(errLogin.Error(), sl.Err(err))
			http.Error(w, errLogin.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		response := struct {
			Token string `json:"token"`
		}{
			Token: token,
		}

		json.NewEncoder(w).Encode(response)
	}
}

func RegisterHandler(logger *slog.Logger, auth authActions) http.HandlerFunc {
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

		err = auth.RegisterNewPlayer(
			r.Context(),
			body["name"],
			body["email"],
			body["password"],
		)
		if err != nil {
			logger.Error(errRegister.Error(), sl.Err(err))
			http.Error(w, errRegister.Error(), http.StatusBadRequest)
			return
		}

	}
}
