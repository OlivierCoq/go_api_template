package routes

import (
	"github.com/OlivierCoq/go_api_project/internal/app"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()

	// Define routes and their handlers here
	r.Get("/health", app.HealthCheck) // Health check endpoint
	r.Get("/workouts/{id}", app.WorkoutHandler.HandleGetWorkoutByID)
	r.Post("/workouts", app.WorkoutHandler.HandleCreateWorkout)
	r.Patch("/workouts/{id}", app.WorkoutHandler.HandleUpdateWorkout)
	// Add more routes as needed

	return r
}
