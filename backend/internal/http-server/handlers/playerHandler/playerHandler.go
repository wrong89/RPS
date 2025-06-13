package playerHandler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"rps/internal/domain/entities"
	authMiddleware "rps/internal/http-server/middleware/auth"
)

type PlayerResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func ProfileHandler(log *slog.Logger, playerRepo entities.PlayerRepository) http.HandlerFunc {
	const op = "handlers.PlayerHandler.Profile"

	log = log.With(
		slog.String("op", op),
	)

	return func(w http.ResponseWriter, r *http.Request) {

		playerID, ok := authMiddleware.GetPlayerID(r)
		if !ok {
			errMsg := "Unauthorized"

			log.Error(errMsg)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		player, err := playerRepo.GetPlayerByID(r.Context(), playerID)
		if err != nil {
			errMsg := "User not found"

			log.Error(errMsg, slog.Int("id", playerID))
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}

		response := PlayerResponse{
			ID:    player.ID,
			Email: player.Email,
			Name:  player.Name,
		}

		log.Debug(
			"profile successfully received",
			slog.Int("id", response.ID),
			slog.String("name", response.Name),
			slog.String("email", response.Email),
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
