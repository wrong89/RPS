package jwt_test

import (
	"rps/internal/domain/entities"
	jwt "rps/internal/lib/jwt/newToken"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	testCases := []struct {
		name          string
		player        entities.Player
		duration      time.Duration
		expectedError error
	}{
		{
			name: "Success",
			player: entities.Player{
				ID:    0,
				Email: "test@test.com",
			},
			duration: time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwt.NewToken(tc.player, tc.duration)
			require.ErrorIs(t, err, tc.expectedError)

			require.NotZero(t, token)
		})
	}
}
