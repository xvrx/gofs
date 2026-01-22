package handlers

import (
	"fmt"
	"net/http"
)

// HomeHandler handles requests to the root URL.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from the handlers package!")
}
