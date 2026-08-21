package quiz_test

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/encertia/backend/internal/quiz"
	"github.com/google/uuid"
)

type mockRepository struct {
	mu        sync.RWMutex
	quizzes   map[uuid.UUID]*quiz.QuizDetail
	deleted   map[uuid.UUID]bool
	userNames map[uuid.UUID]string
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		quizzes:   make(map[uuid.UUID]*quiz.QuizDetail),
		deleted:   make(map[uuid.UUID]bool),
		userNames: make(map[uuid.UUID]string),
	}
}

func (m *mockRepository) ListQuizzes(ctx context.Context, filter quiz.QuizListFilters) ([]quiz.Quiz, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []quiz.Quiz
	for id, qd := range m.quizzes {
		if m.deleted[id] {
			continue
		}

		if filter.CreatorID != nil && *filter.CreatorID != uuid.Nil && qd.CreatorID != *filter.CreatorID {
			continue
		}

		if filter.Status != "" && string(qd.Status) != filter.Status {
			continue
		}

		if filter.Tag != "" {
			hasTag := false
			for _, t := range qd.Tags {
				if t == filter.Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		if filter.Search != "" {
			s := strings.ToLower(filter.Search)
			titleMatch := strings.Contains(strings.ToLower(qd.Title), s)
			descMatch := qd.Description != nil && strings.Contains(strings.ToLower(*qd.Description), s)
			if !titleMatch && !descMatch {
				continue
			}
		}

		matched = append(matched, qd.Quiz)
	}

	totalCount := len(matched)
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 12
	}

	start := (page - 1) * pageSize
	if start >= totalCount {
		return []quiz.Quiz{}, totalCount, nil
	}
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return matched[start:end], totalCount, nil
}

func (m *mockRepository) GetQuizByID(ctx context.Context, id uuid.UUID) (*quiz.QuizDetail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.deleted[id] {
		return nil, nil
	}

	qd, exists := m.quizzes[id]
	if !exists {
		return nil, nil
	}

	// Return a clone
	clone := *qd
	clone.Questions = make([]quiz.QuizQuestion, len(qd.Questions))
	for i, q := range qd.Questions {
		qClone := q
		qClone.Answers = make([]quiz.QuizAnswer, len(q.Answers))
		copy(qClone.Answers, q.Answers)
		clone.Questions[i] = qClone
	}

	return &clone, nil
}

func (m *mockRepository) CreateQuiz(ctx context.Context, creatorID uuid.UUID, input quiz.CreateQuizInput) (*quiz.QuizDetail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	quizID := uuid.New()
	status := input.Status
	if status == "" {
		status = quiz.StatusDraft
	}
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}

	qd := &quiz.QuizDetail{
		Quiz: quiz.Quiz{
			ID:            quizID,
			CreatorID:     creatorID,
			CreatorName:   m.userNames[creatorID],
			Title:         input.Title,
			Description:   input.Description,
			CoverImageURL: input.CoverImageURL,
			Status:        status,
			Tags:          tags,
			QuestionCount: len(input.Questions),
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		Questions: make([]quiz.QuizQuestion, 0, len(input.Questions)),
	}

	for qIdx, qInput := range input.Questions {
		qID := quiz.ParseOrGenerateUUID(qInput.ID)
		timeLimit := qInput.TimeLimitSeconds
		if timeLimit == 0 {
			timeLimit = 20
		}
		qType := qInput.QuestionType
		if qType == "" {
			qType = quiz.QuestionTypeSingle
		}
		orderIdx := qInput.OrderIndex
		if orderIdx == 0 && qIdx > 0 {
			orderIdx = qIdx
		}

		qq := quiz.QuizQuestion{
			ID:               qID,
			QuizID:           quizID,
			Text:             qInput.Text,
			ImageURL:         qInput.ImageURL,
			QuestionType:     qType,
			TimeLimitSeconds: timeLimit,
			OrderIndex:       orderIdx,
			Answers:          make([]quiz.QuizAnswer, 0, len(qInput.Answers)),
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}

		for aIdx, aInput := range qInput.Answers {
			ansOrderIdx := aInput.OrderIndex
			if ansOrderIdx == 0 && aIdx > 0 {
				ansOrderIdx = aIdx
			}
			ans := quiz.QuizAnswer{
				ID:         quiz.ParseOrGenerateUUID(aInput.ID),
				QuestionID: qID,
				Text:       aInput.Text,
				IsCorrect:  aInput.IsCorrect,
				OrderIndex: ansOrderIdx,
				CreatedAt:  time.Now().UTC(),
			}
			qq.Answers = append(qq.Answers, ans)
		}

		qd.Questions = append(qd.Questions, qq)
	}

	m.quizzes[quizID] = qd
	return qd, nil
}

func (m *mockRepository) UpdateQuiz(ctx context.Context, id uuid.UUID, input quiz.UpdateQuizInput) (*quiz.QuizDetail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleted[id] {
		return nil, sql.ErrNoRows
	}

	qd, exists := m.quizzes[id]
	if !exists {
		return nil, sql.ErrNoRows
	}

	qd.Title = input.Title
	qd.Description = input.Description
	qd.CoverImageURL = input.CoverImageURL
	if input.Status != nil {
		qd.Status = *input.Status
	}
	if input.Tags != nil {
		qd.Tags = input.Tags
	}
	qd.UpdatedAt = time.Now().UTC()

	if input.Questions != nil {
		qd.Questions = make([]quiz.QuizQuestion, 0, len(input.Questions))
		for qIdx, qInput := range input.Questions {
			qID := quiz.ParseOrGenerateUUID(qInput.ID)
			timeLimit := qInput.TimeLimitSeconds
			if timeLimit == 0 {
				timeLimit = 20
			}
			qType := qInput.QuestionType
			if qType == "" {
				qType = quiz.QuestionTypeSingle
			}
			orderIdx := qInput.OrderIndex
			if orderIdx == 0 && qIdx > 0 {
				orderIdx = qIdx
			}

			qq := quiz.QuizQuestion{
				ID:               qID,
				QuizID:           id,
				Text:             qInput.Text,
				ImageURL:         qInput.ImageURL,
				QuestionType:     qType,
				TimeLimitSeconds: timeLimit,
				OrderIndex:       orderIdx,
				Answers:          make([]quiz.QuizAnswer, 0, len(qInput.Answers)),
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
			}

			for aIdx, aInput := range qInput.Answers {
				ansOrderIdx := aInput.OrderIndex
				if ansOrderIdx == 0 && aIdx > 0 {
					ansOrderIdx = aIdx
				}
				ans := quiz.QuizAnswer{
					ID:         quiz.ParseOrGenerateUUID(aInput.ID),
					QuestionID: qID,
					Text:       aInput.Text,
					IsCorrect:  aInput.IsCorrect,
					OrderIndex: ansOrderIdx,
					CreatedAt:  time.Now().UTC(),
				}
				qq.Answers = append(qq.Answers, ans)
			}

			qd.Questions = append(qd.Questions, qq)
		}
		qd.QuestionCount = len(qd.Questions)
	}
	return qd, nil
}

func (m *mockRepository) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleted[id] {
		return sql.ErrNoRows
	}

	if _, exists := m.quizzes[id]; !exists {
		return sql.ErrNoRows
	}

	m.deleted[id] = true
	return nil
}

