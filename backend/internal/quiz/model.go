package quiz

import (
	"time"

	"github.com/google/uuid"
)

// QuizStatus defines the state of a quiz.
type QuizStatus string

const (
	StatusDraft     QuizStatus = "draft"
	StatusPublished QuizStatus = "published"
	StatusArchived  QuizStatus = "archived"
)

// QuestionType defines the format of choices for a quiz question.
type QuestionType string

const (
	QuestionTypeSingle   QuestionType = "single_choice"
	QuestionTypeMultiple QuestionType = "multiple_choice"
)

// ValidTimeLimits contains the permitted question timer durations in seconds.
var ValidTimeLimits = map[int]bool{
	5:   true,
	10:  true,
	20:  true,
	30:  true,
	60:  true,
	90:  true,
	120: true,
}

// QuizAnswer represents a single choice answer to a question.
type QuizAnswer struct {
	ID         uuid.UUID `json:"id"`
	QuestionID uuid.UUID `json:"-"`
	Text       string    `json:"text"`
	IsCorrect  bool      `json:"isCorrect"`
	OrderIndex int       `json:"orderIndex"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
}

// QuizQuestion represents a question belonging to a quiz.
type QuizQuestion struct {
	ID               uuid.UUID    `json:"id"`
	QuizID           uuid.UUID    `json:"-"`
	Text             string       `json:"text"`
	ImageURL         *string      `json:"imageUrl"`
	QuestionType     QuestionType `json:"questionType"`
	TimeLimitSeconds int          `json:"timeLimitSeconds"`
	OrderIndex       int          `json:"orderIndex"`
	Answers          []QuizAnswer `json:"answers"`
	CreatedAt        time.Time    `json:"createdAt,omitempty"`
	UpdatedAt        time.Time    `json:"updatedAt,omitempty"`
}

// Quiz represents a quiz metadata entity.
type Quiz struct {
	ID            uuid.UUID  `json:"id"`
	CreatorID     uuid.UUID  `json:"creatorId"`
	CreatorName   string     `json:"creatorName,omitempty"`
	Title         string     `json:"title"`
	Description   *string    `json:"description"`
	CoverImageURL *string    `json:"coverImageUrl"`
	Status        QuizStatus `json:"status"`
	Tags          []string   `json:"tags"`
	QuestionCount int        `json:"questionCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	DeletedAt     *time.Time `json:"-"`
}

// QuizDetail represents the full quiz entity including its questions and answers.
type QuizDetail struct {
	Quiz
	Questions []QuizQuestion `json:"questions"`
}

// SaveAnswerInput is used when creating or updating question answers.
type SaveAnswerInput struct {
	ID         *uuid.UUID `json:"id,omitempty"`
	Text       string     `json:"text"`
	IsCorrect  bool       `json:"isCorrect"`
	OrderIndex int        `json:"orderIndex"`
}

// SaveQuestionInput is used when creating or updating quiz questions.
type SaveQuestionInput struct {
	ID               *uuid.UUID        `json:"id,omitempty"`
	Text             string            `json:"text"`
	ImageURL         *string           `json:"imageUrl,omitempty"`
	QuestionType     QuestionType      `json:"questionType"`
	TimeLimitSeconds int               `json:"timeLimitSeconds"`
	OrderIndex       int               `json:"orderIndex"`
	Answers          []SaveAnswerInput `json:"answers"`
}

// CreateQuizInput represents payload to create a new quiz.
type CreateQuizInput struct {
	Title         string              `json:"title"`
	Description   *string             `json:"description,omitempty"`
	CoverImageURL *string             `json:"coverImageUrl,omitempty"`
	Status        QuizStatus          `json:"status,omitempty"`
	Tags          []string            `json:"tags,omitempty"`
	Questions     []SaveQuestionInput `json:"questions,omitempty"`
}

// UpdateQuizInput represents payload to update an existing quiz.
type UpdateQuizInput struct {
	Title         string              `json:"title"`
	Description   *string             `json:"description,omitempty"`
	CoverImageURL *string             `json:"coverImageUrl,omitempty"`
	Status        *QuizStatus         `json:"status,omitempty"`
	Tags          []string            `json:"tags,omitempty"`
	Questions     []SaveQuestionInput `json:"questions,omitempty"`
}

// DuplicateQuizInput represents payload when duplicating a quiz.
type DuplicateQuizInput struct {
	IncludeAnswers bool    `json:"includeAnswers"`
	Title          *string `json:"title,omitempty"`
}

// QuizListFilters contains filtering and pagination options for quizzes.
type QuizListFilters struct {
	Page      int
	PageSize  int
	Search    string
	Status    string
	Tag       string
	CreatorID *uuid.UUID
}

// QuizPagination holds pagination metadata.
type QuizPagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	TotalPages int `json:"totalPages"`
}

// QuizListResponse represents the paginated response for quizzes.
type QuizListResponse struct {
	Items      []Quiz         `json:"items"`
	Pagination QuizPagination `json:"pagination"`
}
