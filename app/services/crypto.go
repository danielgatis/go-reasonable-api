package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/rotisserie/eris"
)

// GenerateSecureToken generates a cryptographically secure random token.
// The length parameter specifies the number of random bytes to generate.
// The returned token is hex-encoded, so its string length will be 2*length.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", eris.Wrap(err, "failed to generate secure token")
	}
	return hex.EncodeToString(bytes), nil
}

// HashToken generates a SHA-256 hash of the given token.
// This is used to store tokens securely in the database - only the hash
// is stored, never the original token. The returned hash is hex-encoded.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
