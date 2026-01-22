package main

import (
	"fmt"
	"net/http"
	"watcher/config"
	"watcher/handlers"

	"github.com/gorilla/mux" // Import Gorilla Mux
)

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

	router := mux.NewRouter() // Create a new Gorilla Mux router

	// Public routes (no authentication required)
	router.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	router.HandleFunc("/auth/login", handlers.LoginHandler).Methods("POST")

	// for dev
	router.HandleFunc("/outbox/update", handlers.UpdateOutboxHandler).Methods("GET")
	router.HandleFunc("/outbox/get", handlers.GetOutboxData).Methods("GET")
	router.HandleFunc("/docvault/update", handlers.UpdateDocVaultHandler).Methods("GET")
	router.HandleFunc("/docvault/get", handlers.GetDocVaultHandler).Methods("GET")
	router.HandleFunc("/auth/session", handlers.GetSessionHandler).Methods("GET")

	// // Authenticated routes subrouter
	// authenticatedRouter := router.PathPrefix("/").Subrouter()
	// authenticatedRouter.Use(handlers.AuthMiddleware)
	

	// // Register handlers with specific HTTP methods that require authentication
	// authenticatedRouter.HandleFunc("/outbox/update", handlers.UpdateOutboxHandler).Methods("GET")
	// authenticatedRouter.HandleFunc("/outbox/get", handlers.GetOutboxData).Methods("GET")
	// authenticatedRouter.HandleFunc("/docvault/update", handlers.UpdateDocVaultHandler).Methods("GET")
	// authenticatedRouter.HandleFunc("/docvault/get", handlers.GetDocVaultHandler).Methods("GET")
	// authenticatedRouter.HandleFunc("/auth/session", handlers.GetSessionHandler).Methods("GET")


	fmt.Println("Server starting on port http://localhost:3000/ ")
	// Use the Gorilla Mux router
	if err := http.ListenAndServe("localhost:3000", router); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
