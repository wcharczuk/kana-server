package controller

import (
	"github.com/blend/go-sdk/web"
	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/model"
)

// Home is the home controller.
type Home struct {
	BaseController
	Config config.Config
	Model  model.Manager
}

// Register regisers the controller.
func (h Home) Register(app *web.App) {
	app.Views.AddPaths(
		"_views/home.html",
		"_views/home_quiz_stats.html",
		"_views/quiz_settings.html",
	)
	app.GET("/home", h.home, web.SessionRequired)
	app.GET("/home/:id", h.homeQuizStats, web.SessionRequired)
}

// GET /stats
func (h Home) home(r *web.Ctx) web.Result {
	userID, err := h.getUserID(r)
	if err != nil {
		return r.Views.InternalError(err)
	}
	all, err := h.Model.AllQuzzes(r.Context(), userID)
	if err != nil {
		return r.Views.InternalError(err)
	}
	return r.Views.View("home", CreateHomeViewModel(all))
}

// GET /home/:id
func (h Home) homeQuizStats(r *web.Ctx) web.Result {
	userID, err := h.getUserID(r)
	if err != nil {
		return r.Views.InternalError(err)
	}
	quizID, err := web.UUIDValue(r.RouteParam("id"))
	if err != nil {
		return r.Views.BadRequest(err)
	}
	quiz, found, err := h.Model.GetQuiz(r.Context(), quizID)
	if err != nil {
		return r.Views.InternalError(err)
	}
	if !found || !quiz.UserID.Equal(userID) {
		return r.Views.NotFound()
	}
	return r.Views.View("home_quiz_stats", quiz)
}
