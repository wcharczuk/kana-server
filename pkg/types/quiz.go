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
	// Results are the individual prompts and answers.
	Results []QuizResult
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
