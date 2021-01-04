package types

import (
	"time"
)

// QuizPrompt is a specific quiz prompt.
type QuizPrompt struct {
	Quiz       Quiz
	CreatedUTC time.Time
	Prompt     string
	Expected   string
}
