package store

import (
	"database/sql"
	"fmt"
)

type Workout struct {
	ID              int            `json:"id"`
	UserID          int            `json:"user_id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	DurationMinutes int            `json:"duration_minutes"`
	CaloriesBurned  int            `json:"calories_burned"`
	Entries         []WorkoutEntry `json:"entries"`
}

type WorkoutEntry struct {
	ID              int      `json:"id"`
	ExerciseName    string   `json:"exercise_name"`
	Sets            int      `json:"sets"`
	Reps            *int     `json:"reps"`
	DurationSeconds *int     `json:"duration_seconds"`
	Weight          *float64 `json:"weight"`
	Notes           string   `json:"notes"`
	OrderIndex      int      `json:"order_index"`
}

type PostgresWorkoutStore struct {
	db *sql.DB
}

func NewPostgresWorkoutStore(db *sql.DB) *PostgresWorkoutStore {
	return &PostgresWorkoutStore{db: db}
}

type WorkoutStore interface {
	CreateWorkout(*Workout) (*Workout, error)
	GetWorkoutById(id int64) (*Workout, error)
	UpdateWorkout(*Workout) error
	DeleteWorkout(id int64) error
	GetWorkoutOwner(id int64) (int, error)
}

func (pg *PostgresWorkoutStore) CreateWorkout(w *Workout) (*Workout, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	query :=
		`INSERT INTO workouts (user_id, title,description, duration_minutes, calories_burned)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id
		`
	err = tx.QueryRow(query, w.UserID, w.Title, w.Description, w.DurationMinutes, w.CaloriesBurned).Scan(&w.ID)
	if err != nil {
		return nil, err
	}
	for index, entry := range w.Entries {
		query :=
			`INSERT INTO workout_entries (workout_id, exercise_name, sets, reps, duration_seconds, weight, notes, order_index)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING id			
			`
		err = tx.QueryRow(query, w.ID, entry.ExerciseName, entry.Sets, entry.Reps, entry.DurationSeconds, entry.Weight, entry.Notes, entry.OrderIndex).Scan(&w.Entries[index].ID)
		if err != nil {
			return nil, err
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (pg *PostgresWorkoutStore) GetWorkoutById(id int64) (*Workout, error) {
	var workout Workout
	query :=
		`SELECT id,title,description,duration_minutes,calories_burned
		 FROM workouts 
		 WHERE id = $1
		 `
	err := pg.db.QueryRow(query, id).Scan(&workout.ID, &workout.Title, &workout.Description, &workout.DurationMinutes, &workout.CaloriesBurned)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	entryQuery :=
		`SELECT id,exercise_name,sets,reps,duration_seconds,weight,notes,order_index
		 FROM workout_entries
		 WHERE workout_id = $1
		 ORDER BY order_index
		`
	rows, err := pg.db.Query(entryQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry WorkoutEntry
		err := rows.Scan(
			&entry.ID,
			&entry.ExerciseName,
			&entry.Sets,
			&entry.Reps,
			&entry.DurationSeconds,
			&entry.Weight,
			&entry.Notes,
			&entry.OrderIndex,
		)
		if err != nil {
			return nil, err
		}
		workout.Entries = append(workout.Entries, entry)
	}
	return &workout, nil
}
func (pg *PostgresWorkoutStore) UpdateWorkout(w *Workout) error {
	tx, err := pg.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query :=
		`UPDATE workouts
	 SET title = $1, description = $2, duration_minutes = $3, calories_burned = $4
	 WHERE id = $5
	 RETURNING id
	`
	var updatedID int
	err = tx.QueryRow(query, w.Title, w.Description, w.DurationMinutes, w.CaloriesBurned, w.ID).Scan(&updatedID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("workout not found")
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		DELETE FROM workout_entries WHERE workout_id = $1
	`, w.ID)
	if err != nil {
		return err
	}
	for _, entry := range w.Entries {
		query := `
		INSERT INTO workout_entries (workout_id, exercise_name,sets,reps,duration_seconds,weight,notes,order_index)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id
		`
		err = tx.QueryRow(query, w.ID, entry.ExerciseName, entry.Sets, entry.Reps, entry.DurationSeconds, entry.Weight, entry.Notes, entry.OrderIndex).Scan(&entry.ID)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgresWorkoutStore) DeleteWorkout(id int64) error {
	query :=
		`DELETE FROM workouts 
	   WHERE id = $1
	 `
	result, err := pg.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgresWorkoutStore) GetWorkoutOwner(workoutId int64) (int, error) {
	var userId int

	query := `
		SELECT user_id 
		FROM workouts
		WHERE id = $1
	`
	err := pg.db.QueryRow(query, workoutId).Scan(&userId)
	if err != nil {
		return 0, nil
	}
	return userId, nil
}
