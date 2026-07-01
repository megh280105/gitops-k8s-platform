package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type application struct {
	db       *sql.DB
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://taskapi:changeme@localhost:5432/tasks?sslmode=disable"
	}

	db := connectWithRetry(dbURL, 30*time.Second)
	runMigrations(dbURL)

	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "taskapi_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "taskapi_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	prometheus.MustRegister(requests, duration)

	app := &application{db: db, requests: requests, duration: duration}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.handleHealth)
	mux.HandleFunc("GET /ready", app.handleReady)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /tasks", app.handleListTasks)
	mux.HandleFunc("POST /tasks", app.handleCreateTask)
	mux.HandleFunc("GET /tasks/{id}", app.handleGetTask)
	mux.HandleFunc("PUT /tasks/{id}", app.handleUpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", app.handleDeleteTask)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("taskapi listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, app.instrument(mux)))
}

func (app *application) instrument(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		elapsed := time.Since(start).Seconds()
		status := strconv.Itoa(rw.status)
		app.requests.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		app.duration.WithLabelValues(r.Method, r.URL.Path).Observe(elapsed)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (app *application) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (app *application) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := app.db.Ping(); err != nil {
		http.Error(w, `{"status":"not ready"}`, http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func connectWithRetry(url string, timeout time.Duration) *sql.DB {
	deadline := time.Now().Add(timeout)
	for {
		db, err := sql.Open("postgres", url)
		if err == nil {
			if err = db.Ping(); err == nil {
				log.Println("database connected")
				return db
			}
			db.Close()
		}
		if time.Now().After(deadline) {
			log.Fatalf("could not connect to database after %s: %v", timeout, err)
		}
		log.Printf("waiting for database... (%v)", err)
		time.Sleep(2 * time.Second)
	}
}

func runMigrations(dbURL string) {
	m, err := migrate.New("file:///migrations", dbURL)
	if err != nil {
		log.Fatalf("migration init failed: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("migrations applied")
}

