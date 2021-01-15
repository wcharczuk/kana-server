package types

import (
	"strings"
	"time"

	"github.com/blend/go-sdk/uuid"
)

// Quiz is a quiz result.
type QuizResult struct {
	ID          uuid.UUID `db:"id,pk"`
	QuizID      uuid.UUID `db:"quiz_id"`
	CreatedUTC  time.Time `db:"created_utc"`
	AnsweredUTC time.Time `db:"answered_utc"`
	Prompt      string    `db:"prompt"`
	Expected    string    `db:"expected"`
	Actual      string    `db:"actual"`
}

// TableName returns the database tablename for the type.
func (qr QuizResult) TableName() string { return "quiz_result " }

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
