package command

import (
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-pagerduty-plugin/server/pagerduty"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/stretchr/testify/assert"
)

type mockPluginAPI struct {
	schedules []byte
	oncalls   []byte
	err       error
}

func (m *mockPluginAPI) GetSchedules() ([]byte, error) {
	return m.schedules, m.err
}

func (m *mockPluginAPI) GetOnCalls() ([]byte, error) {
	return m.oncalls, m.err
}

type env struct {
	client    *pluginapi.Client
	api       *plugintest.API
	pluginAPI *mockPluginAPI
}

func setupTest() *env {
	api := &plugintest.API{}
	driver := &plugintest.Driver{}
	client := pluginapi.NewClient(api, driver)

	return &env{
		client:    client,
		api:       api,
		pluginAPI: &mockPluginAPI{},
	}
}

func TestPagerDutyCommand(t *testing.T) {
	assert := assert.New(t)
	env := setupTest()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock the command registration
	env.api.On("RegisterCommand", &model.Command{
		Trigger:          commandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "PagerDuty integration",
		AutoCompleteHint: "[help|schedules|oncall]",
		AutocompleteData: model.NewAutocompleteData(commandTrigger, "[subcommand]", "Available subcommands: help, schedules, oncall"),
	}).Return(nil).Maybe()

	cmdHandler := NewCommandHandler(env.client, env.pluginAPI, "http://localhost:8065")

	t.Run("help command", func(t *testing.T) {
		args := &model.CommandArgs{
			Command: "/pagerduty help",
		}
		response, err := cmdHandler.Handle(args)
		assert.Nil(err)
		assert.Contains(response.Text, "PagerDuty Plugin Commands")
	})

	t.Run("schedules command with data", func(t *testing.T) {
		schedulesResp := pagerduty.SchedulesResponse{
			Schedules: []pagerduty.Schedule{
				{
					ID:          "SCHED1",
					Name:        "Primary On-Call",
					Description: "Main support rotation",
					TimeZone:    "America/New_York",
				},
				{
					ID:       "SCHED2",
					Name:     "Backend On-Call",
					TimeZone: "America/Los_Angeles",
				},
			},
		}
		data, _ := json.Marshal(schedulesResp)
		env.pluginAPI.schedules = data
		
		args := &model.CommandArgs{
			Command: "/pagerduty schedules",
		}
		response, err := cmdHandler.Handle(args)
		assert.Nil(err)
		assert.Contains(response.Text, "PagerDuty Schedules")
		assert.Contains(response.Text, "Primary On-Call")
		assert.Contains(response.Text, "Main support rotation")
		assert.Contains(response.Text, "Backend On-Call")
		assert.Contains(response.Text, "America/New_York")
	})

	t.Run("schedules command with no data", func(t *testing.T) {
		schedulesResp := pagerduty.SchedulesResponse{
			Schedules: []pagerduty.Schedule{},
		}
		data, _ := json.Marshal(schedulesResp)
		env.pluginAPI.schedules = data
		
		args := &model.CommandArgs{
			Command: "/pagerduty schedules",
		}
		response, err := cmdHandler.Handle(args)
		assert.Nil(err)
		assert.Contains(response.Text, "No PagerDuty schedules found")
	})

	t.Run("oncall command with data", func(t *testing.T) {
		oncallsResp := pagerduty.OnCallsResponse{
			OnCalls: []pagerduty.OnCall{
				{
					User: pagerduty.User{
						ID:    "USER1",
						Name:  "John Doe",
						Email: "john@example.com",
					},
					Schedule: &pagerduty.Schedule{
						Name: "Primary On-Call",
					},
					EscalationLevel: 0,
				},
				{
					User: pagerduty.User{
						ID:   "USER2",
						Name: "Jane Smith",
					},
					Schedule: &pagerduty.Schedule{
						Name: "Backend On-Call",
					},
					EscalationLevel: 1,
				},
			},
		}
		data, _ := json.Marshal(oncallsResp)
		env.pluginAPI.oncalls = data
		
		args := &model.CommandArgs{
			Command: "/pagerduty oncall",
		}
		response, err := cmdHandler.Handle(args)
		assert.Nil(err)
		assert.Contains(response.Text, "Currently On-Call")
		assert.Contains(response.Text, "John Doe")
		assert.Contains(response.Text, "john@example.com")
		assert.Contains(response.Text, "Jane Smith")
		assert.Contains(response.Text, "Primary On-Call")
		assert.Contains(response.Text, "Backend On-Call")
		assert.Contains(response.Text, "escalation level 1")
	})

	t.Run("oncall command with no data", func(t *testing.T) {
		oncallsResp := pagerduty.OnCallsResponse{
			OnCalls: []pagerduty.OnCall{},
		}
		data, _ := json.Marshal(oncallsResp)
		env.pluginAPI.oncalls = data
		
		args := &model.CommandArgs{
			Command: "/pagerduty oncall",
		}
		response, err := cmdHandler.Handle(args)
		assert.Nil(err)
		assert.Contains(response.Text, "No one is currently on-call")
	})

	t.Run("unknown command", func(t *testing.T) {
		args := &model.CommandArgs{
			Command: "/pagerduty unknown",
		}
		response, err := cmdHandler.Handle(args)
		assert.Nil(err)
		assert.Contains(response.Text, "Unknown subcommand: unknown")
	})
}