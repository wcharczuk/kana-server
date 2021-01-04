package controller

import (
	"net/http"
	"time"

	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/interfaces"
	"github.com/wcharczuk/kana-server/pkg/types"
)

// Quiz is the quiz controller.
type Quiz struct {
	Model interfaces.Model
}

// Register adds the controller methods to the app.
func (q Quiz) Register(app *web.App) {
	app.Views.AddPaths(
		"_views/quiz.html",
		"_views/new_quiz.html",
	)
	app.GET("/quiz.new", q.getQuizNew)
	app.GET("/quiz/:id", q.getQuiz)
	app.POST("/quiz", q.postQuiz)
}

// GET /quiz.new
func (q Quiz) getQuizNew(r *web.Ctx) web.Result {
	return r.Views.View("new_quiz", nil)
}

// POST /quiz
func (q Quiz) postQuiz(r *web.Ctx) web.Result {
	maxQuestions, _ := web.IntValue(r.Param("maxQuestions"))
	maxPrompts, _ := web.IntValue(r.Param("maxPrompts"))
	includeHiragana, err := web.BoolValue(r.Param("hiragana"))
	if err != nil {
		return r.Views.BadRequest(err)
	}
	includeKatakana, err := web.BoolValue(r.Param("katakana"))
	if err != nil {
		return r.Views.BadRequest(err)
	}

	quiz := types.Quiz{
		ID:           uuid.V4(),
		CreatedUTC:   time.Now().UTC(),
		Hiragana:     includeHiragana,
		Katakana:     includeKatakana,
		MaxPrompts:   maxPrompts,
		MaxQuestions: maxQuestions,
	}
	if err := q.Model.CreateQuiz(r.Context(), quiz); err != nil {
		return r.Views.InternalError(err)
	}

	return web.RedirectWithMethodf(http.MethodGet, "/quiz/%s", quiz.ID.String())
}

// GET /quiz/:id
func (q Quiz) getQuiz(r *web.Ctx) web.Result {
	return web.NoContent
}
