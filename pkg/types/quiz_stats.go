package types

import "time"

type QuizStats struct {
	Total          int
	Correct        int
	ElapsedAverage time.Duration
	ElapsedMin     time.Duration
	ElapsedMax     time.Duration
	ElapsedP75     time.Duration
	ElapsedP90     time.Duration
	ElapsedP95     time.Duration
}

// PercentCorrect returns the percentage correct.
func (qs QuizStats) PercentCorrect() float64 {
	if qs.Total == 0 {
		return 0.0
	}
	return float64(qs.Correct) / float64(qs.Total)
}
