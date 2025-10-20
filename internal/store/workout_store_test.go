package store

import (
	"database/sql"
	"testing" // provides testing framework

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost port=5433 user=postgres password=postgres dbname=postgres sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	// return db
	err = Migrate(db, "../../migrations")
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	/*
		- The purpose of truncating the tables before each test is to ensure that the database is in a clean state.
		- This prevents data from previous tests from interfering with the current test, which could lead to false positives or negatives.
		- By truncating the tables, we ensure that each test starts with an empty database, allowing for accurate and reliable test results.
		- The CASCADE option is used to automatically remove any dependent records in related tables, maintaining referential integrity.
		- This is especially important in a relational database where foreign key constraints may exist between tables.
		- Overall, truncating the tables helps maintain isolation between tests and ensures that each test is executed in a controlled environment.
	*/
	_, err = db.Exec(`TRUNCATE workouts, workout_entries CASCADE`)
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	return db
}

/*
	This file has the naming convention *_test.go which tells the Go tool that this file contains test functions.
	Each test function starts with the word Test followed by the name of the function being tested.
	Each test function takes a pointer to testing.T as a parameter, which provides methods for reporting and logging errors during the test.
*/

// Tests in Go are named starting with "Test" and take a pointer to testing.T as a parameter:
func TestCreateWorkout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Your test code here
	store := NewPostgresWorkoutStore(db)

	// Define test cases
	tests := []struct {
		name    string
		workout *Workout
		wantErr bool
	}{
		{
			name: "Valid workout",
			workout: &Workout{
				UserID:          1,
				Title:           "Morning Routine",
				Description:     "A quick morning workout",
				DurationMinutes: 30,
				CaloriesBurned:  250,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Push-ups",
						Sets:         3,
						Reps:         ptrInt(15), // These are pointers to integers because in Go, nil can be used to represent the absence of a value.
						OrderIndex:   1,
					},
					{
						ExerciseName: "Bench press",
						Sets:         3,
						Reps:         ptrInt(10),
						Weight:       FloatPtr(135.5),
						Notes:        "warm up properly",
						OrderIndex:   1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid workout entry",
			workout: &Workout{
				UserID:          1,
				Title:           "Evening Routine",
				Description:     "A quick evening workout",
				DurationMinutes: 20,
				CaloriesBurned:  200,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Squats",
						Sets:         3,
						Reps:         nil, // Invalid because both Reps and DurationSeconds are nil
						OrderIndex:   1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Reps and Duration in seconds simultaneously",
			workout: &Workout{
				UserID:          1,
				Title:           "Afternoon Routine",
				Description:     "A quick afternoon workout",
				DurationMinutes: 25,
				CaloriesBurned:  220,
				Entries: []WorkoutEntry{
					{
						ExerciseName:    "Jumping Jacks",
						Sets:            2,
						Reps:            ptrInt(20),
						DurationSeconds: ptrInt(300), // Invalid because both Reps and DurationSeconds are provided
						OrderIndex:      1,
					},
				},
			},
			wantErr: true,
		},
	}

	// Loop through test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdWorkout, err := store.CreateWorkout(tt.workout)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.workout.Title, createdWorkout.Title)
			assert.Equal(t, tt.workout.Description, createdWorkout.Description)
			assert.Equal(t, tt.workout.DurationMinutes, createdWorkout.DurationMinutes)
			assert.Equal(t, tt.workout.CaloriesBurned, createdWorkout.CaloriesBurned)
			assert.Equal(t, len(tt.workout.Entries), len(createdWorkout.Entries))

			// Verify entries

			retrieved, err := store.GetWorkoutByID(int64(createdWorkout.ID))
			require.NoError(t, err)
			assert.Equal(t, createdWorkout.ID, retrieved.ID)
			assert.Equal(t, len(tt.workout.Entries), len(retrieved.Entries))

			for i := range retrieved.Entries {
				assert.Equal(t, tt.workout.Entries[i].ExerciseName, retrieved.Entries[i].ExerciseName)
				assert.Equal(t, tt.workout.Entries[i].Sets, retrieved.Entries[i].Sets)
				assert.Equal(t, tt.workout.Entries[i].Reps, retrieved.Entries[i].Reps)
				assert.Equal(t, tt.workout.Entries[i].DurationSeconds, retrieved.Entries[i].DurationSeconds)
				assert.Equal(t, tt.workout.Entries[i].OrderIndex, retrieved.Entries[i].OrderIndex)
			}

		})
	}
}

// Helper function for getting a pointer to an int, since Go doesn't have built-in syntax for that
// Extracts the address of an integer variable
func ptrInt(i int) *int {
	return &i // pointer to i.
}

func FloatPtr(f float64) *float64 {
	return &f
}
