package types

import (
	"time"

	"github.com/blend/go-sdk/mathutil"
	"github.com/blend/go-sdk/uuid"
)

// Quiz is a quiz.
type Quiz struct {
	// ID is a unique identifier for the quiz.
	ID uuid.UUID
	// CreatedUTC is the time the quiz was created.
	CreatedUTC time.Time
	// Hiragana indicates if we should include prompts from the hiragana set.
	Hiragana bool
	// Katakana indicates if we should include prompts from the katakana set.
	Katakana bool
	// MaxPrompts is the maximum number of prompts to pull from either prompt set (or in total)
	MaxPrompts int
	// MaxQuestions is the maximum number of questions to ask per quiz.
	MaxQuestions int
	// MaxRepeatHistory is the debounce history list length.
	MaxRepeatHistory int
	// Results are the individual prompts and answers.
	Results []QuizResult
	// Prompts are the individual mappings between kana and roman to quiz.
	Prompts map[string]string
	// PromptWeights are used for selection bias based on incorrect answers.
	PromptWeights map[string]float64
	// PromptHistory are the recent prompts used to debounce them.
	PromptHistory []string
}

// LatestResult returns the latest result.
func (q Quiz) LatestResult() *QuizResult {
	if len(q.Results) > 0 {
		return &q.Results[len(q.Results)-1]
	}
	return nil
}

// ElapsedTimes returns the elapsed time aggregates.
func (q Quiz) Stats() (stats QuizStats) {
	var elapsedTimes []time.Duration

	for _, qr := range q.Results {
		elapsedTimes = append(elapsedTimes, qr.Elapsed())
		if qr.Correct() {
			stats.Correct++
		}
		stats.Total++
	}

	stats.ElapsedAverage = mathutil.MeanDurations(elapsedTimes)
	stats.ElapsedP90 = mathutil.PercentileOfDuration(elapsedTimes, 0.90)
	stats.ElapsedP95 = mathutil.PercentileOfDuration(elapsedTimes, 0.95)
	stats.ElapsedMin, stats.ElapsedMax = mathutil.MinMaxDurations(elapsedTimes)
	return
}
