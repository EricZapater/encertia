package quiz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

const (
	ErrCodeQuizNotFound = "QUIZ_NOT_FOUND"
)

// Service defines the business logic contract for the Quiz domain.
type Service interface {
	ListQuizzes(ctx context.Context, actorID uuid.UUID, actorRole string, filter QuizListFilters) (*QuizListResponse, *shared.AppError)
	GetQuizByID(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID) (*QuizDetail, *shared.AppError)
	CreateQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, input CreateQuizInput) (*QuizDetail, *shared.AppError)
	UpdateQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID, input UpdateQuizInput) (*QuizDetail, *shared.AppError)
	DeleteQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID) (*shared.MessageResponse, *shared.AppError)
	DuplicateQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID, input DuplicateQuizInput) (*QuizDetail, *shared.AppError)
}

type service struct {
	repo Repository
}

// NewService creates a new quiz Service instance.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ListQuizzes returns quizzes filtered by permissions and criteria.
func (s *service) ListQuizzes(ctx context.Context, actorID uuid.UUID, actorRole string, filter QuizListFilters) (*QuizListResponse, *shared.AppError) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 12
	}
	if filter.PageSize > 50 {
		filter.PageSize = 50
	}

	// Non-admin actors can only view their own quizzes
	if actorRole != "admin" {
		filter.CreatorID = &actorID
	}

	items, totalCount, err := s.repo.ListQuizzes(ctx, filter)
	if err != nil {
		return nil, shared.ErrInternal(fmt.Errorf("error recuperant llistat de qüestionaris: %w", err))
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(filter.PageSize)))
	}

	return &QuizListResponse{
		Items: items,
		Pagination: QuizPagination{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// GetQuizByID returns a single quiz by ID with permission checks.
func (s *service) GetQuizByID(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID) (*QuizDetail, *shared.AppError) {
	if quizID == uuid.Nil {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Identificador de qüestionari invàlid.", map[string]interface{}{"field": "id"})
	}

	q, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, shared.ErrInternal(fmt.Errorf("error obtenint qüestionari: %w", err))
	}
	if q == nil {
		return nil, shared.NewAppError(404, ErrCodeQuizNotFound, "No s'ha trobat el qüestionari sol·licitat.", nil, nil)
	}

	// RBAC: Only owner or admin can view
	if actorRole != "admin" && q.CreatorID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per accedir a aquest qüestionari.")
	}

	return q, nil
}

// CreateQuiz handles validation and creation of a new quiz.
func (s *service) CreateQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, input CreateQuizInput) (*QuizDetail, *shared.AppError) {
	input.Title = strings.TrimSpace(input.Title)
	if len(input.Title) < 3 || len(input.Title) > 150 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol del qüestionari ha de tenir entre 3 i 150 caràcters.", map[string]interface{}{"field": "title"})
	}

	if input.Description != nil {
		desc := strings.TrimSpace(*input.Description)
		if len(desc) > 1000 {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "La descripció no pot superar els 1000 caràcters.", map[string]interface{}{"field": "description"})
		}
		input.Description = &desc
	}

	if input.Status == "" {
		input.Status = StatusDraft
	}
	if input.Status != StatusDraft && input.Status != StatusPublished && input.Status != StatusArchived {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "L'estat del qüestionari ha de ser 'draft', 'published' o 'archived'.", map[string]interface{}{"field": "status"})
	}

	// Clean tags
	cleanedTags := make([]string, 0)
	for _, tag := range input.Tags {
		t := strings.TrimSpace(tag)
		if t != "" {
			cleanedTags = append(cleanedTags, t)
		}
	}
	input.Tags = cleanedTags

	// Validate questions
	if appErr := validateQuestions(input.Status, input.Questions); appErr != nil {
		return nil, appErr
	}

	created, err := s.repo.CreateQuiz(ctx, actorID, input)
	if err != nil {
		return nil, shared.ErrInternal(fmt.Errorf("error creant qüestionari: %w", err))
	}

	return created, nil
}

