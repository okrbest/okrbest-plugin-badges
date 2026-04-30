package sqlstore

import (
	"database/sql"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

type StoreAPI interface {
	GetMasterDB() (*sql.DB, error)
	DriverName() string
}

type PluginAPIClient struct {
	Store StoreAPI
	API   plugin.API
}

func NewClient(client *pluginapi.Client, api plugin.API) PluginAPIClient {
	return PluginAPIClient{
		Store: client.Store,
		API:   api,
	}
}
