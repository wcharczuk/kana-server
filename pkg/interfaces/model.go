package interfaces

import (
	"context"

	"github.com/blend/go-sdk/uuid"

	"github.com/wcharczuk/kana-server/pkg/types"
)

type Model interface {
	CreateQuiz(context.Context, types.Quiz) error
	GetQuiz(context.Context, uuid.UUID) (types.Quiz, error)
	AddQuizResult(context.Context, types.QuizResult) error
	Each(context.Context, func(types.Quiz) error) error
}
