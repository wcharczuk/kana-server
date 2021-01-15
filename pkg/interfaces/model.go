package interfaces

import (
	"context"

	"github.com/blend/go-sdk/uuid"

	"github.com/wcharczuk/kana-server/pkg/types"
)

type Model interface {
	All(ctx context.Context) ([]types.Quiz, error)
	CreateQuiz(context.Context, types.Quiz) error
	GetQuiz(context.Context, uuid.UUID) (types.Quiz, error)
	UpdateQuiz(context.Context, types.Quiz) error
	AddQuizResult(context.Context, types.QuizResult) error
}
