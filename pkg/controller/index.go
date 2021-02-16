package controller

import (
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/config"
)

// Index is the root controller.
type Index struct {
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

	app.GET("/", i.index)
	app.GET("/status", i.status)
}

func (i Index) index(r *web.Ctx) web.Result {
	return r.Views.View("index", nil)
}

func (i Index) status(r *web.Ctx) web.Result {
	return web.Text.Result("OK!")
}
