## Development Memories

- `make deploy`: Command used to deploy the Mattermost PagerDuty plugin
- `make`: Build the plugin
- `make watch`: Watch for changes and auto-deploy

## Recent Development

### Paging Functionality (v1.0+)
- Added direct paging capability to page current on-call person from schedule view
- Implemented backend APIs for PagerDuty services and incident creation
- Created intuitive paging dialog with service selection and form validation
- Enhanced schedule UI with relative time display and visual indicators
- Reduced schedule display window from 7 days to 48 hours for better focus
- Added comprehensive CSS class names for future customization

### Architecture Notes
- Backend: Go HTTP handlers in `server/api_pagerduty.go` 
- Frontend: React components in `webapp/src/components/sidebar/`
- PagerDuty Client: Extended `server/pagerduty/client.go` with new endpoints
- Types: Added service and incident types in `server/pagerduty/types.go`

### UX Improvements
- Current on-call person highlighted with colored background and badge
- Relative time formatting ("2h 30m remaining", "Starts in 1d 4h")
- Visual separation between current and upcoming shifts
- One-click paging with contextual buttons
- Success feedback for incident creation