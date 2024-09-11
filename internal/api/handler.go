package api

import (
	"fmt"
	"net/http"
)

// HomeHandler responds with a welcome message.
// This function is designed to be used as an HTTP handler.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is GET.
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Respond with a simple welcome message.
	fmt.Fprintf(w, "Welcome to our Golang API!")
}
