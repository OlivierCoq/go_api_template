package store

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver, sql package uses it via side-effects
)

func Open() (*sql.DB, error) {
	db, err := sql.Open("pgx", "host=localhost port=5432 user=postgres password=postgress dbname=workout_db sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	fmt.Println("Database connection established! 🚀")
	// db.SetMaxOpenConns(), db.SetMaxIdleConns(), and db.SetConnMaxIdleTime()
	return db, nil
}

//db.SetMaxOpenConns(), db.SetMaxIdleConns(), and db.SetConnMaxIdleTime() can be used to fine-tune the connection pool settings based on your application's needs.
// These settings help manage the number of open connections, idle connections, and the duration for which a connection can remain idle before being closed.
// Properly configuring these settings can improve performance and resource utilization, especially under varying workloads.
