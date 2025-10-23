# Go project template

A pretty nifty backend structure for a scalable web app.

This project uses:
 
 - Go
 - Docker
 - Goose (for database migrations)
 - Chi (for route handling)
 - Postgres


## How it works/Project structure

Basically this project splits concerns into packages that are then imported and compiled, like most web apps. The project is structured according to separation of concern, allowing for a more organized and scalable design. 

There are several layers and sections of concern, including the database, API, router, and application.

### The Database

Found in the `store` folder, this is where the database connections and SQL queries are run. The application does not interact with the databse directly, but instead with the *interface* found here, which is a struct containing each function. 

This way, if there are new databases used, the application logic remains the same. 

#### Example interface:

```
type WorkoutStore interface {
	CreateWorkout(*Workout) (*Workout, error)
	GetWorkoutByID(id int64) (*Workout, error)
	UpdateWorkout(*Workout) error
	DeleteWorkout(id int64) error
	GetWorkoutOwner(id int64) (int, error)
}
```

### The API 

The API layer consists of handlers, which call the *interface* found in the previously mentioned database layer. The handlers accept http requests, parse said requests, and send the data to and from the database layer. 

In the API layer, a struct is made consisting of a *logger* and a *store* defined the previous database layer. This struct is then appended with several functions, usually the corresponding CRUD functions for each function in the database *interface*. 

#### Example Request:

```
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
```

### The Router

The router does just that, route our application. Each endpoint is defined by a *path* and a *handler* function that corresponds with said path. These handlers are found in the previously mentioned *API layer*. 

#### Example route:

```
r.Get("/workouts/{id}", app.Middleware.RequireUser(app.WorkoutHandler.HandleGetWorkoutByID))
```

As one would imagine, each guard may or may not be guarded by middleware:

#### Middleware

The middleware folder includes functions that handle authorization levels, logged in status, and route protection. The way it works:

- Grab user (authenticated or anonymous) and set into http.Request context using the `SetUser` function. This alters every single request to contain a user in its context, and for the entire application to panic and shut down if any request does not contain a user. This stops any bad actors from accessing any guarded endpoints.
- Using the user previously set in the `setUser` function, any other function will first check each request and grab the user from context using the `getUser` function.
- Authenticate. When a user logs in, they hit the `tokens/authentication` route, which checks for their credentials, runs several security checks, and returns with a token which is attached to the request headers. This is then attached to any routes that require a user via `RequireUser`. This then wraps around any protected routes. (Redirection is handled via the front end)

### The Application

The previous layers are then encapsulated in the `app.go` file, which is the heart of the application. This consists of a struct that pulls in the *stores* (Database layer), *handlers* (API layer), *middleware*, and *router* under one struct, and then instantiates it:

```
	// Create a new instance of Application struct, which includes the logger, handlers, etc.:
app := &Application { // &Application is pointer to Application struct
		Logger:         logger,
		WorkoutHandler: workoutHandler,
		DB:             pgDB, // Add the database connection to the Application struct
		TokenHandler:   tokenHandler,
		UserHandler:    userHandler,
		Middleware:     userMiddleware,
	}
	return app, nil // nil is for the error argument, meaning no error occurred :)
}

```


## Getting it up and running

When first starting a Go project from scratch, in your folder, you'd run `go mod init <module_path>` .

In this project's case:

### Get Docker up and running (from root):

```
docker compose up --build
```

### Then, in another terminal (also from root):

```
go run main.go
```