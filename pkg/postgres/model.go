package postgres

import (
	"context"
	"fmt"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/dbutil"
	"github.com/blend/go-sdk/uuid"

	"github.com/wcharczuk/kana-server/pkg/interfaces"
	"github.com/wcharczuk/kana-server/pkg/types"
)

var (
	_ interfaces.Model = (*Model)(nil)
)

var (
	quizCols                   = db.Columns(types.Quiz{})
	getQuizzesQuery            = fmt.Sprintf("SELECT %s FROM %s ORDER BY created_utc desc", quizCols.ColumnNamesCSV(), types.Quiz{}.TableName())
	getQuizzesQuizResultsQuery = fmt.Sprintf("SELECT %s FROM %s", quizResultCols.ColumnNamesCSV(), types.QuizResult{}.TableName())
	quizResultCols             = db.Columns(types.QuizResult{})
	getQuizResultsQuery        = fmt.Sprintf("SELECT %s FROM %s WHERE quiz_id = $1", quizResultCols.ColumnNamesCSV(), types.QuizResult{}.TableName())
)

// Model
type Model struct {
	dbutil.BaseManager
}

// All returns all the quizzes.
func (m Model) All(ctx context.Context) (output []*types.Quiz, err error) {
	lookup := map[string]*types.Quiz{}
	err = m.Invoke(ctx).Query(getQuizzesQuery).Each(func(r db.Rows) error {
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
	err = m.Invoke(ctx).Query(getQuizzesQuizResultsQuery).Each(func(r db.Rows) error {
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
func (m Model) CreateQuiz(ctx context.Context, q types.Quiz) error {
	return m.Invoke(ctx).Create(&q)
}

// GetQuiz gets a quiz and associated results.
func (m Model) GetQuiz(ctx context.Context, id uuid.UUID) (output types.Quiz, err error) {
	var found bool
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
func (m Model) UpdateQuiz(ctx context.Context, q types.Quiz) error {
	_, err := m.Invoke(ctx).Update(&q)
	return err
}

// AddQuizResult creates a new quiz result.
func (m Model) AddQuizResult(ctx context.Context, qr types.QuizResult) error {
	return m.Invoke(ctx).Create(&qr)
}
