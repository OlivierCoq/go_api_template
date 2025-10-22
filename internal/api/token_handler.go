package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/OlivierCoq/go_api_template/internal/middleware"
	"github.com/OlivierCoq/go_api_template/internal/store"
	"github.com/OlivierCoq/go_api_template/internal/tokens"
	"github.com/OlivierCoq/go_api_template/internal/utils"
)

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *log.Logger
}

// Used for decoding create token requests
type createTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewTokenHandler creates a new instance of TokenHandler

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *log.Logger) *TokenHandler {
	return &TokenHandler{
		tokenStore: tokenStore,
		userStore:  userStore,
		logger:     logger,
	}
}

func (h *TokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	// Implementation for creating a new token
	var req createTokenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("Error decoding create token request: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}

	user, err := h.userStore.GetUserByUsername(req.Username)
	if err != nil {
		h.logger.Printf("Error fetching user: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	passwordsDoMatch, err := user.PasswordHash.Matches(req.Password)
	if err != nil || !passwordsDoMatch {
		h.logger.Printf("Invalid credentials for user %s", req.Username)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid credentials"})
		return
	}
	if !passwordsDoMatch {
		h.logger.Printf("Invalid credentials for user %s", req.Username)
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid credentials"})
		return
	}

	token, err := h.tokenStore.CreateNewToken(user.ID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		h.logger.Printf("Error creating token: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"auth_token": token.Plaintext})
}

// Logging out:
func (h *TokenHandler) HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	// Implementation for revoking a token (logging out)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Missing Authorization header"})
		return
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid Authorization header format"})
		return
	}

	token := headerParts[1]
	err := h.tokenStore.RevokeToken(token)
	if err != nil {
		h.logger.Printf("Error revoking token: %v", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal Server Error"})
		return
	}

	// remove Authorization header from response and future http context:
	w.Header().Del("Authorization")
	// Middleware will set the user to anonymous user.
	r = middleware.SetUser(r, store.AnonymousUser)

	utils.WriteJSON(w, http.StatusOK, utils.Envelope{"message": "Token revoked successfully"})
}
