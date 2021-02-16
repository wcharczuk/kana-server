package types_test

import (
	"testing"

	"github.com/blend/go-sdk/assert"

	"github.com/wcharczuk/kana-server/pkg/types"
)

func Test_QuizStats_PercentCorrect(t *testing.T) {
	its := assert.New(t)

	its.Zero(types.QuizStats{}.PercentCorrect())
	its.Equal(25.0, types.QuizStats{Total: 4, Correct: 1}.PercentCorrect())
}
