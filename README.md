# Mattermost PagerDuty Plugin

[![Build Status](https://github.com/mattermost/mattermost-pagerduty-plugin/actions/workflows/ci.yml/badge.svg)](https://github.com/mattermost/mattermost-pagerduty-plugin/actions/workflows/ci.yml)

## Overview

The Mattermost PagerDuty Plugin integrates PagerDuty with Mattermost, allowing teams to view on-call schedules and current on-call users directly within Mattermost. This plugin provides quick access to PagerDuty information through a sidebar interface and slash commands.

## Features

### Core Functionality
- **Schedule Browser**: View all PagerDuty schedules in a clean, organized list
- **Timeline View**: Click any schedule to see a detailed 24-hour timeline showing:
  - Who's currently on-call (highlighted with special styling)
  - Upcoming shifts with countdown timers
  - Smooth transitions between on-call personnel
- **Right-Hand Sidebar**: Dedicated sidebar accessible via channel header button
- **Real-time Data**: Always shows current information - no background syncing needed
- **Secure Configuration**: API tokens are stored securely and never exposed in the UI

### User Interface
- **Intuitive Navigation**: Easy back button to switch between schedule list and details
- **Visual Indicators**: Current on-call person prominently displayed with colored background
- **Time Display**: Shows both time and date for clarity (e.g., "Tomorrow at 09:00")
- **Responsive Design**: Clean layout that works well in the Mattermost sidebar
- **Theme Support**: Automatically adapts to your Mattermost theme (light/dark)

### Slash Commands
- `/pagerduty help` - Display available commands and usage
- `/pagerduty schedules` - List all PagerDuty schedules directly in chat
- `/pagerduty oncall` - Show who's currently on-call with shift end times

### Configuration
- **PagerDuty API Token**: Secure token storage for API authentication
- **Custom API URL**: Support for self-hosted or regional PagerDuty instances

## Requirements

- Mattermost Server v6.2.1 or higher
- PagerDuty account with API access
- PagerDuty API token

## Installation

1. Download the latest plugin file from the [releases page](https://github.com/mattermost/mattermost-pagerduty-plugin/releases)
2. In Mattermost, go to **System Console > Plugins > Plugin Management**
3. Upload the plugin file
4. Enable the plugin

## Configuration

After installing the plugin, configure it in **System Console > Plugins > PagerDuty**:

1. **PagerDuty API Token**: Enter your PagerDuty API token
   - Generate a token in PagerDuty: **Configuration > API Access Keys**
   - Ensure the token has read access to schedules and users

2. **PagerDuty API Base URL**: (Optional) Customize if using a non-standard PagerDuty instance
   - Default: `https://api.pagerduty.com`

## Usage

### Opening the Sidebar

1. Look for the PagerDuty icon in the channel header (green icon with "P")
2. Click it to open the right-hand sidebar
3. The sidebar will load and display all your PagerDuty schedules

### Viewing Schedules

1. The main view shows all available schedules with:
   - Schedule name
   - Description (if available)
   - Timezone information
2. Click on any schedule to see detailed on-call information

### Timeline View

When you click on a schedule, you'll see:
- **Current On-Call**: Prominently displayed at the top with a blue background
- **Next 24 Hours**: A timeline showing all upcoming on-call transitions
- **Time Until Next**: Countdown showing when the next person goes on-call
- **Visual Timeline**: Color-coded entries with the current on-call highlighted

### Navigation

- Use the **← back arrow** to return to the schedule list
- Click **Refresh** to get the latest data
- Click the same schedule again to refresh its details

### Slash Commands

Quick commands available in any channel:

#### `/pagerduty help`
Shows available commands and usage instructions.

#### `/pagerduty schedules`
Lists all PagerDuty schedules with:
- Schedule names and descriptions
- Timezone information
- Quick overview of available schedules

#### `/pagerduty oncall`
Shows who's currently on-call with:
- Grouped by schedule
- User names and email addresses
- When their shift ends (e.g., "until 5:00 PM today")
- Escalation levels if applicable

All slash command responses are ephemeral (only visible to you) to avoid cluttering channels.

## Development

### Prerequisites

- Go 1.22 or higher
- Node.js 16 or higher
- npm 8 or higher

### Building the Plugin

1. Clone the repository:
   ```bash
   git clone https://github.com/mattermost/mattermost-pagerduty-plugin.git
   cd mattermost-pagerduty-plugin
   ```

2. Build the plugin:
   ```bash
   make
   ```

This will create the plugin file at `dist/com.mattermost.pagerduty-plugin.tar.gz`.

### Local Development

For local development with automatic deployment:

```bash
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_TOKEN=your-admin-token
make deploy
```

To watch for changes and auto-deploy:

```bash
make watch
```

## Future Enhancements

Here's a list of nice-to-have features that could enhance the PagerDuty plugin:

### 🔔 Notifications & Alerts
- **On-call transition notifications**: Notify users when they're about to go on-call (configurable advance notice)
- **Schedule change alerts**: Notify when someone's on-call schedule is modified
- **Incident notifications**: Real-time PagerDuty incident alerts in Mattermost channels
- **Override notifications**: Alert when schedule overrides are created

### 📅 Schedule Management
- **Schedule overrides**: Create temporary schedule overrides directly from Mattermost
- **Shift swapping**: Request and approve shift swaps between team members
- **Multi-schedule view**: View multiple schedules side-by-side for coordination
- **Calendar export**: Export on-call schedules to iCal/Google Calendar format
- **Historical view**: View past on-call schedules and coverage

### 🎯 Enhanced Features
- **User profiles**: Click on users to see their contact info and current status
- **Team filtering**: Filter schedules by team or service
- **Search functionality**: Search for specific users or schedules
- **Timezone support**: Show schedules in user's local timezone with conversion
- **Mobile optimization**: Responsive design for mobile Mattermost apps

### 🤖 Automation & Integration
- **Incident response**: Create Mattermost channels automatically for PagerDuty incidents
- **Status sync**: Sync on-call status to Mattermost user status
- **Escalation policies**: View and understand escalation policies
- **Service dependencies**: Visualize service dependencies and their on-call teams
- **Slack-style reminders**: Set reminders for on-call handoffs

### 📊 Analytics & Reporting
- **On-call metrics**: Time spent on-call, incident load per person
- **Coverage reports**: Identify gaps in on-call coverage
- **Rotation fairness**: Ensure equal distribution of on-call duties
- **Custom dashboards**: Build team-specific on-call dashboards

### 🔧 Administrative Features
- **Bulk configuration**: Configure multiple schedules at once
- **Role-based access**: Restrict who can view certain schedules
- **Audit logging**: Track who viewed or modified schedule information
- **Backup schedules**: Automatic backup of schedule configurations

### 🎨 User Experience
- **Dark mode optimization**: Better contrast and styling for dark themes
- **Customizable views**: Save preferred schedule views and filters
- **Keyboard shortcuts**: Navigate schedules quickly with keyboard commands
- **Rich schedule details**: Show more context like team descriptions, runbooks
- **Presence indicators**: Show if on-call person is online in Mattermost

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting PRs.

## Security

If you discover a security vulnerability, please email security@mattermost.com instead of using the issue tracker.

## License

This plugin is licensed under the [Apache License 2.0](LICENSE).