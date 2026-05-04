package api

import (
	"apiProject/internal/store"
	"apiProject/internal/tokens"
	"apiProject/internal/utils"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type TokenHandler struct {
	tokenStore store.TokenStore
	userStore  store.UserStore
	logger     *log.Logger
}

type createTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *log.Logger) *TokenHandler {
	return &TokenHandler{
		tokenStore: tokenStore,
		userStore:  userStore,
		logger:     logger,
	}
}

func (h *TokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("Error decoding: createTokenRequest: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}
	user, err := h.userStore.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		h.logger.Printf("Error GetUserByUsername: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	match, err := user.PasswordHash.Matches(req.Password)
	if err != nil {
		h.logger.Printf("Error PasswordHash.Matches: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	if !match {
		h.logger.Printf("Error password does not match: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid credentials"})
		return
	}

	token, err := h.tokenStore.CreateNewToken(user.ID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		h.logger.Printf("Error tokenStore.CreateNewToken: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	utils.WriteJson(w, http.StatusCreated, utils.Envelope{"auth_token": token})
}
