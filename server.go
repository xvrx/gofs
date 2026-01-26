package main

import (
	"fmt"
	"net/http"
	"os"
	"watcher/config"
	"watcher/handlers"

	"github.com/gorilla/mux" // Import Gorilla Mux
	"github.com/robfig/cron/v3"
)

func skip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		return
	}

	// Initialize Redis client
	if err := config.InitRedis(); err != nil {
		fmt.Printf("Error initializing Redis: %s\n", err)
		return
	}

	// Initialize MySQL client
	if err := config.InitMySQL(); err != nil {
		fmt.Printf("Error initializing MySQL: %s\n", err)
		return
	}

	c := cron.New()
	c.AddFunc("@daily", func() {
		fmt.Println("Running daily cleanup...")
		cleanupDirs("tmp/pdfcompression/input", "tmp/pdfcompression/output")
	})
	c.Start()

	router := mux.NewRouter() // Create a new Gorilla Mux router

	// Public routes (no authentication required)
	router.HandleFunc("/auth/login", handlers.LoginHandler).Methods("POST")

	// ------------ Auth Middleware ---------------------
	authenticatedRouter := router.PathPrefix("/").Subrouter()
	// uncomment line below
	// authenticatedRouter.Use(handlers.AuthMiddleware)


	//! ---------- no auth --- dev only -- 
	//! to activate auth on route,comment line below and uncomment out the auth middleware
	authenticatedRouter.Use(skip)



	// ---------- route requiring authentication
	authenticatedRouter.HandleFunc("/", handlers.HomeHandler).Methods("GET")

	authenticatedRouter.HandleFunc("/outbox/update", handlers.UpdateOutboxHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/outbox/get", handlers.GetOutboxData).Methods("GET")
	authenticatedRouter.HandleFunc("/docvault/update", handlers.UpdateDocVaultHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/docvault/get", handlers.GetDocVaultHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/auth/session", handlers.GetSessionHandler).Methods("GET")
	authenticatedRouter.HandleFunc("/mfwp/get/{npwp:[0-9]{15}}", handlers.GetMfwpData).Methods("GET")
	authenticatedRouter.HandleFunc("/utils/pdfcompression", handlers.PDFCompressionHandler).Methods("POST")

	fmt.Println("Server starting on port http://localhost:3000/ ")
	// Use the Gorilla Mux router
	if err := http.ListenAndServe("localhost:3000", router); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func cleanupDirs(dirs ...string) {
	for _, dir := range dirs {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("Error cleaning up directory %s: %v\n", dir, err)
		} else {
			fmt.Printf("Cleaned up directory: %s\n", dir)
		}
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Error recreating directory %s: %v\n", dir, err)
		}
	}
}

