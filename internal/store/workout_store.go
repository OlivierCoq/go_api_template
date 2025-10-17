package store

import (
	"database/sql"
	"fmt"
)

type Workout struct {
	ID              int            `json:"id"`
	UserID          int            `json:"user_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	DurationMinutes int            `json:"duration"` // Duration in minutes
	CaloriesBurned  int            `json:"calories_burned"`
	Entries         []WorkoutEntry `json:"entries"`
}

type WorkoutEntry struct {
	ID              int      `json:"id"`
	ExerciseName    string   `json:"exercise_name"`
	Sets            int      `json:"sets"`
	Reps            *int     `json:"reps"`
	DurationSeconds *int     `json:"duration_seconds"`
	Weight          *float64 `json:"weight"`
	Notes           string   `json:"notes"`
	OrderIndex      int      `json:"order_index"`
}

type PostgresWorkoutStore struct {
	db *sql.DB
}

func NewPostgresWorkoutStore(db *sql.DB) *PostgresWorkoutStore {
	return &PostgresWorkoutStore{db: db}
}

/*
	In the unfortunate event that we have to change databases, we can use an interface to decouple the data access layer from the rest of the application.
	An interface is a collection of method signatures that defines a set of behaviors. When a type implements all the methods in an interface, it can be used as that interface type.
	This allows us to define a contract that any database implementation must adhere to, making it easier to swap out the underlying database without affecting the rest of the application.
*/

type WorkoutStore interface {
	CreateWorkout(*Workout) (*Workout, error)
	GetWorkoutByID(id int64) (*Workout, error)
	UpdateWorkout(*Workout) error
}

func (pg *PostgresWorkoutStore) CreateWorkout(workout *Workout) (*Workout, error) {

	/*
		This is a transaction.
		A transaction is a sequence of operations performed as a single logical unit of work.
		If any of the operations fail, the entire transaction is rolled back, ensuring data integrity. This avoids
		conflicts with ACID principles (Atomicity, Consistency, Isolation, Durability).
	*/

	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// The $ notation is used for parameterized queries in PostgreSQL.
	query := `INSERT INTO workouts (user_id, title, description, duration_minutes, calories_burned)
			  VALUES ($1, $2, $3, $4, $5)
			  RETURNING id`

	/*
		What's happening here:
		1. We prepare an SQL query to insert a new workout into the workouts table.
		2. We use the QueryRow method to execute the query with the provided workout details.
		3. The Scan method retrieves the generated ID of the newly created workout and assigns it to workout.ID.
	*/
	err = tx.QueryRow(query, workout.UserID, workout.Title, workout.Description, workout.DurationMinutes, workout.CaloriesBurned).Scan(&workout.ID)
	if err != nil {
		return nil, err
	}

	// Entries is a slice of WorkoutEntry structs within the Workout struct.
	// We iterate over each entry to insert them into the workout_entries table:

	for _, entry := range workout.Entries {
		entryQuery := `INSERT INTO workout_entries (user_id, workout_id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index)
					   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
					   RETURNING id`
		err = tx.QueryRow(entryQuery, workout.UserID, workout.ID, entry.ExerciseName, entry.Sets, entry.Reps, entry.DurationSeconds, entry.Weight, entry.Notes, entry.OrderIndex).Scan(&entry.ID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// Implementation for creating a workout in PostgreSQL
	return workout, nil
}

func (pg *PostgresWorkoutStore) GetWorkoutByID(id int64) (*Workout, error) {
	workout := &Workout{}

	query := `SELECT id, user_id, title, description, duration_minutes, calories_burned
			  FROM workouts
			  WHERE id = $1`

	/*
		- When scanning db query results, the Scan method must receive pointers to the destination variables.
	*/
	err := pg.db.QueryRow(query, id).Scan(&workout.ID, &workout.UserID, &workout.Title, &workout.Description, &workout.DurationMinutes, &workout.CaloriesBurned)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No workout found with the given ID
		}
		return nil, err
	}

	// Fetch workout entries
	entriesQuery := `SELECT id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index
					 FROM workout_entries
					 WHERE workout_id = $1
					 ORDER BY order_index ASC`

	// rows, because we can have multiple entries per workout:
	rows, err := pg.db.Query(entriesQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure rows are closed after we're done with them

	/*
		Iterate over the rows and scan each entry into a WorkoutEntry struct,
		then append it to the Entries slice of the Workout struct.
	*/
	for rows.Next() {
		var entry WorkoutEntry
		err = rows.Scan(
			&entry.ID,
			&entry.ExerciseName,
			&entry.Sets,
			&entry.Reps,
			&entry.DurationSeconds,
			&entry.Weight,
			&entry.Notes,
			&entry.OrderIndex,
		)
		if err != nil {
			return nil, err
		}
		workout.Entries = append(workout.Entries, entry)
	}

	return workout, nil
	// Implementation for retrieving a workout by ID from PostgreSQL
}

func (pg *PostgresWorkoutStore) UpdateWorkout(workout *Workout) error {

	// transaction
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE workouts
			  SET user_id = $1, title = $2, description = $3, duration_minutes = $4, calories_burned = $5
			  WHERE id = $6`

	res, err := tx.Exec(query, workout.UserID, workout.Title, workout.Description, workout.DurationMinutes, workout.CaloriesBurned, workout.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no workout found with id %d", workout.ID)
	}

	// Updating entries (moved before commit)
	for _, entry := range workout.Entries {
		entryQuery := `UPDATE workout_entries
					   SET exercise_name = $1, sets = $2, reps = $3, duration_seconds = $4, weight = $5, notes = $6, order_index = $7
					   WHERE id = $8 AND workout_id = $9`
		_, err = tx.Exec(entryQuery, entry.ExerciseName, entry.Sets, entry.Reps, entry.DurationSeconds, entry.Weight, entry.Notes, entry.OrderIndex, entry.ID, workout.ID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
