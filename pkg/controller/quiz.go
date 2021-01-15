package controller

import (
	"net/http"
	"time"

	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/interfaces"
	"github.com/wcharczuk/kana-server/pkg/kana"
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
		"_views/stats.html",
	)
	app.GET("/quiz.new", q.getQuizNew)
	app.POST("/quiz", q.postQuiz)
	app.GET("/quiz/:id", q.getQuiz)
	app.POST("/quiz/:id/answer", q.postQuizAnswer)

	app.GET("/stats", q.getStats)
}

// GET /quiz.new
func (q Quiz) getStats(r *web.Ctx) web.Result {
	all, err := q.Model.All(r.Context())
	if err != nil {
		return r.Views.InternalError(err)
	}
	return r.Views.View("stats", all)
}

// GET /quiz.new
func (q Quiz) getQuizNew(r *web.Ctx) web.Result {
	return r.Views.View("new_quiz", nil)
}

// POST /quiz
func (q Quiz) postQuiz(r *web.Ctx) web.Result {
	maxQuestions, _ := web.IntValue(r.Param("maxQuestions"))
	maxPrompts, _ := web.IntValue(r.Param("maxPrompts"))
	maxRepeatHistory, _ := web.IntValue(r.Param("maxRepeatHistory"))

	includeHiragana, _ := r.Param("hiragana")
	includeKatakana, _ := r.Param("katakana")
	includeKatakanaWords, _ := r.Param("katakanaWords")

	var inputs []map[string]string
	if includeHiragana != "" {
		inputs = append(inputs, kana.Hiragana)
	}
	if includeKatakana != "" {
		inputs = append(inputs, kana.Katakana)
	}
	if includeKatakanaWords != "" {
		inputs = append(inputs, kana.KatakanaWords)
	}
	prompts := kana.SelectCount(kana.Merge(inputs...), maxPrompts)
	promptWeights := kana.CreateWeights(prompts)

	quiz := types.Quiz{
		ID:               uuid.V4(),
		CreatedUTC:       time.Now().UTC(),
		Hiragana:         includeHiragana != "",
		Katakana:         includeKatakana != "",
		KatakanaWords:    includeKatakanaWords != "",
		MaxPrompts:       maxPrompts,
		MaxQuestions:     maxQuestions,
		MaxRepeatHistory: maxRepeatHistory,
		Results:          nil,
		Prompts:          prompts,
		PromptWeights:    promptWeights,
		PromptHistory:    nil,
	}
	if err := q.Model.CreateQuiz(r.Context(), quiz); err != nil {
		return r.Views.InternalError(err)
	}
	return web.RedirectWithMethodf(http.MethodGet, "/quiz/%s", quiz.ID.String())
}

// GET /quiz/:id
func (q Quiz) getQuiz(r *web.Ctx) web.Result {
	quizID, err := UUIDValue(r.Param("id"))
	if err != nil {
		return r.Views.BadRequest(err)
	}
	quiz, err := q.Model.GetQuiz(r.Context(), quizID)
	if err != nil {
		return r.Views.InternalError(err)
	}

	prompt, expected := kana.SelectWeighted(quiz.Prompts, quiz.PromptWeights)
	for kana.ListHas(quiz.PromptHistory, prompt) {
		prompt, expected = kana.SelectWeighted(quiz.Prompts, quiz.PromptWeights)
	}
	return r.Views.View("quiz", types.QuizPrompt{
		Quiz:       quiz,
		CreatedUTC: time.Now().UTC(),
		Prompt:     prompt,
		Expected:   expected,
	})
}

// POST /quiz/:id/answer
func (q Quiz) postQuizAnswer(r *web.Ctx) web.Result {
	quizID, err := UUIDValue(r.Param("id"))
	if err != nil {
		return r.Views.BadRequest(err)
	}
	quiz, err := q.Model.GetQuiz(r.Context(), quizID)
	if err != nil {
		return r.Views.InternalError(err)
	}
	createdUTC, err := web.Int64Value(r.Param("createdUTC"))
	if err != nil {
		return r.Views.BadRequest(err)
	}
	prompt, err := r.Param("prompt")
	if err != nil {
		return r.Views.BadRequest(err)
	}
	expected, err := r.Param("expected")
	if err != nil {
		return r.Views.BadRequest(err)
	}
	actual, err := r.Param("actual")
	if err != nil {
		return r.Views.BadRequest(err)
	}

	quizResult := types.QuizResult{
		ID:          uuid.V4(),
		QuizID:      quiz.ID,
		CreatedUTC:  time.Unix(0, createdUTC).UTC(),
		AnsweredUTC: time.Now().UTC(),
		Prompt:      prompt,
		Expected:    expected,
		Actual:      actual,
	}
	if quizResult.Correct() {
		kana.DecreaseWeight(quiz.PromptWeights, prompt)
	} else {
		kana.IncreaseWeight(quiz.PromptWeights, prompt)
	}
	quiz.PromptHistory = kana.ListAddFixedLength(quiz.PromptHistory, prompt, quiz.MaxRepeatHistory)
	if err := q.Model.UpdateQuiz(r.Context(), quiz); err != nil {
		return r.Views.InternalError(err)
	}
	if err := q.Model.AddQuizResult(r.Context(), quizResult); err != nil {
		return r.Views.InternalError(err)
	}
	return web.RedirectWithMethodf(http.MethodGet, "/quiz/%s", quiz.ID.String())
}
