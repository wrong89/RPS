package auth

import (
	"context"
	"log/slog"
	"net/http"
	"rps/internal/domain/logic/auth"
	"strings"
)

type contextKey string

const (
	PlayerIDKey contextKey = "playerID"
)

func AuthMiddleware(log *slog.Logger, authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/auth"),
		)

		log.Info("auth middleware is enabled")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errMsg := "Authorization header required"

				log.Error(errMsg)
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				errMsg := "Invalid authorization format"

				log.Error(errMsg, slog.String("authHeader", authHeader))
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				errMsg := "Invalid or expired token"

				log.Error(errMsg, slog.String("token", tokenString))
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}

			playerIDInt, ok := claims["sub"].(float64)
			if !ok {
				errMsg := "Invalid token claims"

				log.Error(errMsg, slog.Any("claims", claims))
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), PlayerIDKey, int(playerIDInt))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetPlayerID retrieves the player ID from the request context
func GetPlayerID(r *http.Request) (int, bool) {
	playerID, ok := r.Context().Value(PlayerIDKey).(int)
	return playerID, ok
}
