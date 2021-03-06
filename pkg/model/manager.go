package model

import (
	"context"
	"fmt"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/uuid"

	"github.com/wcharczuk/kana-server/pkg/types"
)

var (
	userCols            = db.Columns(types.User{})
	userTableName       = types.User{}.TableName()
	quizCols            = db.Columns(types.Quiz{})
	quizTableName       = types.Quiz{}.TableName()
	quizResultCols      = db.Columns(types.QuizResult{})
	quizResultTableName = types.QuizResult{}.TableName()

	getQuizzesForUserQuery            = fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1 ORDER BY created_utc desc", quizCols.ColumnNamesCSV(), quizTableName)
	getQuizzesForUserQuizResultsQuery = fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", quizResultCols.ColumnNamesCSV(), quizResultTableName)
	getQuizResultsQuery               = fmt.Sprintf("SELECT %s FROM %s WHERE quiz_id = $1", quizResultCols.ColumnNamesCSV(), quizResultTableName)
	getUserByEmailQuery               = fmt.Sprintf("SELECT %s FROM %s WHERE email = $1", userCols.ColumnNamesCSV(), userTableName)
)

// New returns a new model manager.
func New(conn *db.Connection, opts ...db.InvocationOption) *Manager {
	return &Manager{
		BaseManager: dbutil.NewBaseManager(conn, opts...),
	}
}

// Manager implements database functions.
type Manager struct {
	dbutil.BaseManager
}

// AllQuzzes returns all the quizzes.
func (m Manager) AllQuzzes(ctx context.Context, userID uuid.UUID) (output []*types.Quiz, err error) {
	lookup := map[string]*types.Quiz{}
	err = m.Invoke(ctx).Query(getQuizzesForUserQuery, userID).Each(func(r db.Rows) error {
		var q types.Quiz
		if populateErr := db.PopulateInOrder(&q, r, quizCols); populateErr != nil {
			return populateErr
		}
		output = append(output, &q)
		lookup[q.ID.String()] = &q
		return nil
	})
	if err != nil {
		return
	}
	err = m.Invoke(ctx).Query(getQuizzesForUserQuizResultsQuery, userID).Each(func(r db.Rows) error {
		var qr types.QuizResult
		if populateErr := db.PopulateInOrder(&qr, r, quizResultCols); populateErr != nil {
			return populateErr
		}
		if quiz, ok := lookup[qr.QuizID.String()]; ok {
			quiz.Results = append(quiz.Results, qr)
		} else {
			return fmt.Errorf("quiz for quiz result not found: %s", qr.QuizID.String())
		}
		return nil
	})
	return
}

// CreateQuiz creates a quiz.
func (m Manager) CreateQuiz(ctx context.Context, q types.Quiz) error {
	return m.Invoke(ctx).Create(&q)
}

// GetQuiz gets a quiz and associated results.
func (m Manager) GetQuiz(ctx context.Context, id uuid.UUID) (output types.Quiz, found bool, err error) {
	found, err = m.Invoke(ctx).Get(&output, id)
	if err != nil || !found {
		return
	}
	err = m.Invoke(ctx).Query(getQuizResultsQuery, id).Each(func(r db.Rows) error {
		var qr types.QuizResult
		if populateErr := db.PopulateInOrder(&qr, r, quizResultCols); populateErr != nil {
			return populateErr
		}
		output.Results = append(output.Results, qr)
		return nil
	})
	return
}

// UpdateQuiz updates a quiz.
func (m Manager) UpdateQuiz(ctx context.Context, q types.Quiz) error {
	_, err := m.Invoke(ctx).Update(&q)
	return err
}

// AddQuizResult creates a new quiz result.
func (m Manager) AddQuizResult(ctx context.Context, qr types.QuizResult) error {
	return m.Invoke(ctx).Create(&qr)
}

// GetUserByEmail gets a user by email.
func (m Manager) GetUserByEmail(ctx context.Context, email string) (output types.User, found bool, err error) {
	found, err = m.Invoke(ctx).Query(getUserByEmailQuery, email).Out(&output)
	return
}
