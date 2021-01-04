package types

import (
	"strings"
	"time"

	"github.com/blend/go-sdk/uuid"
)

// Quiz is a quiz result.
type QuizResult struct {
	QuizID      uuid.UUID
	CreatedUTC  time.Time
	AnsweredUTC time.Time
	Prompt      string
	Expected    string
	Actual      string
}

// Elapsed returns the elapsed time as a duration from the answered to the created times.
func (qr QuizResult) Elapsed() time.Duration {
	return qr.AnsweredUTC.Sub(qr.CreatedUTC)
}

// Correct returns if the actual answer matches the expected.
//
// It will trim space, and use a case insensitive equals.
func (qr QuizResult) Correct() bool {
	return strings.EqualFold(
		strings.TrimSpace(qr.Expected),
		strings.TrimSpace(qr.Actual),
	)
}
