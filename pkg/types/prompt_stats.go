package types

import "time"

type PromptStats struct {
	Prompt         string
	Weight         float64
	Total          int
	Correct        int
	ElapsedAverage time.Duration
	ElapsedMin     time.Duration
	ElapsedMax     time.Duration
	ElapsedP75     time.Duration
	ElapsedP90     time.Duration
	ElapsedP95     time.Duration
	ElapsedTimes   []time.Duration
}

// PercentCorrect returns the percentage correct.
func (ps PromptStats) PercentCorrect() float64 {
	if ps.Total == 0 {
		return 0.0
	}
	return float64(ps.Correct) / float64(ps.Total)
}
