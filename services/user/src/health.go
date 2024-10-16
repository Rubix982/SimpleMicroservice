package src

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// HandleHealthCheck provides a health check endpoint
func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Info("Health check requested")
}
