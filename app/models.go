package main

import (
	"database/sql"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
}

func listTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query(`SELECT id, title, description, done, created_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func getTask(db *sql.DB, id int) (Task, error) {
	var t Task
	err := db.QueryRow(
		`SELECT id, title, description, done, created_at FROM tasks WHERE id = $1`, id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt)
	return t, err
}

func createTask(db *sql.DB, title, description string) (Task, error) {
	var t Task
	err := db.QueryRow(
		`INSERT INTO tasks (title, description) VALUES ($1, $2)
		 RETURNING id, title, description, done, created_at`,
		title, description,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt)
	return t, err
}

func updateTask(db *sql.DB, id int, title, description string, done bool) (Task, error) {
	var t Task
	err := db.QueryRow(
		`UPDATE tasks SET title=$1, description=$2, done=$3
		 WHERE id=$4
		 RETURNING id, title, description, done, created_at`,
		title, description, done, id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Done, &t.CreatedAt)
	return t, err
}

func deleteTask(db *sql.DB, id int) error {
	res, err := db.Exec(`DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
