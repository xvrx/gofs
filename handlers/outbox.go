package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

// UpdateOutboxHandler reads an Excel file, converts its data to JSON, and saves it.
func UpdateOutboxHandler(w http.ResponseWriter, r *http.Request) {
	excelPath := filepath.Join("src", "libs", "outbox.xlsx")
	jsonPath := "src/libs/outbox.json"
	sheetName := "Sheet1"

	// Open the Excel file
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open Excel file: %v", err), http.StatusInternalServerError)
		return
	}

	// close excel file after done
	defer f.Close()

	// Get all the rows from the specified sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get rows from sheet '%s': %v", sheetName, err), http.StatusInternalServerError)
		return
	}

	if len(rows) < 2 {
		http.Error(w, "No data found in Excel sheet or header row is missing", http.StatusBadRequest)
		return
	}

	// Assume the first row is the header
	header := rows[0]
	data := make([]map[string]any, 0)

	// Iterate over rows (skipping the header)
	for _, row := range rows[1:] {
		rowData := make(map[string]any)
		for i, colCell := range row {
			if i < len(header) {
				// Use header as key
				rowData[header[i]] = colCell
			}
		}
		data = append(data, rowData)
	}

	// Marshal data into JSON format
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal data to JSON: %v", err), http.StatusInternalServerError)
		return
	}

	// Write the JSON data to the output file
	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write JSON file: %v", err), http.StatusInternalServerError)
		return
	}

	// respond with object
	wrapper := map[string]any{
		"status":  "success",
		"message": "Successfully converted",
		"data":    json.RawMessage(jsonData), // embed the already marshalled JSON
	}

	// convert to json object
	finalJSON, err := json.Marshal(wrapper)
	if err != nil {
		http.Error(w, "Failed to marshal wrapper", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(finalJSON)
}

// GetOutboxData serves the content of src/outbox/data.json
func GetOutboxData(w http.ResponseWriter, r *http.Request) {
	jsonPath := filepath.Join("src", "libs", "outbox.json")

	// Read the JSON file from the disk
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read data.json: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a wrapper for the JSON data
	wrapper := map[string]interface{}{
		"status": "true",
		"data":   json.RawMessage(jsonData),
	}

	// Marshal the wrapper into a JSON object
	response, err := json.Marshal(wrapper)
	if err != nil {
		http.Error(w, "Failed to create response object", http.StatusInternalServerError)
		return
	}

	// Set the content type and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
