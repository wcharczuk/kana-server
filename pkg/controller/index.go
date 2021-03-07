package controller

import (
	"net/http"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/model"
)

// Index is the root controller.
type Index struct {
	BaseController
	Model  model.Manager
	Config config.Config
}

// Register regisers the controller.
func (i Index) Register(app *web.App) {
	app.Views.AddPaths(
		"_views/header.html",
		"_views/footer.html",
		"_views/index.html",
	)
	app.ServeStatic("/static", []string{"_static"})

	app.GET("/", i.index, web.SessionAware)
	app.GET("/status", i.status)
	app.GET("/status/postgres", i.statusPostgres)
}

func (i Index) index(r *web.Ctx) web.Result {
	if r.Session != nil {
		return web.Redirect("/home")
	}
	return r.Views.View("index", nil)
}

func (i Index) status(r *web.Ctx) web.Result {
	return web.JSON.Result(
		map[string]interface{}{
			"serviceEnv":  i.Config.ServiceEnvOrDefault(),
			"serviceName": i.Config.ServiceNameOrDefault(),
			"version":     i.Config.VersionOrDefault(),
			"gitRef":      i.Config.GitRef,
		},
	)
}

func (i Index) statusPostgres(r *web.Ctx) web.Result {
	any, err := i.Model.Invoke(r.Context(), db.OptLabel("status")).Query(`select 'ok!'`).Any()
	if err != nil || !any {
		return &web.JSONResult{
			StatusCode: http.StatusInternalServerError,
			Response:   map[string]interface{}{"status": false},
		}
	}
	dbStats := i.Model.Conn.Connection.Stats()
	return &web.JSONResult{
		StatusCode: http.StatusOK,
		Response: map[string]interface{}{
			"status":                 true,
			"db.conns_open":          dbStats.OpenConnections,
			"db.conns_idle":          dbStats.Idle,
			"db.max_open_conns":      dbStats.MaxOpenConnections,
			"db.conns_in_use":        dbStats.InUse,
			"db.wait_count":          dbStats.WaitCount,
			"db.wait_duration":       dbStats.WaitDuration,
			"db.max_idle_closed":     dbStats.MaxIdleClosed,
			"db.max_lifetime_closed": dbStats.MaxLifetimeClosed,
		},
	}
}
