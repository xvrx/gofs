package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"watcher/config"

	"github.com/gorilla/mux"
)

func GetMfwpData(w http.ResponseWriter, r *http.Request) {
	var _ *sql.DB
	vars := mux.Vars(r)
	npwp := vars["npwp"]

	// Validate NPWP: must be 15 digits
	if match, _ := regexp.MatchString(`^\d{15}$`, npwp); !match {
		http.Error(w, "Invalid NPWP format. It must be 15 digits.", http.StatusBadRequest)
		return
	}

	db, ok := config.DB["mfwp"]
	if !ok {
		http.Error(w, "Database connection for 'mfwp' not found", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT * FROM masterfile WHERE NPWP_15 = ? LIMIT 1", npwp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.NotFound(w, r)
		}
		return
	}

	cols, err := rows.Columns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make(map[string]interface{})
	for i, col := range cols {
		val := values[i]
		b, ok := val.([]byte)
		if ok {
			result[col] = string(b)
		} else {
			result[col] = val
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
