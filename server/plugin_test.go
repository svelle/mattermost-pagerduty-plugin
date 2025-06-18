package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/svelle/mattermost-pagerduty-plugin/server/pagerduty"
)

func TestPlugin_OnActivate(t *testing.T) {
	// Test successful activation
	t.Run("successful activation", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		siteURL := "http://localhost:8065"
		api.On("GetConfig").Return(&model.Config{
			ServiceSettings: model.ServiceSettings{
				SiteURL: &siteURL,
			},
		})
		// Capture all log calls
		api.On("LogInfo", mock.Anything).Return().Maybe()
		api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogError", mock.Anything).Return().Maybe()

		plugin := &Plugin{}
		plugin.SetAPI(api)

		err := plugin.OnActivate()

		require.NoError(t, err)
		assert.NotNil(t, plugin.client)
		assert.NotNil(t, plugin.kvstore)
		assert.NotNil(t, plugin.createPagerDutyClient)
	})

	// Test missing site URL
	t.Run("missing site URL", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		api.On("GetConfig").Return(&model.Config{
			ServiceSettings: model.ServiceSettings{
				SiteURL: nil,
			},
		})
		api.On("LogInfo", mock.Anything).Return()
		api.On("LogError", mock.Anything).Return()

		plugin := &Plugin{}
		plugin.SetAPI(api)

		err := plugin.OnActivate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "site URL is not configured")
	})
}

func TestPlugin_OnDeactivate(t *testing.T) {
	api := &plugintest.API{}
	defer api.AssertExpectations(t)

	// Mock the log call
	api.On("LogInfo", mock.Anything).Return()

	plugin := &Plugin{}
	plugin.SetAPI(api)
	plugin.client = pluginapi.NewClient(api, nil)

	err := plugin.OnDeactivate()
	assert.NoError(t, err)
}

func TestPlugin_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		userID         string
		setupPlugin    func(*Plugin)
		expectedStatus int
	}{
		{
			name:           "missing user ID",
			method:         http.MethodGet,
			path:           "/api/v1/schedules",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "nonexistent endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/nonexistent",
			userID:         "test-user-id",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "schedules endpoint",
			method:         http.MethodGet,
			path:           "/api/v1/schedules",
			userID:         "test-user-id",
			expectedStatus: http.StatusNotImplemented, // Will fail due to missing config
			setupPlugin: func(p *Plugin) {
				p.configuration = &configuration{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &plugintest.API{}
			defer api.AssertExpectations(t)

			// Setup common mocks
			api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything).Maybe()
			api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything).Maybe()
			api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything).Maybe()
			api.On("LogError", mock.Anything, mock.Anything, mock.Anything).Maybe()

			plugin := &Plugin{}
			plugin.SetAPI(api)
			plugin.client = pluginapi.NewClient(api, nil)
			plugin.createPagerDutyClient = pagerduty.NewClient

			if tt.setupPlugin != nil {
				tt.setupPlugin(plugin)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.userID != "" {
				r.Header.Set("Mattermost-User-ID", tt.userID)
			}

			plugin.ServeHTTP(nil, w, r)

			result := w.Result()
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedStatus, result.StatusCode)
		})
	}
}

func TestPlugin_Configuration(t *testing.T) {
	t.Run("IsValid", func(t *testing.T) {
		tests := []struct {
			name    string
			config  configuration
			wantErr bool
		}{
			{
				name: "valid configuration",
				config: configuration{
					APIToken:   "test-token",
					APIBaseURL: "https://api.pagerduty.com",
				},
				wantErr: false,
			},
			{
				name: "missing API token",
				config: configuration{
					APIToken:   "",
					APIBaseURL: "https://api.pagerduty.com",
				},
				wantErr: true,
			},
			{
				name: "empty base URL uses default",
				config: configuration{
					APIToken:   "test-token",
					APIBaseURL: "",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.config.IsValid()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("setConfiguration", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.client = pluginapi.NewClient(api, nil)

		// Test setting configuration
		config := &configuration{
			APIToken:   "new-token",
			APIBaseURL: "https://new.pagerduty.com",
		}

		plugin.setConfiguration(config)

		// Verify configuration was set
		got := plugin.getConfiguration()
		assert.Equal(t, config, got)
	})

	t.Run("OnConfigurationChange", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		// Mock LoadPluginConfiguration
		api.On("LoadPluginConfiguration", mock.Anything).Run(func(args mock.Arguments) {
			config := args.Get(0).(*configuration)
			config.APIToken = "test-token"
			config.APIBaseURL = "https://api.pagerduty.com"
		}).Return(nil)

		plugin := &Plugin{}
		plugin.SetAPI(api)
		plugin.client = pluginapi.NewClient(api, nil)

		// Test configuration change
		err := plugin.OnConfigurationChange()
		assert.NoError(t, err)

		// Verify configuration was loaded
		config := plugin.getConfiguration()
		assert.Equal(t, "test-token", config.APIToken)
		assert.Equal(t, "https://api.pagerduty.com", config.APIBaseURL)
	})
}

func TestPlugin_Integration(t *testing.T) {
	t.Run("full plugin lifecycle", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)

		// Setup mocks for activation
		siteURL := "http://localhost:8065"
		api.On("GetConfig").Return(&model.Config{
			ServiceSettings: model.ServiceSettings{
				SiteURL: &siteURL,
			},
		})
		api.On("LogInfo", mock.Anything).Return().Maybe()
		api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogWarn", mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
		api.On("LogError", mock.Anything).Return().Maybe()

		plugin := &Plugin{}
		plugin.SetAPI(api)

		// Activate plugin
		err := plugin.OnActivate()
		require.NoError(t, err)

		// Verify plugin is properly initialized
		assert.NotNil(t, plugin.client)
		assert.NotNil(t, plugin.kvstore)
		assert.NotNil(t, plugin.kvstore)

		// Test configuration
		config := &configuration{
			APIToken:   "test-token",
			APIBaseURL: "https://api.pagerduty.com",
		}
		plugin.setConfiguration(config)

		// Deactivate plugin
		err = plugin.OnDeactivate()
		assert.NoError(t, err)
	})
}
