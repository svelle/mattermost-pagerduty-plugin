package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/svelle/mattermost-pagerduty-plugin/server/pagerduty"
)

func (p *Plugin) handleGetSchedules(w http.ResponseWriter, r *http.Request) {
	p.client.Log.Debug("handleGetSchedules called", "user_id", r.Header.Get("Mattermost-User-ID"))

	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		p.client.Log.Warn("Plugin configuration invalid", "error", err)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.config.invalid",
			Message:    "Plugin not configured",
			StatusCode: http.StatusNotImplemented,
		})
		return
	}

	client := p.createPagerDutyClient(config.APIToken, config.APIBaseURL)
	p.client.Log.Debug("Fetching schedules from PagerDuty API", "base_url", config.APIBaseURL)

	schedules, err := client.GetSchedules(100, 0)
	if err != nil {
		p.client.Log.Error("Failed to get schedules from PagerDuty", "error", err.Error())
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.schedules.error",
			Message:    "Failed to retrieve schedules",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	p.client.Log.Info("Successfully retrieved schedules", "count", len(schedules.Schedules))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedules); err != nil {
		p.client.Log.Error("Failed to encode schedules response", "error", err.Error())
	}
}

func (p *Plugin) handleGetOnCalls(w http.ResponseWriter, r *http.Request) {
	p.client.Log.Debug("handleGetOnCalls called", "user_id", r.Header.Get("Mattermost-User-ID"))

	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		p.client.Log.Warn("Plugin configuration invalid", "error", err)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.config.invalid",
			Message:    "Plugin not configured",
			StatusCode: http.StatusNotImplemented,
		})
		return
	}

	client := p.createPagerDutyClient(config.APIToken, config.APIBaseURL)

	scheduleID := r.URL.Query().Get("schedule_id")
	var oncalls *pagerduty.OnCallsResponse
	var err error

	if scheduleID != "" {
		p.client.Log.Debug("Fetching on-calls for specific schedule", "schedule_id", scheduleID)
		oncalls, err = client.GetOnCallsForSchedule(scheduleID)
	} else {
		p.client.Log.Debug("Fetching current on-calls for all schedules")
		oncalls, err = client.GetCurrentOnCalls()
	}

	if err != nil {
		p.client.Log.Error("Failed to get on-calls from PagerDuty", "error", err.Error(), "schedule_id", scheduleID)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.oncalls.error",
			Message:    "Failed to retrieve on-call users",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	p.client.Log.Info("Successfully retrieved on-calls", "count", len(oncalls.OnCalls))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(oncalls); err != nil {
		p.client.Log.Error("Failed to encode on-calls response", "error", err.Error())
	}
}

func (p *Plugin) handleGetScheduleDetails(w http.ResponseWriter, r *http.Request) {
	p.client.Log.Debug("handleGetScheduleDetails called", "user_id", r.Header.Get("Mattermost-User-ID"))

	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		p.client.Log.Warn("Plugin configuration invalid", "error", err)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.config.invalid",
			Message:    "Plugin not configured",
			StatusCode: http.StatusNotImplemented,
		})
		return
	}

	scheduleID := r.URL.Query().Get("id")
	if scheduleID == "" {
		p.client.Log.Warn("Schedule ID missing in request")
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.schedule.id.missing",
			Message:    "Schedule ID is required",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	client := p.createPagerDutyClient(config.APIToken, config.APIBaseURL)

	// Get schedule with the next 48 hours of coverage
	now := time.Now()
	until := now.Add(48 * time.Hour)

	p.client.Log.Debug("Fetching schedule details", "schedule_id", scheduleID, "from", now.Format(time.RFC3339), "until", until.Format(time.RFC3339))
	schedule, err := client.GetSchedule(scheduleID, now, until)
	if err != nil {
		p.client.Log.Error("Failed to get schedule details from PagerDuty", "error", err.Error(), "schedule_id", scheduleID)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.schedule.error",
			Message:    "Failed to retrieve schedule details",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	p.client.Log.Info("Successfully retrieved schedule details", "schedule_id", scheduleID, "name", schedule.Schedule.Name)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedule); err != nil {
		p.client.Log.Error("Failed to encode schedule response", "error", err.Error())
	}
}

func (p *Plugin) handleGetServices(w http.ResponseWriter, r *http.Request) {
	p.client.Log.Debug("handleGetServices called", "user_id", r.Header.Get("Mattermost-User-ID"))

	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		p.client.Log.Warn("Plugin configuration invalid", "error", err)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.config.invalid",
			Message:    "Plugin not configured",
			StatusCode: http.StatusNotImplemented,
		})
		return
	}

	client := p.createPagerDutyClient(config.APIToken, config.APIBaseURL)
	p.client.Log.Debug("Fetching services from PagerDuty API", "base_url", config.APIBaseURL)

	services, err := client.GetServices(100, 0)
	if err != nil {
		p.client.Log.Error("Failed to get services from PagerDuty", "error", err.Error())
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.services.error",
			Message:    "Failed to retrieve services",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	p.client.Log.Info("Successfully retrieved services", "count", len(services.Services))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(services); err != nil {
		p.client.Log.Error("Failed to encode services response", "error", err.Error())
	}
}

// CreateIncidentRequest represents the request body for creating an incident
type CreateIncidentRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	ServiceID   string   `json:"service_id"`
	AssigneeIDs []string `json:"assignee_ids,omitempty"`
}

func (p *Plugin) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	p.client.Log.Debug("handleCreateIncident called", "user_id", r.Header.Get("Mattermost-User-ID"))

	if r.Method != http.MethodPost {
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.incident.method.invalid",
			Message:    "Method not allowed",
			StatusCode: http.StatusMethodNotAllowed,
		})
		return
	}

	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		p.client.Log.Warn("Plugin configuration invalid", "error", err)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.config.invalid",
			Message:    "Plugin not configured",
			StatusCode: http.StatusNotImplemented,
		})
		return
	}

	var req CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		p.client.Log.Warn("Failed to decode create incident request", "error", err)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.incident.decode.error",
			Message:    "Invalid request body",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	if req.Title == "" || req.ServiceID == "" {
		p.client.Log.Warn("Missing required fields in create incident request", "title", req.Title, "service_id", req.ServiceID)
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.incident.fields.missing",
			Message:    "Title and service_id are required",
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	client := p.createPagerDutyClient(config.APIToken, config.APIBaseURL)
	p.client.Log.Debug("Creating incident in PagerDuty", "title", req.Title, "service_id", req.ServiceID, "assignees", len(req.AssigneeIDs))

	incident, err := client.CreateIncident(req.Title, req.Description, req.ServiceID, req.AssigneeIDs)
	if err != nil {
		p.client.Log.Error("Failed to create incident in PagerDuty", "error", err.Error())
		p.handleError(w, r, &APIError{
			ID:         "api.pagerduty.incident.create.error",
			Message:    "Failed to create incident",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	p.client.Log.Info("Successfully created incident", "incident_id", incident.Incident.ID, "title", incident.Incident.Title)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(incident); err != nil {
		p.client.Log.Error("Failed to encode create incident response", "error", err.Error())
	}
}
