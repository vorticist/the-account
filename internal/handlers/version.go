package handlers

import (
	"fmt"
	"net/http"
)

// VersionCode will be set at build time
var VersionCode string = "non-versioned"

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return the version code
	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintf(w, "%s", VersionCode)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
