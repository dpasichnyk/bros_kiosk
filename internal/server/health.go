package server

import (
	"net/http"
)

// HealthHandler returns a simple OK status
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