// UpdateQuiz validates and updates an existing quiz.
func (s *service) UpdateQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID, input UpdateQuizInput) (*QuizDetail, *shared.AppError) {
	if quizID == uuid.Nil {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Identificador de qüestionari invàlid.", map[string]interface{}{"field": "id"})
	}

	existing, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, shared.ErrInternal(fmt.Errorf("error consultant qüestionari per actualitzar: %w", err))
	}
	if existing == nil {
		return nil, shared.NewAppError(404, ErrCodeQuizNotFound, "No s'ha trobat el qüestionari sol·licitat.", nil, nil)
	}

	// RBAC: Only owner or admin can update
	if actorRole != "admin" && existing.CreatorID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per modificar aquest qüestionari.")
	}

	input.Title = strings.TrimSpace(input.Title)
	if len(input.Title) < 3 || len(input.Title) > 150 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol del qüestionari ha de tenir entre 3 i 150 caràcters.", map[string]interface{}{"field": "title"})
	}

	if input.Description != nil {
		desc := strings.TrimSpace(*input.Description)
		if len(desc) > 1000 {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "La descripció no pot superar els 1000 caràcters.", map[string]interface{}{"field": "description"})
		}
		input.Description = &desc
	}

	targetStatus := existing.Status
	if input.Status != nil {
		if *input.Status != StatusDraft && *input.Status != StatusPublished && *input.Status != StatusArchived {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "L'estat del qüestionari ha de ser 'draft', 'published' o 'archived'.", map[string]interface{}{"field": "status"})
		}
		targetStatus = *input.Status
	}

	// Clean tags
	if input.Tags != nil {
		cleanedTags := make([]string, 0)
		for _, tag := range input.Tags {
			t := strings.TrimSpace(tag)
			if t != "" {
				cleanedTags = append(cleanedTags, t)
			}
		}
		input.Tags = cleanedTags
	}

	// Validate questions
	if input.Questions != nil {
		if appErr := validateQuestions(targetStatus, input.Questions); appErr != nil {
			return nil, appErr
		}
	} else if targetStatus == StatusPublished {
		if len(existing.Questions) == 0 {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Un qüestionari publicat ha de tenir com a mínim 1 pregunta.", map[string]interface{}{"field": "questions"})
		}
	}

	updated, err := s.repo.UpdateQuiz(ctx, quizID, input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.NewAppError(404, ErrCodeQuizNotFound, "No s'ha trobat el qüestionari sol·licitat.", nil, nil)
		}
		return nil, shared.ErrInternal(fmt.Errorf("error actualitzant qüestionari: %w", err))
	}

	return updated, nil
}

// DeleteQuiz soft-deletes a quiz.
func (s *service) DeleteQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID) (*shared.MessageResponse, *shared.AppError) {
	if quizID == uuid.Nil {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Identificador de qüestionari invàlid.", map[string]interface{}{"field": "id"})
	}

	existing, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, shared.ErrInternal(fmt.Errorf("error consultant qüestionari per eliminar: %w", err))
	}
	if existing == nil {
		return nil, shared.NewAppError(404, ErrCodeQuizNotFound, "No s'ha trobat el qüestionari sol·licitat.", nil, nil)
	}

	// RBAC: Only owner or admin can delete
	if actorRole != "admin" && existing.CreatorID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per eliminar aquest qüestionari.")
	}

	if err := s.repo.DeleteQuiz(ctx, quizID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.NewAppError(404, ErrCodeQuizNotFound, "No s'ha trobat el qüestionari sol·licitat.", nil, nil)
		}
		return nil, shared.ErrInternal(fmt.Errorf("error eliminant qüestionari: %w", err))
	}

	return &shared.MessageResponse{Message: "Qüestionari eliminat correctament"}, nil
}

// DuplicateQuiz duplicates a quiz for the requesting actor.
func (s *service) DuplicateQuiz(ctx context.Context, actorID uuid.UUID, actorRole string, quizID uuid.UUID, input DuplicateQuizInput) (*QuizDetail, *shared.AppError) {
	if quizID == uuid.Nil {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Identificador de qüestionari invàlid.", map[string]interface{}{"field": "id"})
	}

	existing, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, shared.ErrInternal(fmt.Errorf("error consultant qüestionari a duplicar: %w", err))
	}
	if existing == nil {
		return nil, shared.NewAppError(404, ErrCodeQuizNotFound, "No s'ha trobat el qüestionari sol·licitat.", nil, nil)
	}

	// RBAC: Only owner or admin can duplicate
	if actorRole != "admin" && existing.CreatorID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per duplicar aquest qüestionari.")
	}

	var newTitle string
	if input.Title != nil && strings.TrimSpace(*input.Title) != "" {
		newTitle = strings.TrimSpace(*input.Title)
		if len(newTitle) < 3 || len(newTitle) > 150 {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol personalitzat de la còpia ha de tenir entre 3 i 150 caràcters.", map[string]interface{}{"field": "title"})
		}
	} else {
		newTitle = "[Còpia] " + existing.Title
		if len(newTitle) > 150 {
			newTitle = newTitle[:150]
		}
	}

	duplicated, err := s.repo.DuplicateQuiz(ctx, quizID, actorID, newTitle, input.IncludeAnswers)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, shared.NewAppError(404, ErrCodeQuizNotFound, "No s'ha trobat el qüestionari sol·licitat.", nil, nil)
		}
		return nil, shared.ErrInternal(fmt.Errorf("error duplicant qüestionari: %w", err))
	}

	return duplicated, nil
}

