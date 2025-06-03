package jwt

import (
	"os"
	"rps/internal/domain/entities"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(player entities.Player, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = player.ID
	claims["email"] = player.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	secret := os.Getenv("JWT_TOKEN_SECRET")

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
