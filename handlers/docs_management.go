package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func CreateDocHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("create new doc route! \n")

	// Parse the multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get category and title
	category := r.FormValue("category")
	title := r.FormValue("title")

	if category == "" || title == "" {
		http.Error(w, "Category and title are required", http.StatusBadRequest)
		return
	}

	// Sanitize title for directory name
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		http.Error(w, "Failed to compile regex", http.StatusInternalServerError)
		return
	}
	sanitizedTitle := strings.ToLower(reg.ReplaceAllString(title, "-"))

	// Create directory structure
	docPath := filepath.Join("documentations", category, sanitizedTitle)
	if err := os.MkdirAll(docPath, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create document directory: %s", err), http.StatusInternalServerError)
		return
	}

	// Create subdirectories
	for _, dir := range []string{"media", "refs", "objects"} {
		if err := os.MkdirAll(filepath.Join(docPath, dir), 0755); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create subdirectory %s: %s", dir, err), http.StatusInternalServerError)
			return
		}
	}

	// Create main.md
	mainMdPath := filepath.Join(docPath, "main.md")
	dst, err := os.Create(mainMdPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create main.md: %s", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to main.md: %s", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Document created successfully",
		"path":    docPath,
		"file":    handler.Filename,
	})
}