// validateQuestions ensures question count, time limits, and answer rules match the quiz state.
func validateQuestions(status QuizStatus, questions []SaveQuestionInput) *shared.AppError {
	if status == StatusPublished {
		if len(questions) == 0 {
			return shared.ErrBadRequest(shared.ErrCodeValidation, "Un qüestionari publicat ha de tenir com a mínim 1 pregunta.", map[string]interface{}{"field": "questions"})
		}
	}

	for i, q := range questions {
		qText := strings.TrimSpace(q.Text)
		if qText == "" || len(qText) > 500 {
			return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("La pregunta #%d ha de tenir un text d'entre 1 i 500 caràcters.", i+1), map[string]interface{}{"questionIndex": i, "field": "text"})
		}

		if q.QuestionType != "" && q.QuestionType != QuestionTypeSingle && q.QuestionType != QuestionTypeMultiple {
			return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("El tipus de la pregunta #%d ha de ser 'single_choice' o 'multiple_choice'.", i+1), map[string]interface{}{"questionIndex": i, "field": "questionType"})
		}

		if q.TimeLimitSeconds != 0 && !ValidTimeLimits[q.TimeLimitSeconds] {
			return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("El temps límit de la pregunta #%d ha de ser un de: 5, 10, 20, 30, 60, 90, 120 segons.", i+1), map[string]interface{}{"questionIndex": i, "field": "timeLimitSeconds"})
		}

		if status == StatusPublished {
			if len(q.Answers) < 2 || len(q.Answers) > 6 {
				return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("La pregunta #%d ha de tenir entre 2 i 6 opcions de resposta.", i+1), map[string]interface{}{"questionIndex": i, "field": "answers"})
			}

			correctCount := 0
			for j, a := range q.Answers {
				aText := strings.TrimSpace(a.Text)
				if aText == "" || len(aText) > 300 {
					return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("La resposta #%d de la pregunta #%d ha de tenir un text d'entre 1 i 300 caràcters.", j+1, i+1), map[string]interface{}{"questionIndex": i, "answerIndex": j, "field": "text"})
				}
				if a.OrderIndex < 0 || a.OrderIndex > 5 {
					return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("L'índex d'ordre de la resposta #%d ha d'estar entre 0 i 5.", j+1), map[string]interface{}{"questionIndex": i, "answerIndex": j, "field": "orderIndex"})
				}
				if a.IsCorrect {
					correctCount++
				}
			}

			qType := q.QuestionType
			if qType == "" {
				qType = QuestionTypeSingle
			}

			if qType == QuestionTypeSingle && correctCount != 1 {
				return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("La pregunta #%d és de resposta única i ha de tenir exactament 1 resposta correcta seleccionada (trobades: %d).", i+1, correctCount), map[string]interface{}{"questionIndex": i, "correctCount": correctCount})
			}
			if qType == QuestionTypeMultiple && correctCount < 1 {
				return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("La pregunta #%d és de resposta múltiple i ha de tenir com a mínim 1 resposta correcta seleccionada.", i+1), map[string]interface{}{"questionIndex": i})
			}
		} else {
			// For draft / archived, validate limits if answers are supplied
			if len(q.Answers) > 6 {
				return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("La pregunta #%d no pot tenir més de 6 opcions de resposta.", i+1), map[string]interface{}{"questionIndex": i, "field": "answers"})
			}
			for j, a := range q.Answers {
				aText := strings.TrimSpace(a.Text)
				if aText == "" || len(aText) > 300 {
					return shared.ErrBadRequest(shared.ErrCodeValidation, fmt.Sprintf("La resposta #%d de la pregunta #%d ha de tenir un text d'entre 1 i 300 caràcters.", j+1, i+1), map[string]interface{}{"questionIndex": i, "answerIndex": j, "field": "text"})
				}
			}
		}
	}

	return nil
}
