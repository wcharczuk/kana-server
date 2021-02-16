package types_test

import (
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"

	"github.com/wcharczuk/kana-server/pkg/types"
)

func Test_Quiz_PromptStats(t *testing.T) {
	its := assert.New(t)

	q := types.Quiz{
		Results: []types.QuizResult{
			{ //
				Prompt:      "p0",
				Expected:    "pa0",
				Actual:      "pa0",
				CreatedUTC:  time.Now().UTC(),
				AnsweredUTC: time.Now().UTC().Add(time.Second),
			},
			{ //
				Prompt:      "p1",
				Expected:    "pa1",
				Actual:      "pa1",
				CreatedUTC:  time.Now().UTC(),
				AnsweredUTC: time.Now().UTC().Add(2 * time.Second),
			},
			{ //
				Prompt:      "p2",
				Expected:    "pa2",
				Actual:      "pa2",
				CreatedUTC:  time.Now().UTC(),
				AnsweredUTC: time.Now().UTC().Add(3 * time.Second),
			},
			{ //
				Prompt:      "p0",
				Expected:    "pa0",
				Actual:      "npa0",
				CreatedUTC:  time.Now().UTC(),
				AnsweredUTC: time.Now().UTC().Add(2 * time.Second),
			},
			{ //
				Prompt:      "p1",
				Expected:    "pa1",
				Actual:      "pa1",
				CreatedUTC:  time.Now().UTC(),
				AnsweredUTC: time.Now().UTC().Add(3 * time.Second),
			},
			{ //
				Prompt:      "p2",
				Expected:    "pa2",
				Actual:      "pa2",
				CreatedUTC:  time.Now().UTC(),
				AnsweredUTC: time.Now().UTC().Add(time.Second),
			},
			{ //
				Prompt:      "p1",
				Expected:    "pa1",
				Actual:      "pa1",
				CreatedUTC:  time.Now().UTC(),
				AnsweredUTC: time.Now().UTC().Add(time.Second),
			},
		},
	}

	ps := q.PromptStats()
	its.Len(ps, 3)
}
