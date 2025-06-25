package pagerduty

import "time"

type Schedule struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	TimeZone         string           `json:"time_zone"`
	Summary          string           `json:"summary"`
	ScheduleLayers   []ScheduleLayer  `json:"schedule_layers,omitempty"`
	OverrideSubcycle OverrideSubcycle `json:"override_subcycle,omitempty"`
	FinalSchedule    FinalSchedule    `json:"final_schedule,omitempty"`
}

type ScheduleLayer struct {
	ID                        string          `json:"id"`
	Name                      string          `json:"name"`
	Start                     time.Time       `json:"start"`
	End                       *time.Time      `json:"end"`
	RotationVirtualStart      time.Time       `json:"rotation_virtual_start"`
	RotationTurnLengthSeconds int             `json:"rotation_turn_length_seconds"`
	Users                     []UserReference `json:"users"`
}

type OverrideSubcycle struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type FinalSchedule struct {
	Name                    string                  `json:"name"`
	RenderedScheduleEntries []RenderedScheduleEntry `json:"rendered_schedule_entries"`
}

type ScheduleEntry struct {
	User  UserReference `json:"user"`
	Start time.Time     `json:"start"`
	End   time.Time     `json:"end"`
}

type UserReference struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

type User struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Email          string          `json:"email"`
	Type           string          `json:"type"`
	Summary        string          `json:"summary"`
	Description    string          `json:"description"`
	Role           string          `json:"role"`
	TimeZone       string          `json:"time_zone"`
	Color          string          `json:"color"`
	AvatarURL      string          `json:"avatar_url"`
	ContactMethods []ContactMethod `json:"contact_methods,omitempty"`
}

type ContactMethod struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
	Label   string `json:"label"`
	Address string `json:"address"`
}

type OnCall struct {
	User             User              `json:"user"`
	Schedule         Schedule          `json:"schedule"`
	EscalationPolicy *EscalationPolicy `json:"escalation_policy,omitempty"`
	EscalationLevel  int               `json:"escalation_level"`
	Start            string            `json:"start"`
	End              string            `json:"end"`
}

type EscalationPolicy struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NumLoops    int    `json:"num_loops"`
}

type ListResponse struct {
	Limit  int  `json:"limit"`
	Offset int  `json:"offset"`
	More   bool `json:"more"`
	Total  int  `json:"total"`
}

type SchedulesResponse struct {
	ListResponse
	Schedules []Schedule `json:"schedules"`
}

type OnCallsResponse struct {
	ListResponse
	OnCalls []OnCall `json:"oncalls"`
}

type ErrorResponse struct {
	Error struct {
		Message string   `json:"message"`
		Code    int      `json:"code"`
		Errors  []string `json:"errors"`
	} `json:"error"`
}

// ScheduleResponse wraps a single schedule with details
type ScheduleResponse struct {
	Schedule ScheduleDetail `json:"schedule"`
}

// ScheduleDetail extends Schedule with additional fields for single schedule response
type ScheduleDetail struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	TimeZone         string            `json:"time_zone"`
	Summary          string            `json:"summary"`
	ScheduleLayers   []ScheduleLayer   `json:"schedule_layers,omitempty"`
	OverrideSubcycle *OverrideSubcycle `json:"override_subcycle,omitempty"`
	FinalSchedule    *FinalSchedule    `json:"final_schedule,omitempty"`
}

// RenderedScheduleEntry represents a schedule entry with user details
type RenderedScheduleEntry struct {
	User  User   `json:"user"`
	Start string `json:"start"`
	End   string `json:"end"`
}

// Service represents a PagerDuty service
type Service struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Summary     string `json:"summary"`
	Status      string `json:"status"`
}

// ServicesResponse wraps the services list response
type ServicesResponse struct {
	ListResponse
	Services []Service `json:"services"`
}

// ServiceReference represents a reference to a service
type ServiceReference struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// AssigneeReference represents a reference to an assignee
type AssigneeReference struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// Assignment represents an incident assignment
type Assignment struct {
	Assignee AssigneeReference `json:"assignee"`
}

// Incident represents a PagerDuty incident
type Incident struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Title        string         `json:"title"`
	Description  string         `json:"description,omitempty"`
	Service      ServiceReference `json:"service"`
	Assignments  []Assignment   `json:"assignments,omitempty"`
	Status       string         `json:"status,omitempty"`
	CreatedAt    string         `json:"created_at,omitempty"`
	IncidentKey  string         `json:"incident_key,omitempty"`
	HtmlURL      string         `json:"html_url,omitempty"`
}

// CreateIncidentRequest represents the request to create an incident
type CreateIncidentRequest struct {
	Incident Incident `json:"incident"`
}

// CreateIncidentResponse wraps the incident creation response
type CreateIncidentResponse struct {
	Incident Incident `json:"incident"`
}
