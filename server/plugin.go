package main

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/pkg/errors"

	"github.com/larkox/mattermost-plugin-badges/server/sqlstore"
)

type Plugin struct {
	plugin.MattermostPlugin

	configurationLock sync.RWMutex
	configuration     *configuration

	mm               *pluginapi.Client
	BotUserID        string
	store            Store
	sqlStore         *sqlstore.SQLStore
	router           *mux.Router
	badgeAdminUserID string
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) OnActivate() error {
	p.mm = pluginapi.NewClient(p.API, p.Driver)
	botID, err := p.mm.Bot.EnsureBot(&model.Bot{
		Username:    "badges",
		DisplayName: "Badges Bot",
		Description: "Created by the Badges plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to ensure badges bot")
	}
	p.BotUserID = botID

	apiClient := sqlstore.NewClient(p.mm, p.API)
	sqlStore, err := sqlstore.New(apiClient)
	if err != nil {
		return errors.Wrap(err, "failed creating the SQL store")
	}
	p.sqlStore = sqlStore

	mutex, err := cluster.NewMutex(p.API, "Badges_dbMutex")
	if err != nil {
		return errors.Wrap(err, "failed creating cluster mutex")
	}
	if err := func() error {
		mutex.Lock()
		defer mutex.Unlock()

		if err := sqlStore.RunMigrations(); err != nil {
			return errors.Wrap(err, "failed to run migrations")
		}
		if err := sqlStore.MigrateFromKV(p.API); err != nil {
			p.API.LogWarn("KV to DB migration encountered an error", "error", err.Error())
		}
		return nil
	}(); err != nil {
		return err
	}

	p.store = NewSQLStoreAdapter(sqlStore, p.API)
	p.initializeAPI()

	return p.mm.SlashCommand.Register(p.getCommand())
}
