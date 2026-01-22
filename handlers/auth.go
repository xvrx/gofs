package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"watcher/config"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// AuthData represents the structure of the authentication data stored in Redis
type AuthData struct {
	UserID       string `json:"user_id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	NIP          string `json:"nip"`
	DepartmentID string `json:"department_id"`
	Jabatan      string `json:"jabatan"`
	IP           string `json:"ip"`
	IPvX         string `json:"ipvx"`
}

// LoginRequest represents the structure of the login request payload
type LoginRequest struct {
	NIP      string `json:"nip"`
	Password string `json:"password"`
}

type contextKey string

const (
	AuthContextKey contextKey = "auth"
)

// getSessionDataFromRedis retrieves session data from Redis using the session token from cookies.
func getSessionDataFromRedis(r *http.Request) (*AuthData, error) {
	// client cookie side = session_token
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, fmt.Errorf("session token cookie not found: %w", err)
	}
	sessionToken := cookie.Value

	// if token exist, match session_token in redisDB
	ctx := context.Background()                                    // ctx is required when making redis Get() call                              // empty context - no call time limit
	val, err := config.RedisClient.Get(ctx, sessionToken).Result() // Use sessionToken directly as key
	if err == redis.Nil {
		return nil, fmt.Errorf("session not found in Redis")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	var authData AuthData
	err = json.Unmarshal([]byte(val), &authData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth data: %w", err)
	}

	return &authData, nil
}




// AuthMiddleware checks for a valid session token and authenticates the request.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authData, err := getSessionDataFromRedis(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Store AuthData in request context
		ctx := context.WithValue(r.Context(), AuthContextKey, authData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}












// GetSessionHandler checks for an existing session in Redis and returns the authentication data.
func GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	authData, err := getSessionDataFromRedis(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{"status": true, "data": authData}
	json.NewEncoder(w).Encode(response)
}





// LoginHandler handles user login requests
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 1. Query MySQL to validate credentials
	var user AuthData // Re-using AuthData struct for user info
	query := "SELECT user_id, role, name, nip, jabatan, department_id FROM users WHERE nip = ? AND password = ?"
	row := config.DB.QueryRow(query, req.NIP, req.Password)
	err = row.Scan(&user.UserID, &user.Role, &user.Name, &user.NIP, &user.Jabatan, &user.DepartmentID)

	if err == sql.ErrNoRows {
		http.Error(w, "Invalid NIP or password", http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Generate Session Token
	sessionToken := uuid.New().String()
	sessionKey := sessionToken // Store sessionToken directly as the key

	// Populate IP addresses (can be improved with actual IP detection)
	user.IP = r.RemoteAddr   // Basic remote address
	user.IPvX = r.RemoteAddr // Placeholder, actual IPvX detection is more complex

	// Marshal user data to JSON for Redis
	userDataJSON, err := json.Marshal(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal user data: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. Store Session in Redis (e.g., expire after 24 hours)
	ctx := context.Background()
	err = config.RedisClient.Set(ctx, sessionKey, userDataJSON, 24*time.Hour).Err()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to store session in Redis: %v", err), http.StatusInternalServerError)
		return
	}

	// 4. Set HTTP-only Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false, // vscode client
		// Secure:   true, // if using real client
		Path: "/", // use cookies on all path
	})

	// 5. Respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{"status": true, "message": "Login successful"}
	json.NewEncoder(w).Encode(response)
}
