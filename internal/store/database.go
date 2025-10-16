package store

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

func Open() (*sql.DB, error) {
	db, err := sql.Open("pgx", "host=localhost port=5432 user=postgres password=postgress dbname=workout_db sslmode=disable")

	return db, err
}
