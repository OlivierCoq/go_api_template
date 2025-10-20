package store

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost port=5433 user=postgres password=postgres dbname=test_db sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	// return db
	err = Migrate(db, "../../migrations")
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	_, err = db.Exec(`TRUNCATE workouts, workout_entries CASCADE`)
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	return db
}
