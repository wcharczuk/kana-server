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
