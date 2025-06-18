package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// ServeHTTP handles HTTP requests to the plugin.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()

	// Middleware to require that the user is logged in
	router.Use(p.MattermostAuthorizationRequired)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// PagerDuty endpoints
	apiRouter.HandleFunc("/schedules", p.handleGetSchedules).Methods(http.MethodGet)
	apiRouter.HandleFunc("/oncalls", p.handleGetOnCalls).Methods(http.MethodGet)
	apiRouter.HandleFunc("/schedule", p.handleGetScheduleDetails).Methods(http.MethodGet)

	router.ServeHTTP(w, r)
}

func (p *Plugin) MattermostAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type APIError struct {
	ID         string `json:"id"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (p *Plugin) handleError(w http.ResponseWriter, r *http.Request, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)

	if encErr := json.NewEncoder(w).Encode(err); encErr != nil {
		p.client.Log.Error("Failed to encode error response", "error", encErr.Error())
	}
}
