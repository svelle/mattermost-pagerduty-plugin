package command

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mattermost-pagerduty-plugin/server/pagerduty"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type Handler struct {
	client      *pluginapi.Client
	pluginAPI   PluginAPI
	siteURL     string
}

type Command interface {
	Handle(args *model.CommandArgs) (*model.CommandResponse, error)
}

type PluginAPI interface {
	GetSchedules() ([]byte, error)
	GetOnCalls() ([]byte, error)
}

const (
	commandTrigger = "pagerduty"
	
	subcommandHelp      = "help"
	subcommandSchedules = "schedules"
	subcommandOnCall    = "oncall"
)

// NewCommandHandler creates a new command handler
func NewCommandHandler(client *pluginapi.Client, pluginAPI PluginAPI, siteURL string) Command {
	handler := &Handler{
		client:    client,
		pluginAPI: pluginAPI,
		siteURL:   siteURL,
	}

	err := client.SlashCommand.Register(&model.Command{
		Trigger:          commandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "PagerDuty integration",
		AutoCompleteHint: "[help|schedules|oncall]",
		AutocompleteData: handler.getAutocompleteData(),
	})
	if err != nil {
		client.Log.Error("Failed to register command", "error", err)
	}
	return handler
}

func (h *Handler) getAutocompleteData() *model.AutocompleteData {
	data := model.NewAutocompleteData(commandTrigger, "[subcommand]", "Available subcommands: help, schedules, oncall")

	help := model.NewAutocompleteData(subcommandHelp, "", "Shows this help message")
	data.AddCommand(help)

	schedules := model.NewAutocompleteData(subcommandSchedules, "", "List all PagerDuty schedules")
	data.AddCommand(schedules)

	oncall := model.NewAutocompleteData(subcommandOnCall, "", "Show who's currently on-call")
	data.AddCommand(oncall)

	return data
}

// Handle executes the command
func (h *Handler) Handle(args *model.CommandArgs) (*model.CommandResponse, error) {
	fields := strings.Fields(args.Command)
	if len(fields) < 2 {
		return h.executeHelpCommand(args), nil
	}

	subcommand := fields[1]
	switch subcommand {
	case subcommandHelp:
		return h.executeHelpCommand(args), nil
	case subcommandSchedules:
		return h.executeSchedulesCommand(args), nil
	case subcommandOnCall:
		return h.executeOnCallCommand(args), nil
	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown subcommand: %s\n\n%s", subcommand, h.getHelpText()),
		}, nil
	}
}

func (h *Handler) executeHelpCommand(args *model.CommandArgs) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         h.getHelpText(),
	}
}

func (h *Handler) getHelpText() string {
	return `### PagerDuty Plugin Commands

* **/pagerduty help** - Show this help message
* **/pagerduty schedules** - List all PagerDuty schedules
* **/pagerduty oncall** - Show who's currently on-call

You can also click the PagerDuty icon in the channel header to open the sidebar view for more detailed information.`
}

func (h *Handler) executeSchedulesCommand(args *model.CommandArgs) *model.CommandResponse {
	data, err := h.pluginAPI.GetSchedules()
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Failed to retrieve schedules: %v", err),
		}
	}

	var response pagerduty.SchedulesResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Failed to parse schedule data",
		}
	}

	if len(response.Schedules) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "No PagerDuty schedules found.",
		}
	}

	// Build the response text
	var text strings.Builder
	text.WriteString("### PagerDuty Schedules\n\n")
	
	for _, schedule := range response.Schedules {
		text.WriteString(fmt.Sprintf("**%s**", schedule.Name))
		if schedule.Description != "" {
			text.WriteString(fmt.Sprintf(" - %s", schedule.Description))
		}
		text.WriteString(fmt.Sprintf("\n_Timezone: %s_\n\n", schedule.TimeZone))
	}

	text.WriteString("\n_Use `/pagerduty oncall` to see who's currently on-call, or click the PagerDuty icon in the channel header for detailed schedule views._")

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         text.String(),
	}
}

func (h *Handler) executeOnCallCommand(args *model.CommandArgs) *model.CommandResponse {
	data, err := h.pluginAPI.GetOnCalls()
	if err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Failed to retrieve on-call users: %v", err),
		}
	}

	var response pagerduty.OnCallsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Failed to parse on-call data",
		}
	}

	if len(response.OnCalls) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "No one is currently on-call.",
		}
	}

	// Group on-calls by schedule
	oncallsBySchedule := make(map[string][]pagerduty.OnCall)
	for _, oncall := range response.OnCalls {
		scheduleName := "Unknown Schedule"
		if oncall.Schedule != nil {
			scheduleName = oncall.Schedule.Name
		}
		oncallsBySchedule[scheduleName] = append(oncallsBySchedule[scheduleName], oncall)
	}

	// Sort schedule names for consistent output
	var scheduleNames []string
	for name := range oncallsBySchedule {
		scheduleNames = append(scheduleNames, name)
	}
	sort.Strings(scheduleNames)

	// Build the response text
	var text strings.Builder
	text.WriteString("### Currently On-Call\n\n")

	for _, scheduleName := range scheduleNames {
		oncalls := oncallsBySchedule[scheduleName]
		text.WriteString(fmt.Sprintf("**%s**\n", scheduleName))
		
		for _, oncall := range oncalls {
			text.WriteString(fmt.Sprintf("• %s", oncall.User.Name))
			if oncall.User.Email != "" {
				text.WriteString(fmt.Sprintf(" (%s)", oncall.User.Email))
			}
			
			// Show when their shift ends if available
			if oncall.End != nil {
				endTime := oncall.End.Local()
				now := time.Now()
				
				if endTime.Day() == now.Day() {
					text.WriteString(fmt.Sprintf(" - until %s today", endTime.Format("3:04 PM")))
				} else if endTime.Day() == now.Add(24*time.Hour).Day() {
					text.WriteString(fmt.Sprintf(" - until %s tomorrow", endTime.Format("3:04 PM")))
				} else {
					text.WriteString(fmt.Sprintf(" - until %s", endTime.Format("Mon 3:04 PM")))
				}
			}
			
			if oncall.EscalationLevel > 0 {
				text.WriteString(fmt.Sprintf(" _(escalation level %d)_", oncall.EscalationLevel))
			}
			text.WriteString("\n")
		}
		text.WriteString("\n")
	}

	text.WriteString("_Click the PagerDuty icon in the channel header to see the full 24-hour schedule timeline._")

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         text.String(),
	}
}