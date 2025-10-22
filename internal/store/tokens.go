package store

import (
	"database/sql"
	"time"

	"github.com/OlivierCoq/go_api_template/internal/tokens"
)

type PostgresTokenStore struct {
	db *sql.DB
}

func NewPostgresTokenStore(db *sql.DB) *PostgresTokenStore {
	return &PostgresTokenStore{db: db}
}

// Interface for TokenStore to allow decoupling and easier testing:
type TokenStore interface {
	Insert(token *tokens.Token) error
	CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error)
	DeleteAllTokensForUser(scope string, userID int) error
	RevokeToken(tokenHash string) error
}

// Insert a new token into the database
func (t *PostgresTokenStore) CreateNewToken(userID int, ttl time.Duration, scope string) (*tokens.Token, error) {
	token, err := tokens.GenerateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.Insert(token)
	return token, err
}

func (t *PostgresTokenStore) Insert(token *tokens.Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)
	`
	_, err := t.db.Exec(query, token.Hash, token.UserID, token.Expiry, token.Scope)
	return err
}

func (t *PostgresTokenStore) DeleteAllTokensForUser(scope string, userID int) error {
	query := `
		DELETE FROM tokens
		WHERE user_id = $1 AND scope = $2
	`
	_, err := t.db.Exec(query, userID, scope)
	return err
}

// Logging out:
// RevokeToken
func (t *PostgresTokenStore) RevokeToken(tokenHash string) error {
	query := `
		DELETE FROM tokens
		WHERE hash = $1
	`
	_, err := t.db.Exec(query, tokenHash)
	return err
}
