package main

import (
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/logger"
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
	if err := conn.Open(); err != nil {
		logger.MaybeFatalExit(log, err)
	}
	log.Infof("using database dsn: %s", cfg.DB.CreateLoggingDSN())

	server := web.MustNew(
		web.OptConfig(cfg.Web),
		web.OptLog(log),
	)
	model := model.Manager{
		BaseManager: dbutil.NewBaseManager(conn),
	}
	server.Register(
		controller.Index{Config: cfg},
		controller.Quiz{Model: model},
	)
	// disable later
	server.Views.LiveReload = true
	if err := graceful.Shutdown(server); err != nil {
		logger.MaybeFatalExit(log, err)
	}
}
