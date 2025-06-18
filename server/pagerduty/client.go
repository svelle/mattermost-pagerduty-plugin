package pagerduty

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultBaseURL = "https://api.pagerduty.com"
	apiVersion     = "2"
)

// HTTPClient interface for mocking in tests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	baseURL    string
	apiToken   string
	httpClient HTTPClient
}

func NewClient(apiToken, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, params url.Values) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse URL")
	}

	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Authorization", "Token token="+c.apiToken)
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version="+apiVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode >= 400 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return nil, fmt.Errorf("PagerDuty API error: %s (code: %d)", errorResp.Error.Message, errorResp.Error.Code)
		}
		return nil, fmt.Errorf("PagerDuty API error: HTTP %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) GetSchedules(limit, offset int) (*SchedulesResponse, error) {
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", offset))

	body, err := c.doRequest("GET", "/schedules", params)
	if err != nil {
		return nil, err
	}

	var response SchedulesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal schedules response")
	}

	return &response, nil
}

func (c *Client) GetSchedule(scheduleID string, since, until time.Time) (*ScheduleResponse, error) {
	params := url.Values{}
	params.Set("since", since.Format(time.RFC3339))
	params.Set("until", until.Format(time.RFC3339))

	body, err := c.doRequest("GET", fmt.Sprintf("/schedules/%s", scheduleID), params)
	if err != nil {
		return nil, err
	}

	var response ScheduleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal schedule response")
	}

	return &response, nil
}

func (c *Client) GetOnCalls(params url.Values) (*OnCallsResponse, error) {
	if params == nil {
		params = url.Values{}
	}

	body, err := c.doRequest("GET", "/oncalls", params)
	if err != nil {
		return nil, err
	}

	var response OnCallsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal oncalls response")
	}

	return &response, nil
}

func (c *Client) GetCurrentOnCalls() (*OnCallsResponse, error) {
	params := url.Values{}
	params.Set("time_zone", "UTC")
	params.Add("include[]", "users")
	params.Add("include[]", "schedules")
	params.Set("earliest", "true")

	return c.GetOnCalls(params)
}

func (c *Client) GetOnCallsForSchedule(scheduleID string) (*OnCallsResponse, error) {
	params := url.Values{}
	params.Set("schedule_ids[]", scheduleID)
	params.Set("include[]", "users")
	params.Set("earliest", "true")

	return c.GetOnCalls(params)
}
