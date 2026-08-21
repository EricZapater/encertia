package evaluation

import "time"

type Evaluation struct {
	ID              string     `json:"id"`
	QuizID          string     `json:"quizId"`
	StudentID       string     `json:"studentId"`
	CalculatedGrade float64    `json:"calculatedGrade"`
	FinalGrade      *float64   `json:"finalGrade,omitempty"`
	IsGraded        bool       `json:"isGraded"`
	GradedBy        *string    `json:"gradedBy,omitempty"`
	GradedAt        *time.Time `json:"gradedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type EvaluationQuizSummary struct {
	QuizID        string    `json:"quizId"`
	QuizTitle     string    `json:"quizTitle"`
	TotalMatches  int       `json:"totalMatches"`
	TotalStudents int       `json:"totalStudents"`
	GradedCount   int       `json:"gradedCount"`
	LastMatchAt   time.Time `json:"lastMatchAt"`
}

type AnswerDistributionItem struct {
	AnswerID   string  `json:"answerId"`
	AnswerText string  `json:"answerText"`
	IsCorrect  bool    `json:"isCorrect"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type QuestionStats struct {
	QuestionID         string                   `json:"questionId"`
	QuestionText       string                   `json:"questionText"`
	QuestionIndex      int                      `json:"questionIndex"`
	HitRate            float64                  `json:"hitRate"`
	AvgResponseTimeMs  int                      `json:"avgResponseTimeMs"`
	AnswerDistribution []AnswerDistributionItem `json:"answerDistribution"`
	NoAnswerCount      int                      `json:"noAnswerCount"`
}

type StudentEvaluationSummary struct {
	StudentID       string   `json:"studentId"`
	StudentName     string   `json:"studentName"`
	MatchesCount    int      `json:"matchesCount"`
	CalculatedGrade float64  `json:"calculatedGrade"`
	FinalGrade      *float64 `json:"finalGrade"`
	IsGraded        bool     `json:"isGraded"`
}

type QuizEvaluationResponse struct {
	QuizID       string                     `json:"quizId"`
	QuizTitle    string                     `json:"quizTitle"`
	TotalMatches int                        `json:"totalMatches"`
	Stats        []QuestionStats            `json:"stats"`
	Students     []StudentEvaluationSummary `json:"students"`
}

type StudentAnswerDetail struct {
	QuestionID        string   `json:"questionId"`
	QuestionText      string   `json:"questionText"`
	QuestionIndex     int      `json:"questionIndex"`
	SelectedAnswerIDs []string `json:"selectedAnswerIds"`
	CorrectAnswerIDs  []string `json:"correctAnswerIds"`
	IsCorrect         bool     `json:"isCorrect"`
	ResponseTimeMs    int      `json:"responseTimeMs"`
}

type StudentMatchResult struct {
	MatchID        string                `json:"matchId"`
	MatchDate      time.Time             `json:"matchDate"`
	Score          int                   `json:"score"`
	TotalQuestions int                   `json:"totalQuestions"`
	Answers        []StudentAnswerDetail `json:"answers"`
}

type StudentEvaluationDetail struct {
	EvaluationID    string               `json:"evaluationId"`
	StudentID       string               `json:"studentId"`
	StudentName     string               `json:"studentName"`
	CalculatedGrade float64              `json:"calculatedGrade"`
	FinalGrade      *float64             `json:"finalGrade"`
	IsGraded        bool                 `json:"isGraded"`
	GradedBy        *string              `json:"gradedBy"`
	GradedAt        *time.Time           `json:"gradedAt"`
	Matches         []StudentMatchResult `json:"matches"`
}

type GradeRequest struct {
	FinalGrade float64 `json:"finalGrade" binding:"required"`
}

type GradeResponse struct {
	EvaluationID    string    `json:"evaluationId"`
	CalculatedGrade float64   `json:"calculatedGrade"`
	FinalGrade      float64   `json:"finalGrade"`
	IsGraded        bool      `json:"isGraded"`
	GradedBy        string    `json:"gradedBy"`
	GradedAt        time.Time `json:"gradedAt"`
}