func (m *mockRepository) DuplicateQuiz(ctx context.Context, originalID, newCreatorID uuid.UUID, newTitle string, includeAnswers bool) (*quiz.QuizDetail, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleted[originalID] {
		return nil, sql.ErrNoRows
	}

	orig, exists := m.quizzes[originalID]
	if !exists {
		return nil, sql.ErrNoRows
	}

	newID := uuid.New()
	copyQD := &quiz.QuizDetail{
		Quiz: quiz.Quiz{
			ID:            newID,
			CreatorID:     newCreatorID,
			CreatorName:   m.userNames[newCreatorID],
			Title:         newTitle,
			Description:   orig.Description,
			CoverImageURL: orig.CoverImageURL,
			Status:        quiz.StatusDraft,
			Tags:          orig.Tags,
			QuestionCount: len(orig.Questions),
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		Questions: make([]quiz.QuizQuestion, 0, len(orig.Questions)),
	}

	for _, q := range orig.Questions {
		newQID := uuid.New()
		newQ := quiz.QuizQuestion{
			ID:               newQID,
			QuizID:           newID,
			Text:             q.Text,
			ImageURL:         q.ImageURL,
			QuestionType:     q.QuestionType,
			TimeLimitSeconds: q.TimeLimitSeconds,
			OrderIndex:       q.OrderIndex,
			Answers:          make([]quiz.QuizAnswer, 0),
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		}

		if includeAnswers {
			for _, a := range q.Answers {
				newA := quiz.QuizAnswer{
					ID:         uuid.New(),
					QuestionID: newQID,
					Text:       a.Text,
					IsCorrect:  a.IsCorrect,
					OrderIndex: a.OrderIndex,
					CreatedAt:  time.Now().UTC(),
				}
				newQ.Answers = append(newQ.Answers, newA)
			}
		}

		copyQD.Questions = append(copyQD.Questions, newQ)
	}

	m.quizzes[newID] = copyQD
	return copyQD, nil
}
