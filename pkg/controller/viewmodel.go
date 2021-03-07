package controller

import (
	"time"

	"github.com/wcharczuk/kana-server/pkg/types"
)

// CreateHomeViewModel creates a home view model.
func CreateHomeViewModel(all []*types.Quiz) (hvm HomeViewModel) {
	hvm.TotalQuizzes = len(all)
	for _, q := range all {
		hvm.TotalQuizResults += len(q.Results)
		for _, qr := range q.Results {
			if qr.Correct() {
				hvm.TotalQuizResultsCorrect++
			}
			hvm.TotalQuizDuration += qr.Elapsed()
		}
	}
	hvm.Quizzes = all
	return
}

// HomeViewModel is the viewmodel for the home page.
type HomeViewModel struct {
	TotalQuizzes            int
	TotalQuizResults        int
	TotalQuizResultsCorrect int
	TotalQuizDuration       time.Duration
	Quizzes                 []*types.Quiz
}

// TotalQuizCorrectPct returns the percentage of results correct.
func (hvm HomeViewModel) TotalQuizCorrectPct() float64 {
	if hvm.TotalQuizResults == 0 {
		return 0
	}
	return (float64(hvm.TotalQuizResultsCorrect) / float64(hvm.TotalQuizResults)) * 100.0
}
