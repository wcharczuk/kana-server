package types

import (
	"sort"
	"time"

	"github.com/blend/go-sdk/mathutil"
	"github.com/blend/go-sdk/uuid"

	"github.com/wcharczuk/kana-server/pkg/kana"
)

// Quiz is a quiz.
type Quiz struct {
	// ID is a unique identifier for the quiz.
	ID uuid.UUID `db:"id,pk"`
	// UserID is the user the quiz corresponds to.
	UserID uuid.UUID `db:"user_id"`
	// CreatedUTC is the time the quiz was created.
	CreatedUTC time.Time `db:"created_utc"`
	// LastAnsweredUTC is the last time the quiz was answered.
	LastAnsweredUTC time.Time `db:"last_answered_utc"`
	// Hiragana indicates if we should include prompts from the hiragana set.
	Hiragana bool `db:"hiragana"`
	// Katakana indicates if we should include prompts from the katakana set.
	Katakana bool `db:"katakana"`
	// KatakanaWords indicates if we should include prompts from the katakana words set.
	KatakanaWords bool `db:"katakana_words"`
	// MaxQuestions is the maximum number of questions to ask per quiz.
	MaxQuestions int `db:"max_questions"`
	// MaxPrompts is the maximum number of prompts to pull from either prompt set (or in total)
	MaxPrompts int `db:"max_prompts"`
	// MaxRepeatHistory is the debounce history list length.
	MaxRepeatHistory int `db:"max_repeat_history"`
	// Results are the individual prompts and answers.
	Results []QuizResult `db:"-"`
	// Prompts are the individual mappings between kana and roman to quiz.
	Prompts map[string]string `db:"prompts,json"`
	// PromptWeights are used for selection bias based on incorrect answers.
	PromptWeights map[string]float64 `db:"prompt_weights,json"`
	// PromptHistory are the recent prompts used to debounce them.
	PromptHistory []string `db:"prompt_history,json"`
}

// TableName returns the database tablename for the type.
func (q Quiz) TableName() string { return "quiz" }

// IsZero returns if the quiz is set or not.
func (q Quiz) IsZero() bool {
	return q.ID == nil || q.ID.IsZero()
}

// LatestResult returns the latest result.
func (q Quiz) LatestResult() *QuizResult {
	if len(q.Results) > 0 {
		return &q.Results[len(q.Results)-1]
	}
	return nil
}

// Stats returns the stats for the quiz.
func (q Quiz) Stats() (stats QuizStats) {
	var elapsedTimes []time.Duration

	for _, qr := range q.Results {
		elapsedTimes = append(elapsedTimes, qr.Elapsed())
		if qr.Correct() {
			stats.Correct++
		}
		stats.Total++
	}

	sortedElapsedTimes := mathutil.CopySortDurations(elapsedTimes)
	stats.ElapsedAverage = mathutil.MeanDurations(sortedElapsedTimes)
	stats.ElapsedP90 = mathutil.PercentileSortedDurations(sortedElapsedTimes, 90.0)
	stats.ElapsedP95 = mathutil.PercentileSortedDurations(sortedElapsedTimes, 95.0)
	stats.ElapsedMin, stats.ElapsedMax = mathutil.MinMaxDurations(sortedElapsedTimes)
	return
}

// PromptStats returns stats for each prompt.
func (q Quiz) PromptStats() (output []*PromptStats) {
	lookup := make(map[string]*PromptStats)

	for _, res := range q.Results {
		stats, ok := lookup[res.Prompt]
		if ok {
			stats.Total++
			if res.Correct() {
				stats.Correct++
			}
			stats.ElapsedTimes = append(stats.ElapsedTimes, res.Elapsed())
		} else {
			var newStats PromptStats
			newStats.Prompt = res.Prompt
			newStats.Weight = q.PromptWeights[res.Prompt]
			newStats.Total = 1
			if res.Correct() {
				newStats.Correct = 1
			}
			newStats.ElapsedTimes = append(newStats.ElapsedTimes, res.Elapsed())
			output = append(output, &newStats)
			lookup[res.Prompt] = &newStats
		}
	}

	for _, stats := range output {
		stats.ElapsedAverage = mathutil.MeanDurations(stats.ElapsedTimes)
		stats.ElapsedP90 = mathutil.PercentileOfDuration(stats.ElapsedTimes, 90.0)
		stats.ElapsedP95 = mathutil.PercentileOfDuration(stats.ElapsedTimes, 95.0)
		stats.ElapsedMin, stats.ElapsedMax = mathutil.MinMaxDurations(stats.ElapsedTimes)
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].Prompt < output[j].Prompt
	})
	return
}

// NewTestQuiz returns a new test quiz.
func NewTestQuiz(userID uuid.UUID) *Quiz {
	prompts := kana.SelectCount(kana.Merge(kana.Hiragana, kana.Katakana), 10)
	return &Quiz{
		ID:               uuid.V4(),
		UserID:           userID,
		CreatedUTC:       time.Now().UTC(),
		Hiragana:         true,
		Katakana:         true,
		KatakanaWords:    false,
		MaxPrompts:       10,
		MaxQuestions:     0,
		MaxRepeatHistory: 5,
		Prompts:          prompts,
		PromptWeights:    kana.CreateWeights(prompts),
	}
}
