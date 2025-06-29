package main

import (
	"sync"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"

	"github.com/svelle/mattermost-pagerduty-plugin/server/pagerduty"
	"github.com/svelle/mattermost-pagerduty-plugin/server/store/kvstore"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// kvstore is the client used to read/write KV records for this plugin.
	kvstore kvstore.KVStore

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// createPagerDutyClient is a function to create PagerDuty clients.
	// This can be overridden in tests to inject mock clients.
	createPagerDutyClient func(apiToken, baseURL string) *pagerduty.Client
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.MattermostPlugin.API, p.MattermostPlugin.Driver)
	p.client.Log.Info("PagerDuty plugin activating")

	// Initialize the PagerDuty client factory with the default implementation
	p.createPagerDutyClient = pagerduty.NewClient

	p.kvstore = kvstore.NewKVStore(p.client)

	config := p.MattermostPlugin.API.GetConfig()
	if config.ServiceSettings.SiteURL == nil {
		p.client.Log.Error("Site URL is not configured")
		return errors.New("site URL is not configured")
	}
	siteURL := *config.ServiceSettings.SiteURL
	p.client.Log.Debug("Site URL configured", "url", siteURL)

	// Slash commands and bot removed - sidebar only

	// Log plugin configuration status
	pluginConfig := p.getConfiguration()
	if err := pluginConfig.IsValid(); err != nil {
		p.client.Log.Warn("Plugin configuration is not valid", "error", err)
	} else {
		p.client.Log.Info("Plugin configuration is valid", "base_url", pluginConfig.APIBaseURL)
	}

	p.client.Log.Info("PagerDuty plugin activated successfully")
	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	if p.client != nil {
		p.client.Log.Info("PagerDuty plugin deactivating")
	}
	return nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
