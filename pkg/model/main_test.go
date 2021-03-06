package model_test

import (
	"context"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"
	"github.com/blend/go-sdk/testutil"

	"github.com/wcharczuk/kana-server/pkg/model"
	"github.com/wcharczuk/kana-server/pkg/types"
)

func TestMain(m *testing.M) {
	log := logger.All()
	testutil.New(m,
		testutil.OptLog(log),
		testutil.OptWithDefaultDB(),
		testutil.OptBefore(
			func(ctx context.Context) error {
				return model.Migrations().Apply(ctx, testutil.DefaultDB())
			},
		),
		testutil.OptAfter(
			func(ctx context.Context) error {
				log.Close()
				return nil
			},
		),
	).Run()
}

// NewTest returns a new test context.
func NewTest(t *testing.T) (*assert.Assertions, *model.Manager, func()) {
	its := assert.New(t)
	tx, err := testutil.DefaultDB().Begin()
	its.Nil(err)
	return its, model.New(testutil.DefaultDB(), db.OptTx(tx)), func() {
		_ = tx.Rollback()
	}
}

// CreateTestUser creates a test user.
func CreateTestUser(its *assert.Assertions, mgr *model.Manager) *types.User {
	u0 := types.NewTestUser()
	its.Nil(mgr.Invoke(its.Background()).Create(&u0))
	return &u0
}

// CreateTestQuiz creates a test quiz.
func CreateTestQuiz(its *assert.Assertions, mgr *model.Manager, user *types.User) *types.Quiz {
	q0 := types.NewTestQuiz(user.ID)
	its.Nil(mgr.Invoke(its.Background()).Create(q0))

	q0r0 := types.NewTestQuizResultCorrect(q0)
	its.Nil(mgr.Invoke(its.Background()).Create(q0r0))
	q0r1 := types.NewTestQuizResultCorrect(q0)
	its.Nil(mgr.Invoke(its.Background()).Create(q0r1))
	q0r2 := types.NewTestQuizResultCorrect(q0)
	its.Nil(mgr.Invoke(its.Background()).Create(q0r2))
	q0r3 := types.NewTestQuizResultIncorrect(q0)
	its.Nil(mgr.Invoke(its.Background()).Create(q0r3))
	q0r4 := types.NewTestQuizResultCorrect(q0)
	its.Nil(mgr.Invoke(its.Background()).Create(q0r4))
	q0r5 := types.NewTestQuizResultCorrect(q0)
	its.Nil(mgr.Invoke(its.Background()).Create(q0r5))
	q0r6 := types.NewTestQuizResultIncorrect(q0)
	its.Nil(mgr.Invoke(its.Background()).Create(q0r6))
	return q0
}
