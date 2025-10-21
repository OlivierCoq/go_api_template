package app

// housing = data
// business logic = service
// interface = handler

import (
	"database/sql"
	"fmt"      // for formatted I/O operations
	"log"      // for logging messages
	"net/http" // for building HTTP servers and clients
	"os"       // for logging to standard output (console)

	"github.com/OlivierCoq/go_api_project/internal/api"   // Importing the api package to use its handlers
	"github.com/OlivierCoq/go_api_project/internal/store" // Importing the store package for database access
	"github.com/OlivierCoq/go_api_project/migrations"
)

type Application struct {
	// Logger is for logging messages to the console or a file
	Logger         *log.Logger
	WorkoutHandler *api.WorkoutHandler
	UserHandler    *api.UserHandler
	TokenHandler   *api.TokenHandler
	DB             *sql.DB // Add the database connection field
}

func NewApplication() (*Application, error) {

	// Create a new logger instance:
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	/*
		- os.Stdout = output destination for the log messages. Standard output (console)
		- "" = prefix for each log message (empty string means no prefix)
		- log.Ldate = include date in log messages
		- log.Ltime = include time in log messages
		- "" = prefix for each log message (empty string means no prefix)

		Example log message:
		2024/10/05 14:23:45 Application started. Werk it! 🚀
	*/

	// Database connection
	pgDB, err := store.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	// Stores
	workoutStore := store.NewPostgresWorkoutStore(pgDB)
	userStore := store.NewPostgresUserStore(pgDB)
	tokenStore := store.NewPostgresTokenStore(pgDB)

	// Handlers
	workoutHandler := api.NewWorkoutHandler(workoutStore, logger)
	userHandler := api.NewUserHandler(userStore, logger)
	tokenHandler := api.NewTokenHandler(tokenStore, userStore, logger)

	// Run database migrations using the embedded filesystem:
	// the "." means the current directory, which is where the migration files are located in the embedded FS
	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		// panic and crash the app if migration fails:
		panic(err)
	}

	// Create a new instance of Application struct, which includes the logger, handlers, etc.:
	app := &Application{ // &Application is pointer to Application struct
		Logger:         logger,
		WorkoutHandler: workoutHandler,
		DB:             pgDB, // Add the database connection to the Application struct
		TokenHandler:   tokenHandler,
		UserHandler:    userHandler,
	}
	return app, nil // nil is for the error argument, meaning no error occurred :)
}

// Methods:

// Health Check Handler
// Moved from main.go to here, to be a method of Application struct. We're updating the Application struct to have a method called HealthCheck
func (a *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	/*
		- Purpose: To verify that the server is running and responsive.
		- Called by client (UI, some frontent)
		- needs 2 arguments: ResponseWriter and Request
		- ResponseWriter: used to send a response back to the client
		- Request: contains all the information about the incoming HTTP request. This is a pointer because it can be large and we want to avoid copying it. We
		also need it to persist and modify it, especially when dealing with middleware or request body.
		- In a real-world scenario, you might want to include more detailed health information,
		  such as database connectivity, external service status, etc.
		- In this example, we simply write "OK" to the response with a 200 status code.
	*/
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Status is available. A okay! 🟢\n")
}
