package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/OlivierCoq/go_api_project/internal/store"
	"github.com/OlivierCoq/go_api_project/internal/tokens"
	"github.com/OlivierCoq/go_api_project/internal/utils"
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
