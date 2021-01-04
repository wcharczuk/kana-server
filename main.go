package main

import (
	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/controller"
	"github.com/wcharczuk/kana-server/pkg/local"
)

func main() {
	var cfg config.Config
	configutil.MustRead(&cfg)

	log := logger.MustNew(
		logger.OptConfig(cfg.Logger),
		logger.OptAll(),
	)
	server := web.MustNew(
		web.OptConfig(cfg.Web),
		web.OptLog(log),
	)

	model := new(local.Model)

	server.Register(
		controller.Index{Config: cfg},
		controller.Quiz{Model: model},
	)

	if err := graceful.Shutdown(server); err != nil {
		logger.MaybeFatalExit(log, err)
	}
}
