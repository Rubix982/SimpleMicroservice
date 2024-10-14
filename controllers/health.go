package controllers

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

// HandleHealthCheck provides a health check endpoint
func HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Info("Health check requested")
}
