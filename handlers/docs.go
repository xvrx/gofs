package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"watcher/config"
)

type FileInfo struct {
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	LastUpdated time.Time `json:"last_updated"`
}

func DocsGenerateHandler(w http.ResponseWriter, r *http.Request) {
	db, ok := config.DB["documentations"]
	if !ok {
		http.Error(w, "Database connection 'documentations' not found", http.StatusInternalServerError)
		return
	}

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS raw_list (
		id INT AUTO_INCREMENT PRIMARY KEY,
		file_name VARCHAR(255) NOT NULL,
		file_path VARCHAR(255) NOT NULL,
		last_updated TIMESTAMP NOT NULL,
		UNIQUE KEY (file_path)
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create table: %s", err), http.StatusInternalServerError)
		return
	}

	dirPath := "documentations/raw"

	// Check if the directory exists, create it if it doesn't
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create directory: %s", err), http.StatusInternalServerError)
			return
		}
	}

	var files []FileInfo
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, FileInfo{
				FileName:    info.Name(),
				FilePath:    path,
				LastUpdated: info.ModTime(),
			})
		}
		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read directory: %s", err), http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "No files found in directory to process"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to begin transaction: %s", err), http.StatusInternalServerError)
		return
	}

	stmt, err := tx.Prepare(`
		INSERT INTO raw_list (file_name, file_path, last_updated)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE last_updated = VALUES(last_updated)
	`)
	if err != nil {
		tx.Rollback()
		http.Error(w, fmt.Sprintf("Failed to prepare statement: %s", err), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	for _, file := range files {
		_, err := stmt.Exec(file.FileName, strings.Replace(file.FilePath, "\\", "/", -1), file.LastUpdated)
		if err != nil {
			tx.Rollback()
			http.Error(w, fmt.Sprintf("Failed to execute statement for file %s: %s", file.FileName, err), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to commit transaction: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully inserted/updated file list in the database"})
}
