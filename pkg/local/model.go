package local

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/blend/go-sdk/uuid"

	"github.com/wcharczuk/kana-server/pkg/interfaces"
	"github.com/wcharczuk/kana-server/pkg/types"
)

// ErrNotFound is returned if a given quiz is not found.
var ErrNotFound = errors.New("local model; not found")

// Assert the local model manager implements the model interface.
var _ interfaces.Model = (*Model)(nil)

// Model is a local memory implementation of the model manager.
type Model struct {
	mu   sync.RWMutex
	data map[string]*types.Quiz
}

// All retruns all quizzes in the datastore.
func (m *Model) All(ctx context.Context) (output []*types.Quiz, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, q := range m.data {
		output = append(output, q)
	}
	sort.Slice(output, func(i, j int) bool {
		return output[i].CreatedUTC.Before(output[j].CreatedUTC)
	})
	return
}

// CreateQuiz creates a quiz.
func (m *Model) CreateQuiz(_ context.Context, q types.Quiz) error {
	m.mu.Lock()
	if m.data == nil {
		m.data = make(map[string]*types.Quiz)
	}
	m.data[q.ID.String()] = &q
	m.mu.Unlock()
	return nil
}

// GetQuiz returns a quiz.
func (m *Model) GetQuiz(_ context.Context, id uuid.UUID) (types.Quiz, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.data == nil {
		return types.Quiz{}, ErrNotFound
	}
	if value, found := m.data[id.String()]; found {
		return *value, nil
	}
	return types.Quiz{}, ErrNotFound
}

// UpdateQuiz updates a quiz.
func (m *Model) UpdateQuiz(_ context.Context, q types.Quiz) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		return ErrNotFound
	}
	m.data[q.ID.String()] = &q
	return nil
}

func (m *Model) AddQuizResult(_ context.Context, qr types.QuizResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		return ErrNotFound
	}
	quiz, ok := m.data[qr.QuizID.String()]
	if !ok {
		return ErrNotFound
	}
	quiz.Results = append(quiz.Results, qr)
	return nil
}

func (m *Model) Each(_ context.Context, action func(types.Quiz) error) (err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, value := range m.data {
		if err = action(*value); err != nil {
			return
		}
	}
	return
}
