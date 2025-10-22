package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/OlivierCoq/go_api_template/internal/middleware"
	"github.com/OlivierCoq/go_api_template/internal/store"
	"github.com/OlivierCoq/go_api_template/internal/utils"
)

type WorkoutHandler struct {
	// Add fields as necessary, e.g., a reference to the application or database
	workoutStore store.WorkoutStore // Interface to interact with workout data. This promotes db decoupling and easier testing.
	logger       *log.Logger
}

// NewWorkoutHandler creates a new instance of WorkoutHandler
func NewWorkoutHandler(workoutStore store.WorkoutStore, logger *log.Logger) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore: workoutStore,
		logger:       logger,
	}
}

// Define methods for WorkoutHandler to handle workout-related requests. CRUD operations, etc.

// Create
func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {

	var workout store.Workout
	// Decode the POST request body into the Workout struct (from the front end):
	// Note: remember, the & is used to get the memory address of the variable, so we can modify its value directly.
	err := json.NewDecoder(r.Body).Decode(&workout) // For clarity, see struct in store/workout_store.go
	if err != nil {

		http.Error(w, "Invalid request payload", http.StatusInternalServerError)
		return
	}

	// Ensure that this is being created by an authenticated user
	currentUser := middleware.GetUser(r)
	if currentUser.IsAnonymous() {
		http.Error(w, "Authentication required to create workout", http.StatusUnauthorized)
		return
	}

	workout.UserID = currentUser.ID // Associate the workout with the current user's ID

	// Feedback from the store
	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		http.Error(w, "Failed to create workout", http.StatusInternalServerError)
		return
	}

	// Respond with the created workout as JSON to the frontend:
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(createdWorkout)
	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"workout": createdWorkout}) // 201
}

// Read
func (wh *WorkoutHandler) HandleGetWorkoutByID(w http.ResponseWriter, r *http.Request) {
	// Implementation for getting a workout by ID
	workoutID, err := utils.ReadIDParam(r, "id")
	if err != nil {
		wh.logger.Printf("Error getWorkoutByID : %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Invalid workout ID"}) // 500
		return
	}

	// fmt.Fprintf(w, "Workout ID: %d\n", workoutID)
	workout, err := wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		wh.logger.Printf("Workout not found: %v", err)
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Workout not found"})
		return
	}
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(workout)
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": workout}) // 200
}

// Update
func (wh *WorkoutHandler) HandleUpdateWorkout(w http.ResponseWriter, r *http.Request) {
	// Implementation for updating a workout
	paramsWorkoutID, err := utils.ReadIDParam(r, "id")
	if err != nil {
		wh.logger.Printf("Error reading workout ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Workout ID is required"}) // 400
		return
	}

	// workoutID, err := strconv.ParseInt(paramsWorkoutID, 10, 64)
	// if err != nil {
	// 	http.NotFound(w, r)
	// 	return
	// }

	// Fetch existing workout from DB to ensure it exists and get current data
	workout, err := wh.workoutStore.GetWorkoutByID(paramsWorkoutID)
	if err != nil {
		// http.Error(w, "Workout not found", http.StatusNotFound)
		wh.logger.Printf("Workout not found: %v", err)
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Workout not found"}) // 404
		return
	}

	// We use json tags here for parsing purposes. We use pointers to differentiate between zero values and missing fields.
	var updateWorkoutRequest struct {
		Title           *string               `json:"title"`
		Description     *string               `json:"description"`
		DurationMinutes *int                  `json:"duration"`
		CaloriesBurned  *int                  `json:"calories_burned"`
		Entries         *[]store.WorkoutEntry `json:"entries"`
	}

	err = json.NewDecoder(r.Body).Decode(&updateWorkoutRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	/*
			- What's happening here is that we're checking if each field in the updateWorkoutRequest struct is non-nil (meaning it was provided in the request).
		- If it's non-nil, we dereference the pointer to get the actual value and update the corresponding field in the workout struct.
		- This way, only the fields that were provided in the request will be updated, while others will remain unchanged.
	*/
	if updateWorkoutRequest.Title != nil {
		workout.Title = *updateWorkoutRequest.Title
	}
	if updateWorkoutRequest.Description != nil {
		workout.Description = *updateWorkoutRequest.Description
	}
	if updateWorkoutRequest.DurationMinutes != nil {
		workout.DurationMinutes = *updateWorkoutRequest.DurationMinutes
	}
	if updateWorkoutRequest.CaloriesBurned != nil {
		workout.CaloriesBurned = *updateWorkoutRequest.CaloriesBurned
	}
	if updateWorkoutRequest.Entries != nil {
		workout.Entries = *updateWorkoutRequest.Entries
	}

	// Associate the workout with the current user's ID
	currentUser := middleware.GetUser(r)
	if currentUser.IsAnonymous() {
		http.Error(w, "Authentication required to update workout", http.StatusUnauthorized)
		return
	}
	workout.UserID = currentUser.ID

	// Ensure that the current user is the owner of the workout
	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(paramsWorkoutID)
	if err != nil {
		wh.logger.Printf("Error retrieving workout owner: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve workout owner"})
		return
	}
	if workoutOwner != currentUser.ID {
		http.Error(w, "You do not have permission to update this workout", http.StatusForbidden)
		return
	}

	// Update workout in the store
	err = wh.workoutStore.UpdateWorkout(workout)
	if err != nil {
		wh.logger.Printf("Failed to update workout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to update workout"})
		return
	}

	// Respond with entire updated workout as JSON to the frontend:
	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"workout": workout}) // 200
}

// Delete
func (wh *WorkoutHandler) HandleDeleteWorkout(w http.ResponseWriter, r *http.Request) {
	// Implementation for deleting a workout

	workoutID, err := utils.ReadIDParam(r, "id")
	if err != nil {
		wh.logger.Printf("Error reading workout ID: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid workout ID"}) // 400
		return
	}
	// Check if workout exists
	_, err = wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		wh.logger.Printf("Workout not found: %v", err)
		utils.WriteJSON(w, http.StatusNotFound, utils.Envelope{"error": "Workout not found"}) // 404
		return
	}

	// Ensure that the current user is the owner of the workout before deletion:
	currentUser := middleware.GetUser(r)
	if currentUser.IsAnonymous() {
		http.Error(w, "Authentication required to delete workout", http.StatusUnauthorized)
		return
	}

	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutID)
	if err != nil {
		wh.logger.Printf("Error retrieving workout owner: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve workout owner"})
		return
	}
	if workoutOwner != currentUser.ID {
		wh.logger.Printf("User %d attempted to delete workout %d owned by user %d", currentUser.ID, workoutID, workoutOwner)
		utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "You do not have permission to delete this workout"})
		return
	}
	// Ensure that the workout exists before attempting deletion
	_, err = wh.workoutStore.GetWorkoutByID(workoutID)
	if err != nil {
		wh.logger.Printf("Error retrieving workout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to retrieve workout"})
		return
	}

	// Delete workout
	err = wh.workoutStore.DeleteWorkout(workoutID)
	if err != nil {
		wh.logger.Printf("Failed to delete workout: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to delete workout"}) // 500
		return
	}

}
