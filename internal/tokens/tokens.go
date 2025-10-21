package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

// Scope
const (
	ScopeAuth = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`  // The plaintext token to be sent to the user
	Hash      []byte    `json:"-"`      // The hashed version of the token for secure storage
	UserID    int       `json:"-"`      // ID of the user the token is associated with
	Expiry    time.Time `json:"expiry"` // Unix timestamp
	Scope     string    `json:"-"`      // e.g., "authentication", "password_reset"
}

func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	// Implementation for generating a new token
	// This typically involves creating a random string for the plaintext token,
	// hashing it for secure storage, and setting the expiry time based on the ttl.
	// For simplicity, this is a placeholder implementation.

	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	emptyBytes := make([]byte, 16) // Placeholder for random bytes
	_, err := rand.Read(emptyBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(emptyBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil

}
