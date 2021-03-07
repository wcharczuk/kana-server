package main

import (
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/controller"
	"github.com/wcharczuk/kana-server/pkg/model"
)

func main() {
	var cfg config.Config
	configutil.MustRead(&cfg)

	log := logger.MustNew(
		logger.OptConfig(cfg.Logger),
		logger.OptAll(),
	)
	defer log.Close()

	conn, err := db.New(
		db.OptConfig(cfg.DB),
		db.OptLog(log),
	)
	if err != nil {
		logger.MaybeFatalExit(log, err)
	}
	if err = conn.Open(); err != nil {
		logger.MaybeFatalExit(log, err)
	}
	log.Infof("using database dsn: %s", cfg.DB.CreateLoggingDSN())

	oauthMgr, err := oauth.New(
		oauth.OptConfig(cfg.OAuth),
	)
	if err != nil {
		logger.MaybeFatalExit(log, err)
	}

	server := web.MustNew(
		web.OptConfig(cfg.Web),
		web.OptLog(log),
	)
	modelMgr := model.Manager{
		BaseManager: dbutil.NewBaseManager(conn),
	}
	server.Register(
		controller.Index{Config: cfg},
		controller.Home{Config: cfg, Model: modelMgr},
		controller.Auth{Config: cfg, Model: modelMgr, OAuth: oauthMgr},
		controller.Quiz{Config: cfg, Model: modelMgr},
	)

	server.Views.LiveReload = !cfg.Meta.IsProdlike()
	if err := graceful.Shutdown(server); err != nil {
		logger.MaybeFatalExit(log, err)
	}
}
