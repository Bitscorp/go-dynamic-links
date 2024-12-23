package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() *sql.DB {
	database, err := sql.Open("sqlite3", "./links.db")
	if err != nil {
		log.Fatal(err)
	}

	// Create links table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS links (
		project_name TEXT PRIMARY KEY,
		ios_link TEXT,
		android_link TEXT,
		web_link TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = database.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	return database
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Initialize database
	db = initDB()
	defer db.Close()

	r := mux.NewRouter()

	// Register routes
	r.HandleFunc("/projects/{name}/links/new", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		projectName := vars["name"]

		var links struct {
			IOS     string `json:"ios_link"`
			Android string `json:"android_link"`
			Web     string `json:"web_link"`
		}

		if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec(`
			INSERT INTO links (project_name, ios_link, android_link, web_link)
			VALUES (?, ?, ?, ?)`,
			projectName, links.IOS, links.Android, links.Web)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}).Methods("POST")

	r.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		projectName := vars["name"]

		var links struct {
			IOS     string
			Android string
			Web     string
		}

		err := db.QueryRow(`
			SELECT ios_link, android_link, web_link
			FROM links WHERE project_name = ?`,
			projectName).Scan(&links.IOS, &links.Android, &links.Web)

		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userAgent := r.Header.Get("User-Agent")
		switch {
		case strings.Contains(userAgent, "iPhone") || strings.Contains(userAgent, "iPad"):
			http.Redirect(w, r, links.IOS, http.StatusFound)
		case strings.Contains(userAgent, "Android"):
			http.Redirect(w, r, links.Android, http.StatusFound)
		default:
			http.Redirect(w, r, links.Web, http.StatusFound)
		}
	}).Methods("GET")

	r.HandleFunc("/health", healthCheck).Methods("GET")

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}