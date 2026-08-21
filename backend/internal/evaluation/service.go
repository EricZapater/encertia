package evaluation

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/encertia/backend/internal/quiz"
)

var (
	ErrUnauthorized       = errors.New("unauthorized access to evaluation")
	ErrInvalidGrade       = errors.New("grade must be between 0.00 and 10.00")
	ErrEvaluationNotFound = errors.New("evaluation not found")
)

type Service interface {
	ListEvaluations(userID, role string) ([]EvaluationQuizSummary, error)
	GetQuizEvaluation(quizID, userID, role string) (*QuizEvaluationResponse, error)
	GetStudentEvaluation(quizID, studentID, userID, role string) (*StudentEvaluationDetail, error)
	GradeStudent(quizID, studentID, teacherID, role string, finalGrade float64) (*GradeResponse, error)
	OnMatchFinished(matchID string) error
}

type service struct {
	repo        Repository
	quizService quiz.Service
}

func NewService(repo Repository, quizService quiz.Service) Service {
	return &service{
		repo:        repo,
		quizService: quizService,
	}
}

func (s *service) ListEvaluations(userID, role string) ([]EvaluationQuizSummary, error) {
	if role == "student" {
		return nil, ErrUnauthorized
	}
	isAdmin := role == "admin"
	return s.repo.ListEvaluations(userID, isAdmin)
}

func (s *service) GetQuizEvaluation(quizID, userID, role string) (*QuizEvaluationResponse, error) {
	if role == "student" {
		return nil, ErrUnauthorized
	}

	if role != "admin" {
		uID, err := uuid.Parse(userID)
		if err != nil {
			return nil, ErrUnauthorized
		}
		qID, err := uuid.Parse(quizID)
		if err != nil {
			return nil, errors.New("quiz not found")
		}
		q, appErr := s.quizService.GetQuizByID(context.Background(), uID, role, qID)
		if appErr != nil {
			return nil, errors.New("quiz not found")
		}
		if q.CreatorID != uID {
			return nil, ErrUnauthorized
		}
	}

	return s.repo.GetQuizEvaluation(quizID)
}

func (s *service) GetStudentEvaluation(quizID, studentID, userID, role string) (*StudentEvaluationDetail, error) {
	if role == "student" {
		return nil, ErrUnauthorized
	}

	if role != "admin" {
		uID, err := uuid.Parse(userID)
		if err != nil {
			return nil, ErrUnauthorized
		}
		qID, err := uuid.Parse(quizID)
		if err != nil {
			return nil, errors.New("quiz not found")
		}
		q, appErr := s.quizService.GetQuizByID(context.Background(), uID, role, qID)
		if appErr != nil {
			return nil, errors.New("quiz not found")
		}
		if q.CreatorID != uID {
			return nil, ErrUnauthorized
		}
	}

	return s.repo.GetStudentEvaluation(quizID, studentID)
}

func (s *service) GradeStudent(quizID, studentID, teacherID, role string, finalGrade float64) (*GradeResponse, error) {
	if role == "student" {
		return nil, ErrUnauthorized
	}

	if role != "admin" {
		uID, err := uuid.Parse(teacherID)
		if err != nil {
			return nil, ErrUnauthorized
		}
		qID, err := uuid.Parse(quizID)
		if err != nil {
			return nil, errors.New("quiz not found")
		}
		q, appErr := s.quizService.GetQuizByID(context.Background(), uID, role, qID)
		if appErr != nil {
			return nil, errors.New("quiz not found")
		}
		if q.CreatorID != uID {
			return nil, ErrUnauthorized
		}
	}

	if _, err := uuid.Parse(studentID); err != nil {
		return nil, errors.New("invalid student ID")
	}

	if finalGrade < 0.0 || finalGrade > 10.0 {
		return nil, ErrInvalidGrade
	}

	// Truncate/round to 2 decimals
	finalGrade = math.Round(finalGrade*100) / 100

	return s.repo.GradeStudent(quizID, studentID, teacherID, finalGrade)
}

func (s *service) OnMatchFinished(matchID string) error {
	if err := s.repo.UpsertCalculatedGradeForMatch(matchID); err != nil {
		return fmt.Errorf("error processing match finished event for evaluations: %w", err)
	}
	return nil
}
