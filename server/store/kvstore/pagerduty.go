package kvstore

import (
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"
)

// We expose our calls to the KVStore pluginapi methods through this interface for testability and stability.
// This allows us to better control which values are stored with which keys.

type Client struct {
	client *pluginapi.Client
}

func NewKVStore(client *pluginapi.Client) KVStore {
	return Client{
		client: client,
	}
}

// GetCachedSchedules retrieves cached schedule data
func (kv Client) GetCachedSchedules() ([]byte, error) {
	var data []byte
	err := kv.client.KV.Get("pagerduty_schedules_cache", &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cached schedules")
	}
	return data, nil
}

// SetCachedSchedules stores schedule data in cache
func (kv Client) SetCachedSchedules(data []byte) error {
	_, err := kv.client.KV.Set("pagerduty_schedules_cache", data)
	if err != nil {
		return errors.Wrap(err, "failed to cache schedules")
	}
	return nil
}
