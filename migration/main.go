package main

import (
	"context"
	"flag"

	"github.com/blend/go-sdk/configutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/env"
	"github.com/blend/go-sdk/logger"

	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/model"
)

var (
	skipCreateDatabase = flag.Bool("skip-create-database", false, "if we should skip creating the database")
	skipMigrations     = flag.Bool("skip-migrations", false, "if we should skip migrations")
)

func main() {
	flag.Parse()

	var cfg config.Config
	configutil.MustRead(&cfg)

	log := logger.MustNew(
		logger.OptConfig(cfg.Logger),
		logger.OptAll(),
	)

	dbConfig, err := cfg.DB.Reparse()
	if err != nil {
		logger.MaybeFatalExit(log, err)
	}

	if !*skipCreateDatabase {
		if err := dbutil.CreateDatabaseIfNotExists(context.Background(), env.ServiceEnvDev, dbConfig.Database, db.OptLog(log)); err != nil {
			logger.MaybeFatalExit(log, err)
		}
	}

	if !*skipMigrations {
		conn, err := db.New(db.OptConfig(cfg.DB), db.OptLog(log))
		if err != nil {
			logger.MaybeFatalExit(log, err)
		}
		if err := conn.Open(); err != nil {
			logger.MaybeFatalExit(log, err)
		}

		suite := model.Migrations()
		suite.Log = log
		if err := suite.Apply(context.Background(), conn); err != nil {
			logger.MaybeFatalExit(log, err)
		}
	}
}
