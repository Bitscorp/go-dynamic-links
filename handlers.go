package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Links struct {
	IOS     string `json:"ios"`
	Android string `json:"android"`
	Web     string `json:"web"`
}

func createLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["name"]

	var links Links
	if err := json.NewDecoder(r.Body).Decode(&links); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert or update links in database
	query := `
		INSERT OR REPLACE INTO links (project_name, ios_link, android_link, web_link)
		VALUES (?, ?, ?, ?)`

	_, err := db.Exec(query, projectName, links.IOS, links.Android, links.Web)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(links)
}

func redirectToApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	// Query the database for links
	var links Links
	query := `SELECT ios_link, android_link, web_link FROM links WHERE project_name = ?`
	err := db.QueryRow(query, name).Scan(&links.IOS, &links.Android, &links.Web)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Get user agent
	userAgent := strings.ToLower(r.UserAgent())

	// Determine platform and redirect
	switch {
	case strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad"):
		http.Redirect(w, r, links.IOS, http.StatusFound)
	case strings.Contains(userAgent, "android"):
		http.Redirect(w, r, links.Android, http.StatusFound)
	default:
		http.Redirect(w, r, links.Web, http.StatusFound)
	}
}