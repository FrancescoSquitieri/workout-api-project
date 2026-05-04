package api

import (
	"apiProject/internal/store"
	"apiProject/internal/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type registerUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
}

type UserHandler struct {
	userStore store.UserStore
	logger    *log.Logger
}

func NewUserHandler(userStore store.UserStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore: userStore,
		logger:    logger,
	}
}

func (h *UserHandler) validateRegisterRequest(req *registerUserRequest) error {
	if req.Username == "" {
		return errors.New("Username is required")
	}
	if req.Email == "" {
		return errors.New("Email is required")
	}
	if req.Password == "" {
		return errors.New("Email is required")
	}
	return nil
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("Decoding: register request: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request payload"})
		return
	}
	err = h.validateRegisterRequest(&req)
	if err != nil {
		h.logger.Printf("validateRegisterRequest: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}
	user := &store.User{
		Username: req.Username,
		Email:    req.Email,
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	err = user.PasswordHash.Set(req.Password)
	if err != nil {
		h.logger.Printf("PasswordHash.Set: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	err = h.userStore.CreateUser(user)
	if err != nil {
		h.logger.Printf("CreateUser: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	utils.WriteJson(w, http.StatusCreated, utils.Envelope{"user": user})
}
