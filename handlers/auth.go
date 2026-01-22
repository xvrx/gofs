package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"watcher/config" // Import the config package

	"github.com/go-redis/redis/v8"
)

// AuthData represents the structure of the authentication data stored in Redis
type AuthData struct {
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id"`
	Jabatan      string `json:"jabatan"`
	IP           string `json:"ip"`
	IPvX         string `json:"ipvx"`
}

// GetSessionHandler checks for an existing key in Redis and returns the authentication data.
func GetSessionHandler(w http.ResponseWriter, r *http.Request) {

	// For demonstration, using a hardcoded key. In a real app, this would come from a cookie/header.
	sessionKey := "817931767|sess:3c94a742-97a7-4b6d-b8c9-73274220be1f"

	ctx := context.Background()
	val, err := config.RedisClient.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		http.Error(w, "Session key not found in Redis", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get session from Redis: %v", err), http.StatusInternalServerError)
		return
	}

	var authData AuthData
	err = json.Unmarshal([]byte(val), &authData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal auth data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{"status": true, "data": authData}
	json.NewEncoder(w).Encode(response)
}
