package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (app *application) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := listTasks(app.db)
	if err != nil {
		http.Error(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (app *application) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Title) == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	task, err := createTask(app.db, body.Title, body.Description)
	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (app *application) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	task, err := getTask(app.db, id)
	if err == sql.ErrNoRows {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get task", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (app *application) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Done        bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	task, err := updateTask(app.db, id, body.Title, body.Description, body.Done)
	if err == sql.ErrNoRows {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (app *application) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := deleteTask(app.db, id); err == sql.ErrNoRows {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pathID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
