package store

import (
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDb(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "host=localhost user=postgres password=postgres dbname=workoutDB_test port=5433 sslmode=disable")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	err = Migrate(db, "../../migrations")
	if err != nil {
		t.Fatalf("migrating test db: %v", err)
	}
	_, err = db.Exec(`TRUNCATE workouts, workout_entries CASCADE`)
	if err != nil {
		t.Fatalf("truncating tables error: %v", err)
	}
	return db
}

func TestCreatewokrout(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	store := NewPostgresWorkoutStore(db)

	tests := []struct {
		name      string
		workout   *Workout
		wantError bool
	}{
		{
			name: "Valid workout",
			workout: &Workout{
				Title:           "push day",
				Description:     "upper body day",
				DurationMinutes: 60,
				CaloriesBurned:  200,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Bench press",
						Sets:         3,
						Reps:         IntPtr(10),
						Weight:       FloatPtr(135.5),
						Notes:        "warm up properly",
						OrderIndex:   1,
					},
				},
			},
			wantError: false,
		},
		{
			name: "workout with invalid entries",
			workout: &Workout{
				Title:           "full body",
				Description:     "complete workout",
				DurationMinutes: 90,
				CaloriesBurned:  500,
				Entries: []WorkoutEntry{
					{
						ExerciseName: "Plank",
						Sets:         3,
						Reps:         IntPtr(60),
						Notes:        "keep form",
						OrderIndex:   1,
					},
					{
						ExerciseName:    "Squats",
						Sets:            4,
						Reps:            IntPtr(12),
						DurationSeconds: IntPtr(60),
						Weight:          FloatPtr(185.0),
						Notes:           "Full depth",
						OrderIndex:      2,
					},
				},
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdWorkout, err := store.CreateWorkout(tt.workout)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.workout.Title, createdWorkout.Title)
			assert.Equal(t, tt.workout.Description, createdWorkout.Description)
			assert.Equal(t, tt.workout.DurationMinutes, createdWorkout.DurationMinutes)
			assert.Equal(t, tt.workout.CaloriesBurned, createdWorkout.CaloriesBurned)
			assert.Equal(t, tt.workout.ID, createdWorkout.ID)
			assert.Equal(t, len(tt.workout.Entries), len(createdWorkout.Entries))

			retrieved, err := store.GetWorkoutById(int64(createdWorkout.ID))
			require.NoError(t, err)
			assert.Equal(t, retrieved.Title, createdWorkout.Title)
			assert.Equal(t, retrieved.Description, createdWorkout.Description)
			assert.Equal(t, retrieved.DurationMinutes, createdWorkout.DurationMinutes)
			assert.Equal(t, retrieved.CaloriesBurned, createdWorkout.CaloriesBurned)
			assert.Equal(t, retrieved.ID, createdWorkout.ID)
			assert.Equal(t, len(retrieved.Entries), len(createdWorkout.Entries))

			for i, entry := range retrieved.Entries {
				assert.Equal(t, tt.workout.Entries[i].ExerciseName, entry.ExerciseName)
				assert.Equal(t, tt.workout.Entries[i].Sets, entry.Sets)
				assert.Equal(t, tt.workout.Entries[i].OrderIndex, entry.OrderIndex)
			}
		})
	}
}

func IntPtr(i int) *int {
	return &i
}
func FloatPtr(i float64) *float64 {
	return &i
}
