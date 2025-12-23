package services_test

import (
	"testing"

	"go-reasonable-api/app/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSecureToken(t *testing.T) {
	tests := []struct {
		name           string
		length         int
		expectedLength int
	}{
		{
			name:           "generates token with length 16",
			length:         16,
			expectedLength: 32, // hex encoding doubles the length
		},
		{
			name:           "generates token with length 32",
			length:         32,
			expectedLength: 64,
		},
		{
			name:           "generates token with length 64",
			length:         64,
			expectedLength: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := services.GenerateSecureToken(tt.length)

			require.NoError(t, err)
			assert.Len(t, token, tt.expectedLength)
		})
	}
}

func TestGenerateSecureToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	iterations := 100

	for range iterations {
		token, err := services.GenerateSecureToken(32)
		require.NoError(t, err)

		assert.False(t, tokens[token], "generated duplicate token")
		tokens[token] = true
	}
}

func TestHashToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "hashes simple token",
			token: "simple-token",
		},
		{
			name:  "hashes empty token",
			token: "",
		},
		{
			name:  "hashes long token",
			token: "a-very-long-token-that-should-still-produce-a-fixed-length-hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := services.HashToken(tt.token)

			// SHA256 produces 64 hex characters
			assert.Len(t, hash, 64)
		})
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	token := "test-token"

	hash1 := services.HashToken(token)
	hash2 := services.HashToken(token)

	assert.Equal(t, hash1, hash2, "same token should produce same hash")
}

func TestHashToken_DifferentInputs(t *testing.T) {
	token1 := "token-1"
	token2 := "token-2"

	hash1 := services.HashToken(token1)
	hash2 := services.HashToken(token2)

	assert.NotEqual(t, hash1, hash2, "different tokens should produce different hashes")
}
