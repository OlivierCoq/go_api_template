package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/OlivierCoq/go_api_project/internal/app"
)

/*
Note: in Go, constructors typically return a pointer to a struct and an error.
- If the struct is created successfully, the error is nil.
- If there is an issue during creation, the struct pointer is nil and the error contains the details.
- The point of using a structure is to group related data and methods together, making the code more organized and easier to manage.
*/
func main() {

	var port int // port number for the server to listen on
	/*
	  - the flag package in Go provides a way to define and parse command-line flags.
	  - In this case, we are defining an integer flag named "port" with a default value of 8080 and a description "Port to run the server on".
	  - The &port is a pointer to the variable where the parsed value will be stored.
	*/
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
	flag.Parse() // Parse command-line flags. Does heavy lifting of parsing flags

	// Initialize the application (taken from internal/app/app.go):
	app, err := app.NewApplication()
	if err != nil {
		// Worst case scenario, we panic here with the error. Will crash the app
		panic(err)
	}

	app.Logger.Println("Application started. Werk it! 🚀")

	// Set up routes and handlers

	/*
		Here, we would set up our routes and handlers e.g., http.HandleFunc("/some-endpoint", someHandlerFunction)
		Routes have 2 arguments: path (where the function is), and the handler function itself:
			(In Go, functions are first-class citizens, meaning they can be passed around as values
			- This allows us to pass the function name without parentheses, which would call the function immediately)
	*/
	// Here, we are passing the HealthCheck function as a value to the HandleFunc method (Methods are defined below)
	http.HandleFunc("/health", HealthCheck)

	// Set up server:

	// declare a new server with specific configurations
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port), // returns variable port as a string with a colon in front of it
		IdleTimeout:  time.Minute,              // how long to wait before closing idle connections
		ReadTimeout:  10 * time.Second,         // max duration for reading the entire request, including the body
		WriteTimeout: 30 * time.Second,         // max duration before timing out writes of the response
	}
	app.Logger.Printf("Starting server on port %d\n", port)

	// Start the server
	err = server.ListenAndServe()
	// Wait for crashes or shutdown. Always fail first.
	if err != nil {
		app.Logger.Println(err)
	}
	app.Logger.Println("Application stopped. Bye! 👋")
}

// Methods

// Health Check Handler
func HealthCheck(w http.ResponseWriter, r *http.Request) {
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

/*

	Notes:
	- In Go, we typically use the standard library's log package for logging.
		// To get env files in Go, we can use the os package and Getenv function
	portStr := os.Getenv("PORT") // Get the value of the PORT environment variable
	if portStr != "" {
		port, err = strconv.Atoi(portStr) // Convert the string to an integer
		if err != nil {
			log.Fatalf("Invalid port number: %v", err)
		}
	}


*/
