package api

import (
	"apiProject/internal/middleware"
	"apiProject/internal/store"
	"apiProject/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type WorkoutHandler struct {
	workoutStore store.WorkoutStore
	logger       *log.Logger
}

func NewWorkoutHandler(workoutStore store.WorkoutStore, logger *log.Logger) *WorkoutHandler {
	return &WorkoutHandler{
		workoutStore,
		logger,
	}
}

func (wh *WorkoutHandler) HandleGetWorkoutById(w http.ResponseWriter, r *http.Request) {
	workoutId, err := utils.ReadIdParameter(r)
	if err != nil {
		wh.logger.Printf("Error: read id param: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid workout id"})
		return
	}

	workout, err := wh.workoutStore.GetWorkoutById(workoutId)
	if err != nil {
		wh.logger.Printf("Error: GetWorkoutById: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}
	utils.WriteJson(w, http.StatusOK, utils.Envelope{"workout": workout})
}

func (wh *WorkoutHandler) HandleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var workout store.Workout
	err := json.NewDecoder(r.Body).Decode(&workout)
	if err != nil {
		wh.logger.Printf("Error: decoding failed HandleCreateWorkout: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request sent"})
		return
	}

	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser.IsAnonymous() {
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Ypu must be logged in"})
		return
	}
	workout.UserID = currentUser.ID

	createdWorkout, err := wh.workoutStore.CreateWorkout(&workout)
	if err != nil {
		wh.logger.Printf("Error: CreateWorkout: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "Failed to create workout"})
		return
	}
	utils.WriteJson(w, http.StatusCreated, utils.Envelope{"workout": createdWorkout})
}

func (wh *WorkoutHandler) HandleUpdateWorkoutById(w http.ResponseWriter, r *http.Request) {
	workoutId, err := utils.ReadIdParameter(r)
	if err != nil {
		wh.logger.Printf("Error: read id param HandleUpdateWorkoutById: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid workout update id"})
		return
	}

	existingWorkout, err := wh.workoutStore.GetWorkoutById(workoutId)
	if err != nil {
		wh.logger.Printf("Error: GetWorkoutById: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}
	if existingWorkout == nil {
		http.NotFound(w, r)
		return
	}
	var updateWorkoutRequest struct {
		Title           *string              `json:"title"`
		Description     *string              `json:"description"`
		DurationMinutes *int                 `json:"duration_minutes"`
		CaloriesBurned  *int                 `json:"calories_burned"`
		Entries         []store.WorkoutEntry `json:"entries"`
	}
	err = json.NewDecoder(r.Body).Decode(&updateWorkoutRequest)
	if err != nil {
		wh.logger.Printf("Error: Decoding update request: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid request payload"})
		return
	}
	if updateWorkoutRequest.Title != nil {
		existingWorkout.Title = *updateWorkoutRequest.Title
	}
	if updateWorkoutRequest.Description != nil {
		existingWorkout.Description = *updateWorkoutRequest.Description
	}
	if updateWorkoutRequest.DurationMinutes != nil {
		existingWorkout.DurationMinutes = *updateWorkoutRequest.DurationMinutes
	}
	if updateWorkoutRequest.CaloriesBurned != nil {
		existingWorkout.CaloriesBurned = *updateWorkoutRequest.CaloriesBurned
	}
	if updateWorkoutRequest.Entries != nil {
		existingWorkout.Entries = updateWorkoutRequest.Entries
	}
	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser.IsAnonymous() {
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "You must be logged in"})
		return
	}
	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, utils.Envelope{"error": "You must be logged in"})
			return
		}
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}
	if workoutOwner != currentUser.ID {
		utils.WriteJson(w, http.StatusForbidden, utils.Envelope{"error": "You're not authorized to update this workout"})
		return
	}
	err = wh.workoutStore.UpdateWorkout(existingWorkout)
	if err != nil {
		wh.logger.Printf("Error: UpdateWorkout: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}
	utils.WriteJson(w, http.StatusOK, utils.Envelope{"workout": existingWorkout})
}

func (wh *WorkoutHandler) HandleDeleteWorkout(w http.ResponseWriter, r *http.Request) {
	workoutId, err := utils.ReadIdParameter(r)
	if err != nil {
		wh.logger.Printf("Error: read id param HandleDeleteWorkout: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Invalid workout delete id"})
		return
	}
	currentUser := middleware.GetUser(r)
	if currentUser == nil || currentUser.IsAnonymous() {
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "You must be logged in"})
		return
	}
	workoutOwner, err := wh.workoutStore.GetWorkoutOwner(workoutId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteJson(w, http.StatusNotFound, utils.Envelope{"error": "You must be logged in"})
			return
		}
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{"error": "Internal server error"})
		return
	}
	if workoutOwner != currentUser.ID {
		utils.WriteJson(w, http.StatusForbidden, utils.Envelope{"error": "You're not authorized to update this workout"})
		return
	}
	err = wh.workoutStore.DeleteWorkout(workoutId)
	if err == sql.ErrNoRows {
		wh.logger.Printf("Error: workoutStore.DeleteWorkout: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Internal server error"})
		return
	}
	if err != nil {
		wh.logger.Printf("Error: workoutStore.DeleteWorkout: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{"error": "Internal server error"})
		return
	}
	utils.WriteJson(w, http.StatusNoContent, utils.Envelope{"success": true, "message": "Workout deleted"})
}
