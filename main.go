package main

import (
	"context"
	"flag"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/graceful"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/web"
	"github.com/blend/go-sdk/webutil"

	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/controller"
	"github.com/wcharczuk/kana-server/pkg/model"
)

var (
	flagCreateDatabase bool
	flagMigrations     bool
	flagServer         bool
)

func init() {
	flag.BoolVar(&flagCreateDatabase, "create-database", false, "if we should creating the database")
	flag.BoolVar(&flagMigrations, "migrations", false, "if we should run migrations")
	flag.BoolVar(&flagServer, "server", true, "if we should run the server")
	flag.Parse()
}

func main() {
	var cfg config.Config
	configutil.MustRead(&cfg)

	log := logger.MustNew(
		logger.OptConfig(cfg.Logger),
		logger.OptAll(),
	)
	defer log.Close()

	dbConfig, err := cfg.DB.Reparse()
	if err != nil {
		logger.MaybeFatalExit(log, err)
	}

	if flagCreateDatabase {
		log.Infof("ensuring database exists: %s", dbConfig.Database)
		if err = dbutil.CreateDatabaseIfNotExists(context.Background(), env.ServiceEnvDev, dbConfig.Database, db.OptLog(log)); err != nil {
			logger.MaybeFatalExit(log, err)
		}
	} else {
		log.Debug("skipping ensuring database exists")
	}

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

	if flagMigrations {
		log.Info("running migrations")
		suite := model.Migrations()
		suite.Log = log
		if err = suite.Apply(context.Background(), conn); err != nil {
			logger.MaybeFatalExit(log, err)
		}
	} else {
		log.Debug("skipping migrations")
	}

	if flagServer {
		oauthMgr, err := oauth.New(
			oauth.OptConfig(cfg.OAuth),
		)
		if err != nil {
			logger.MaybeFatalExit(log, err)
		}
		app := web.MustNew(
			web.OptConfig(cfg.Web),
			web.OptLog(log),
		)
		modelMgr := model.Manager{
			BaseManager: dbutil.NewBaseManager(conn),
		}
		app.Register(
			controller.Index{Config: cfg},
			controller.Home{Config: cfg, Model: modelMgr},
			controller.Auth{Config: cfg, Model: modelMgr, OAuth: oauthMgr},
			controller.Quiz{Config: cfg, Model: modelMgr},
		)
		app.Views.LiveReload = !cfg.Meta.IsProdlike()
		if cfg.IsProdlike() {
			log.Info("using https upgrader")
			app.BaseMiddleware = append(app.BaseMiddleware, httpsUpgrade)
		}
		if err := graceful.Shutdown(app); err != nil {
			logger.MaybeFatalExit(log, err)
		}
	}
}

func httpsUpgrade(action web.Action) web.Action {
	return func(r *web.Ctx) web.Result {
		if r.Request.URL.Scheme == webutil.SchemeHTTP {
			webutil.HTTPSRedirectFunc(r.Response, r.Request)
			return nil
		}
		return action(r)
	}
}
