package controller

import (
	"net/http"
	"time"

	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/kana"
	"github.com/wcharczuk/kana-server/pkg/model"
	"github.com/wcharczuk/kana-server/pkg/types"
)

// Quiz is the quiz controller.
type Quiz struct {
	BaseController
	Config config.Config
	Model  model.Manager
}

// Register adds the controller methods to the app.
func (q Quiz) Register(app *web.App) {
	app.Views.AddPaths(
		"_views/quiz.html",
		"_views/quiz_new.html",
	)

	app.GET("/quiz.new", q.getQuizNew, web.SessionRequired)
	app.POST("/quiz.new", q.postQuizNew, web.SessionRequired)
	app.GET("/quiz/:id", q.getQuizPrompt, web.SessionRequired)
	app.POST("/quiz/:id/answer", q.postQuizAnswer, web.SessionRequired)
}

// GET /quiz.new
func (q Quiz) getQuizNew(r *web.Ctx) web.Result {
	return r.Views.View("quiz_new", nil)
}

// POST /quiz.new
func (q Quiz) postQuizNew(r *web.Ctx) web.Result {
	userID, err := q.getUserID(r)
	if err != nil {
		return r.Views.InternalError(err)
	}

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
		UserID:           userID,
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
func (q Quiz) getQuizPrompt(r *web.Ctx) web.Result {
	userID, err := q.getUserID(r)
	if err != nil {
		return r.Views.InternalError(err)
	}
	quizID, err := web.UUIDValue(r.Param("id"))
	if err != nil {
		return r.Views.BadRequest(err)
	}
	quiz, found, err := q.Model.GetQuiz(r.Context(), quizID)
	if err != nil {
		return r.Views.InternalError(err)
	}
	if !found || !quiz.UserID.Equal(userID) {
		return r.Views.NotFound()
	}

	// filter out the prompts (and weights)
	// for which we have recent history
	nonQueried := make(map[string]string)
	for key, value := range quiz.Prompts {
		nonQueried[key] = value
	}
	nonQueriedWeights := make(map[string]float64)
	for key, value := range quiz.PromptWeights {
		nonQueriedWeights[key] = value
	}
	for _, queried := range quiz.PromptHistory {
		delete(nonQueried, queried)
		delete(nonQueriedWeights, queried)
	}

	prompt, expected := kana.SelectWeighted(nonQueried, nonQueriedWeights)
	return r.Views.View("quiz", types.QuizPrompt{
		Quiz:       quiz,
		CreatedUTC: time.Now().UTC(),
		Prompt:     prompt,
		Expected:   expected,
	})
}

// POST /quiz/:id/answer
func (q Quiz) postQuizAnswer(r *web.Ctx) web.Result {
	userID, err := q.getUserID(r)
	if err != nil {
		return r.Views.InternalError(err)
	}
	quizID, err := web.UUIDValue(r.Param("id"))
	if err != nil {
		return r.Views.BadRequest(err)
	}
	quiz, found, err := q.Model.GetQuiz(r.Context(), quizID)
	if err != nil {
		return r.Views.InternalError(err)
	}
	if !found || !quiz.UserID.Equal(userID) {
		return r.Views.NotFound()
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
		UserID:      userID,
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
	quiz.LastAnsweredUTC = time.Now().UTC()
	quiz.PromptHistory = kana.ListAddFixedLength(quiz.PromptHistory, prompt, quiz.MaxRepeatHistory)
	if err := q.Model.UpdateQuiz(r.Context(), quiz); err != nil {
		return r.Views.InternalError(err)
	}
	if err := q.Model.AddQuizResult(r.Context(), quizResult); err != nil {
		return r.Views.InternalError(err)
	}

	if quiz.MaxQuestions > 0 {
		if len(quiz.Results)+1 >= quiz.MaxQuestions {
			return web.RedirectWithMethodf(http.MethodGet, "/stats/%s", quiz.ID.String())
		}
	}

	return web.RedirectWithMethodf(http.MethodGet, "/quiz/%s", quiz.ID.String())
}
