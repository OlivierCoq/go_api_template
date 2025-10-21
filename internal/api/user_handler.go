package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"github.com/OlivierCoq/go_api_project/internal/store"
	"github.com/OlivierCoq/go_api_project/internal/utils"
)

// This is for user registration requests
type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

type UserHandler struct {
	// Add fields as necessary, e.g., a reference to the application or database
	userStore store.UserStore // Interface to interact with user data. This promotes db decoupling and easier testing.
	logger    *log.Logger
}

// NewUserHandler creates a new instance of UserHandler
func NewUserHandler(userStore store.UserStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    logger,
	}
}

// Validation:
func (h *UserHandler) validateRegisterUserRequest(req *RegisterUserRequest) error {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return errors.New("username, email, and password are required")
	}
	if len(req.Username) > 50 {
		return errors.New("username must be less than 50 characters")
	}
	// Email
	if len(req.Email) > 100 {
		return errors.New("email must be less than 100 characters")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}
	// Password
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(req.Password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(req.Password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(req.Password)
	if !(hasLower && hasUpper && hasDigit) {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, and one number")
	}

	return nil
}

// Define methods for UserHandler to handle user-related requests. CRUD operations, etc.

// Create/Register User
func (h *UserHandler) HandleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var req RegisterUserRequest
	// Decode the POST request body into the RegisterUserRequest struct:
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("Invalid request payload: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"}) // 400
		return
	}

	// Validate the request
	err = h.validateRegisterUserRequest(&req)
	if err != nil {
		h.logger.Printf("Validation error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()}) // 400
		return
	}

	// Send to DB via userStore
	user := &store.User{
		Username: req.Username,
		Email:    req.Email,
	}

	if req.Bio != "" {
		user.Bio = req.Bio
	}

	err = user.PasswordHash.Set(req.Password)
	if err != nil {
		h.logger.Printf("Error setting password hash: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"}) // 500
		return
	}

	createdUser, err := h.userStore.CreateUser(user)
	if err != nil {
		h.logger.Printf("Error creating user: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create user"}) // 500
		return
	}

	// Respond with the created user (excluding password hash) as JSON to the frontend:
	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"user": createdUser}) // 201

}
