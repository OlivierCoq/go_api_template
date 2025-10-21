package store

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type password struct {
	plainText *string
	hash      []byte
}

// Password hashing and verification methods can be added here
/*
	In software development, a salt is a value that determins the computational complexity of the hashing algorithm.
	It is used to make password hashing more secure by increasing the time it takes to compute the hash.
	The higher the cost, the more computationally expensive it is to hash a password, which makes it harder for attackers to brute-force or crack passwords.

	Bcrypt uses a cost parameter that determines how many rounds of hashing are performed.
	The default cost is 10, but it can be increased for added security.
	However, increasing the cost also increases the time it takes to hash a password, so it's important to find a balance between security and performance.
*/
func (p *password) Set(plainText string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainText), 12)
	if err != nil {
		return err
	}
	p.plainText = &plainText
	p.hash = hash
	return nil
}

// For verifying if the provided plaintext password matches the stored hash. Authentication and login purposes.
func (p *password) Matches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainText))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			switch {
			case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
				return false, nil
			default:
				return false, err
			}
		}
	}
	return true, nil
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash password  `json:"-"`
	Bio          string    `json:"bio"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AnonymousUser is a placeholder for unauthenticated users.
var AnonymousUser = &User{}

// IsAnonymous checks if the user is anonymous.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{db: db}
}

// Interface for UserStore to allow decoupling and easier testing:
type UserStore interface {
	CreateUser(*User) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(*User) error
	GetUserToken(scope, tokenPlaintext string) (*User, error)
}

// CRU operations:

// Create user:
func (s *PostgresUserStore) CreateUser(user *User) (*User, error) {

	query := `
		INSERT INTO users (username, email, password_hash, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`
	err := s.db.QueryRow(query, user.Username, user.Email, user.PasswordHash.hash, user.Bio).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Read (Get) user by username:
func (s *PostgresUserStore) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, bio, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	user := &User{
		PasswordHash: password{},
	}
	err := s.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.hash,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // No user found
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

// Update user:
func (s *PostgresUserStore) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, bio = $3, updated_at = NOW()
		WHERE id = $4
	`
	result, err := s.db.Exec(query, user.Username, user.Email, user.Bio, user.ID)
	if err != nil {
		return err
	}
	// rows:
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Check if any row was actually updated
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *PostgresUserStore) GetUserToken(scope, plaintextPassword string) (*User, error) {
	// Implementation for retrieving a user by token from PostgreSQL

	tokenHash := sha256.Sum256([]byte(plaintextPassword))

	// INNER JOIN tokens t ON u.id = t.user_id (Not sure if order matters here)
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.bio, u.created_at, u.updated_at
		FROM users u
		INNER JOIN tokens t ON t.user_id = u.id
		WHERE t.hash = $1 AND t.scope = $2 AND t.expiry > $3
	`
	// the t.expiry is to ensure the token is still valid (not expired)
	user := &User{
		PasswordHash: password{},
	}
	err := s.db.QueryRow(query, tokenHash[:], scope, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash.hash,
		&user.Bio,
		&user.CreatedAt,
		&user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // No user found
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
