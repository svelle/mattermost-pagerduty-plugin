package pagerduty

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func newMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		apiToken    string
		baseURL     string
		wantBaseURL string
	}{
		{
			name:        "with custom base URL",
			apiToken:    "test-token",
			baseURL:     "https://custom.pagerduty.com",
			wantBaseURL: "https://custom.pagerduty.com",
		},
		{
			name:        "with empty base URL uses default",
			apiToken:    "test-token",
			baseURL:     "",
			wantBaseURL: defaultBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiToken, tt.baseURL)
			assert.NotNil(t, client)
			assert.Equal(t, tt.apiToken, client.apiToken)
			assert.Equal(t, tt.wantBaseURL, client.baseURL)
			assert.NotNil(t, client.httpClient)
		})
	}
}

func TestClient_doRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		params      url.Values
		mockFunc    func(req *http.Request) (*http.Response, error)
		wantBody    string
		wantErr     bool
		errContains string
	}{
		{
			name:   "successful GET request",
			method: "GET",
			path:   "/api/v1/schedules",
			params: url.Values{"limit": []string{"10"}},
			mockFunc: func(req *http.Request) (*http.Response, error) {
				// Verify request
				assert.Equal(t, "GET", req.Method)
				assert.Equal(t, "https://api.pagerduty.com/api/v1/schedules?limit=10", req.URL.String())
				assert.Equal(t, "application/vnd.pagerduty+json;version=2", req.Header.Get("Accept"))
				assert.Equal(t, "Token token=test-token", req.Header.Get("Authorization"))

				return newMockResponse(200, `{"schedules": []}`), nil
			},
			wantBody: `{"schedules": []}`,
			wantErr:  false,
		},
		{
			name:   "handles 401 unauthorized",
			method: "GET",
			path:   "/api/v1/schedules",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse(401, `{"error": {"message": "Unauthorized"}}`), nil
			},
			wantErr:     true,
			errContains: "Unauthorized",
		},
		{
			name:   "handles network error",
			method: "GET",
			path:   "/api/v1/schedules",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("network error")
			},
			wantErr:     true,
			errContains: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				baseURL:  "https://api.pagerduty.com",
				apiToken: "test-token",
				httpClient: &mockHTTPClient{
					doFunc: tt.mockFunc,
				},
			}

			body, err := client.doRequest(tt.method, tt.path, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestClient_GetSchedules(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		offset   int
		mockFunc func(req *http.Request) (*http.Response, error)
		want     *SchedulesResponse
		wantErr  bool
	}{
		{
			name:   "successful response",
			limit:  25,
			offset: 0,
			mockFunc: func(req *http.Request) (*http.Response, error) {
				// Verify query parameters
				assert.Equal(t, "25", req.URL.Query().Get("limit"))
				assert.Equal(t, "0", req.URL.Query().Get("offset"))

				return newMockResponse(200, `{
					"schedules": [
						{
							"id": "SCHED1",
							"name": "Primary On-Call",
							"time_zone": "America/New_York"
						}
					],
					"limit": 25,
					"offset": 0,
					"total": 1
				}`), nil
			},
			want: &SchedulesResponse{
				Schedules: []Schedule{
					{
						ID:       "SCHED1",
						Name:     "Primary On-Call",
						TimeZone: "America/New_York",
					},
				},
				ListResponse: ListResponse{
					Limit:  25,
					Offset: 0,
					Total:  1,
				},
			},
			wantErr: false,
		},
		{
			name:   "API error",
			limit:  10,
			offset: 0,
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse(500, `{"error": {"message": "Internal server error"}}`), nil
			},
			wantErr: true,
		},
		{
			name:   "invalid JSON response",
			limit:  10,
			offset: 0,
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse(200, `{invalid json`), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				baseURL:  "https://api.pagerduty.com",
				apiToken: "test-token",
				httpClient: &mockHTTPClient{
					doFunc: tt.mockFunc,
				},
			}

			got, err := client.GetSchedules(tt.limit, tt.offset)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClient_GetOnCalls(t *testing.T) {
	tests := []struct {
		name     string
		params   url.Values
		mockFunc func(req *http.Request) (*http.Response, error)
		want     *OnCallsResponse
		wantErr  bool
	}{
		{
			name: "successful response with parameters",
			params: url.Values{
				"schedule_ids[]": []string{"SCHED1", "SCHED2"},
				"since":          []string{"2024-01-01T00:00:00Z"},
				"until":          []string{"2024-01-02T00:00:00Z"},
			},
			mockFunc: func(req *http.Request) (*http.Response, error) {
				// Verify query parameters
				query := req.URL.Query()
				assert.Equal(t, []string{"SCHED1", "SCHED2"}, query["schedule_ids[]"])
				assert.Equal(t, "2024-01-01T00:00:00Z", query.Get("since"))
				assert.Equal(t, "2024-01-02T00:00:00Z", query.Get("until"))

				return newMockResponse(200, `{
					"oncalls": [
						{
							"user": {
								"id": "USER1",
								"name": "John Doe",
								"email": "john@example.com"
							},
							"schedule": {
								"id": "SCHED1",
								"name": "Primary On-Call"
							},
							"escalation_level": 1,
							"start": "2024-01-01T00:00:00Z",
							"end": "2024-01-02T00:00:00Z"
						}
					]
				}`), nil
			},
			want: &OnCallsResponse{
				OnCalls: []OnCall{
					{
						User: User{
							ID:    "USER1",
							Name:  "John Doe",
							Email: "john@example.com",
						},
						Schedule: Schedule{
							ID:   "SCHED1",
							Name: "Primary On-Call",
						},
						EscalationLevel: 1,
						Start:           "2024-01-01T00:00:00Z",
						End:             "2024-01-02T00:00:00Z",
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "nil parameters",
			params: nil,
			mockFunc: func(req *http.Request) (*http.Response, error) {
				// Should have empty query
				assert.Empty(t, req.URL.Query())
				return newMockResponse(200, `{"oncalls": []}`), nil
			},
			want: &OnCallsResponse{
				OnCalls: []OnCall{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				baseURL:  "https://api.pagerduty.com",
				apiToken: "test-token",
				httpClient: &mockHTTPClient{
					doFunc: tt.mockFunc,
				},
			}

			got, err := client.GetOnCalls(tt.params)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClient_GetSchedule(t *testing.T) {
	tests := []struct {
		name       string
		scheduleID string
		since      time.Time
		until      time.Time
		mockFunc   func(req *http.Request) (*http.Response, error)
		want       *ScheduleResponse
		wantErr    bool
	}{
		{
			name:       "successful response",
			scheduleID: "SCHED1",
			since:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			until:      time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			mockFunc: func(req *http.Request) (*http.Response, error) {
				// Verify path and parameters
				assert.Contains(t, req.URL.Path, "SCHED1")
				assert.Equal(t, "2024-01-01T00:00:00Z", req.URL.Query().Get("since"))
				assert.Equal(t, "2024-01-02T00:00:00Z", req.URL.Query().Get("until"))

				return newMockResponse(200, `{
					"schedule": {
						"id": "SCHED1",
						"name": "Primary On-Call",
						"time_zone": "America/New_York",
						"final_schedule": {
							"rendered_schedule_entries": [
								{
									"start": "2024-01-01T00:00:00Z",
									"end": "2024-01-02T00:00:00Z",
									"user": {
										"id": "USER1",
										"name": "John Doe"
									}
								}
							]
						}
					}
				}`), nil
			},
			want: &ScheduleResponse{
				Schedule: ScheduleDetail{
					ID:       "SCHED1",
					Name:     "Primary On-Call",
					TimeZone: "America/New_York",
					FinalSchedule: &FinalSchedule{
						RenderedScheduleEntries: []RenderedScheduleEntry{
							{
								Start: "2024-01-01T00:00:00Z",
								End:   "2024-01-02T00:00:00Z",
								User: User{
									ID:   "USER1",
									Name: "John Doe",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			scheduleID: "INVALID",
			mockFunc: func(req *http.Request) (*http.Response, error) {
				return newMockResponse(404, `{"error": {"message": "Schedule not found"}}`), nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				baseURL:  "https://api.pagerduty.com",
				apiToken: "test-token",
				httpClient: &mockHTTPClient{
					doFunc: tt.mockFunc,
				},
			}

			got, err := client.GetSchedule(tt.scheduleID, tt.since, tt.until)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClient_GetCurrentOnCalls(t *testing.T) {
	client := &Client{
		baseURL:  "https://api.pagerduty.com",
		apiToken: "test-token",
		httpClient: &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				// Verify the convenience method sets correct parameters
				query := req.URL.Query()
				assert.Equal(t, "UTC", query.Get("time_zone"))
				assert.Equal(t, []string{"users", "schedules"}, query["include[]"])
				assert.Equal(t, "true", query.Get("earliest"))

				return newMockResponse(200, `{
					"oncalls": [{
						"user": {"id": "USER1", "name": "John Doe"},
						"schedule": {"id": "SCHED1", "name": "Primary"},
						"escalation_level": 1,
						"start": "2024-01-01T00:00:00Z",
						"end": "2024-01-02T00:00:00Z"
					}]
				}`), nil
			},
		},
	}

	response, err := client.GetCurrentOnCalls()
	require.NoError(t, err)
	assert.Len(t, response.OnCalls, 1)
	assert.Equal(t, "USER1", response.OnCalls[0].User.ID)
}

func TestClient_GetOnCallsForSchedule(t *testing.T) {
	scheduleID := "SCHED123"
	client := &Client{
		baseURL:  "https://api.pagerduty.com",
		apiToken: "test-token",
		httpClient: &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				// Verify the convenience method sets correct parameters
				query := req.URL.Query()
				assert.Equal(t, []string{scheduleID}, query["schedule_ids[]"])
				assert.Equal(t, []string{"users"}, query["include[]"])
				assert.Equal(t, "true", query.Get("earliest"))

				return newMockResponse(200, `{"oncalls": []}`), nil
			},
		},
	}

	response, err := client.GetOnCallsForSchedule(scheduleID)
	require.NoError(t, err)
	assert.NotNil(t, response)
}

// Test the actual HTTP client interface
func TestClient_HTTPClientInterface(t *testing.T) {
	// Ensure our mock implements the same interface as http.Client
	var _ HTTPClient = &http.Client{}
	var _ HTTPClient = &mockHTTPClient{}
}
