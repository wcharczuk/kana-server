package model_test

import (
	"testing"

	"github.com/wcharczuk/kana-server/pkg/types"
)

func Test_Model_AllQuizzes(t *testing.T) {
	its, mgr, done := NewTest(t)
	defer done()

	u0 := CreateTestUser(its, mgr)
	u1 := CreateTestUser(its, mgr)

	q0 := CreateTestQuiz(its, mgr, u0)
	q1 := CreateTestQuiz(its, mgr, u0)
	q2 := CreateTestQuiz(its, mgr, u0)
	q3 := CreateTestQuiz(its, mgr, u1)

	all, err := mgr.AllQuzzes(its.Background(), u0.ID)
	its.Nil(err)
	its.Len(all, 3)
	its.Any(all, func(v interface{}) bool { return v.(*types.Quiz).ID.Equal(q0.ID) })
	its.Any(all, func(v interface{}) bool { return v.(*types.Quiz).ID.Equal(q1.ID) })
	its.Any(all, func(v interface{}) bool { return v.(*types.Quiz).ID.Equal(q2.ID) })
	its.None(all, func(v interface{}) bool { return v.(*types.Quiz).ID.Equal(q3.ID) })
	its.All(all, func(v interface{}) bool { return len(v.(*types.Quiz).Results) > 0 })
	its.All(all, func(v interface{}) bool {
		typed, _ := v.(*types.Quiz)
		for _, qr := range typed.Results {
			if !qr.QuizID.Equal(typed.ID) {
				return false
			}
		}
		return true
	})
}
