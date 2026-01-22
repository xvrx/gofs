package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DocItem struct {
	FileName string `json:"fileName"`
	FullPath string `json:"fullpath"`
}

// UpdateDocVaultHandler processes files in src/scanned, categorizes them by owner,
// writes the data to data.json, and returns the full categorized data.
func UpdateDocVaultHandler(w http.ResponseWriter, r *http.Request) {
	scannedDir := "src/scanned"

	files, err := os.ReadDir(scannedDir)
	if err != nil {
		http.Error(w, "Failed to read scanned directory", http.StatusInternalServerError)
		return
	}

	docsByOwner := make(map[string][]DocItem)

	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			parts := strings.SplitN(filename, "_", 2)
			if len(parts) == 2 {
				owner := strings.ToLower(parts[0])
				docItem := DocItem{
					FileName: strings.TrimSuffix(parts[1], filepath.Ext(parts[1])),
					FullPath: "/" + filename,
				}
				docsByOwner[owner] = append(docsByOwner[owner], docItem)
			}
		}
	}

	// Always write the full data to data.json
	jsonData, err := json.MarshalIndent(docsByOwner, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal data to JSON", http.StatusInternalServerError)
		return
	}

	jsonPath := filepath.Join("src", "libs", "scanned.json")
	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		http.Error(w, "Failed to write JSON file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{"status": true, "data": docsByOwner} // Always return full data
	json.NewEncoder(w).Encode(response)
}

// GetDocVaultHandler reads data from data.json and filters it based on the owner query parameter.
func GetDocVaultHandler(w http.ResponseWriter, r *http.Request) {
	jsonPath := filepath.Join("src", "libs", "scanned.json")

	// Get owner from query parameter
	// filterOwner := r.URL.Query().Get("owner")
	filterOwner := "bayu"

	// Read the data.json file
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		http.Error(w, "Failed to read scanned.json", http.StatusInternalServerError)
		return
	}

	var docsByOwner map[string][]DocItem
	err = json.Unmarshal(jsonData, &docsByOwner)
	if err != nil {
		http.Error(w, "Failed to unmarshal scanned.json", http.StatusInternalServerError)
		return
	}

	var responseData any

	if filterOwner != "" {
		filterOwner = strings.ToLower(filterOwner)
		if filteredDocs, ok := docsByOwner[filterOwner]; ok {
			responseData = map[string][]DocItem{filterOwner: filteredDocs}
		} else {
			responseData = map[string][]DocItem{filterOwner: {}}
		}
	} else {
		responseData = docsByOwner
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{"status": true, "data": responseData}
	json.NewEncoder(w).Encode(response)
}
