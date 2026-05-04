package routes

import (
	"apiProject/internal/app"

	"github.com/go-chi/chi/v5"
)

func SetupRoutes(app *app.Application) *chi.Mux {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware.Authenticate)
		r.Get("/workouts/{id}", app.AuthMiddleware.RequireUser(app.WorkoutHandler.HandleGetWorkoutById))
		r.Post("/workouts", app.AuthMiddleware.RequireUser(app.WorkoutHandler.HandleCreateWorkout))
		r.Put("/workouts/{id}", app.AuthMiddleware.RequireUser(app.WorkoutHandler.HandleUpdateWorkoutById))
		r.Delete("/workouts/{id}", app.AuthMiddleware.RequireUser(app.WorkoutHandler.HandleDeleteWorkout))
	})

	r.Get("/health", app.HealthCheck)

	r.Post("/users", app.UserHandler.HandleCreateUser)
	r.Post("/tokens/auth", app.TokenHandler.HandleCreateToken)
	return r
}
