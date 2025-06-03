package authHandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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
	errInvalidBody  = errors.New("invalid body")
	errInvalidEmail = errors.New("invalid email")
	errLogin        = errors.New("login error")
	errRegister     = errors.New("register error")
)

type testStruct struct {
	Token      string `json:"token"`
	SomeRandom string `json:"some_random"`
}

type bodyData = map[string]string

func LoginHandler(logger *slog.Logger, auth authActions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := getBodyFromRequest(r)
		if err != nil {
			logger.Error(errInvalidBody.Error(), sl.Err(err))
			http.Error(w, errInvalidBody.Error(), http.StatusBadRequest)
			return
		}

		valid := validators.ValidateBody(body, "email", "password")
		if !valid {
			logger.Error("body is invalid")
			http.Error(w, errInvalidBody.Error(), http.StatusBadRequest)
			return
		}

		valid = validators.ValidateEmail(body["email"])
		if !valid {
			logger.Error("email is invalid")
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

		response := testStruct{
			Token:      token,
			SomeRandom: "Hello World",
		}

		json.NewEncoder(w).Encode(response)
	}
}

func RegisterHandler(logger *slog.Logger, auth authActions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := getBodyFromRequest(r)
		if err != nil {
			logger.Error(errInvalidBody.Error(), sl.Err(err))
			http.Error(w, errInvalidBody.Error(), http.StatusBadRequest)
			return
		}

		valid := validators.ValidateBody(body, "name", "email", "password")
		if !valid {
			logger.Error("body is invalid")
			http.Error(w, errInvalidBody.Error(), http.StatusBadRequest)
			return
		}

		valid = validators.ValidateEmail(body["email"])
		if !valid {
			logger.Error("email is invalid")
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

		w.Header().Set("Content-Type", "application/json")

		response := testStruct{
			SomeRandom: "success",
		}

		json.NewEncoder(w).Encode(response)
	}
}

func getBodyFromRequest(r *http.Request) (bodyData, error) {
	const op = "authHandler.getBodyFromRequest"
	defer r.Body.Close()

	var result bodyData

	err := json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		return result, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}
