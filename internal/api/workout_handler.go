package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/OlivierCoq/go_api_project/internal/store"
	"github.com/go-chi/chi/v5"
)

type WorkoutHandler struct {
	// Add fields as necessary, e.g., a reference to the application or database
	workoutStore store.WorkoutStore // Interface to interact with workout data. This promotes db decoupling and easier testing.
}

// NewWorkoutHandler creates a new instance of WorkoutHandler
func NewWorkoutHandler(workoutStore store.WorkoutStore) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore: workoutStore,
	}
}

// Define methods for WorkoutHandler to handle workout-related requests. CRUD operations, etc.

func (wh *WorkoutHandler) HandleGetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	// Implementation for getting a workout by ID
	paramsWorkoutID := chi.URLParam(r, "id")
	if paramsWorkoutID == "" {
		http.Error(w, "Workout ID is required", http.StatusBadRequest)
		return
	}

	workoutID, err := strconv.ParseInt(paramsWorkoutID, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// fmt.Fprintf(w, "Workout ID: %d\n", workoutID)
	workout, err := wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		http.Error(w, "Workout not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workout)
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {

	var workout store.Workout
	// Decode the POST request body into the Workout struct (from the front end):
	// Note: remember, the & is used to get the memory address of the variable, so we can modify its value directly.
	err := json.NewDecoder(r.Body).Decode(&workout) // For clarity, see struct in store/workout_store.go
	if err != nil {

		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
		return
	}

	// Feedback from the store
	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	// Respond with the created workout as JSON to the frontend:
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdWorkout)
}
