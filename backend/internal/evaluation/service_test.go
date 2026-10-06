package evaluation_test

import (
	"testing"
	"time"

	"github.com/encertia/backend/internal/evaluation"
	"github.com/google/uuid"
)

type mockRepository struct {
	evaluations map[string]*evaluation.Evaluation
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		evaluations: make(map[string]*evaluation.Evaluation),
	}
}

func (m *mockRepository) ListEvaluations(teacherID string, isAdmin bool, groupID string) ([]evaluation.EvaluationQuizSummary, error) {
	if groupID == "filtered-empty-group" {
		return []evaluation.EvaluationQuizSummary{}, nil
	}
	return []evaluation.EvaluationQuizSummary{
		{
			QuizID:        uuid.New().String(),
			QuizTitle:     "Test Quiz",
			TotalMatches:  1,
			TotalStudents: 2,
			GradedCount:   1,
			LastMatchAt:   time.Now(),
		},
	}, nil
}

func (m *mockRepository) GetQuizEvaluation(quizID string, groupID string) (*evaluation.QuizEvaluationResponse, error) {
	grpID := groupID
	grpName := "Grup 1A"
	if groupID == "" {
		grpID = "group-1"
	}

	return &evaluation.QuizEvaluationResponse{
		QuizID:       quizID,
		QuizTitle:    "Test Quiz",
		TotalMatches: 1,
		Students: []evaluation.StudentEvaluationSummary{
			{
				StudentID:       "user-1",
				StudentName:     "Registered User",
				GroupID:         &grpID,
				GroupName:       &grpName,
				MatchesCount:    1,
				CalculatedGrade: 8.5,
				IsGraded:        true,
			},
			{
				StudentID:       "player-anon-1",
				StudentName:     "AnonPlayer",
				GroupID:         &grpID,
				GroupName:       &grpName,
				MatchesCount:    1,
				CalculatedGrade: 7.0,
				IsGraded:        false,
			},
		},
	}, nil
}

func (m *mockRepository) GetStudentEvaluation(quizID, studentID string) (*evaluation.StudentEvaluationDetail, error) {
	grpID := "group-1"
	grpName := "Grup 1A"
	return &evaluation.StudentEvaluationDetail{
		EvaluationID:    "eval-1",
		StudentID:       studentID,
		StudentName:     "Test Student",
		GroupID:         &grpID,
		GroupName:       &grpName,
		CalculatedGrade: 8.5,
		IsGraded:        false,
	}, nil
}

func (m *mockRepository) GradeStudent(quizID, studentID, teacherID string, finalGrade float64) (*evaluation.GradeResponse, error) {
	now := time.Now()
	key := quizID + ":" + studentID
	m.evaluations[key] = &evaluation.Evaluation{
		ID:              "eval-" + studentID,
		QuizID:          quizID,
		CalculatedGrade: 8.0,
		FinalGrade:      &finalGrade,
		IsGraded:        true,
		GradedBy:        &teacherID,
		GradedAt:        &now,
	}

	return &evaluation.GradeResponse{
		EvaluationID:    "eval-" + studentID,
		CalculatedGrade: 8.0,
		FinalGrade:      finalGrade,
		IsGraded:        true,
		GradedBy:        teacherID,
		GradedAt:        now,
	}, nil
}

func (m *mockRepository) UpsertCalculatedGradeForMatch(matchID string) error {
	return nil
}

func TestEvaluationService_ListAndGrade(t *testing.T) {
	repo := newMockRepository()
	svc := evaluation.NewService(repo, nil)

	// Test student role gets unauthorized
	_, err := svc.ListEvaluations("user-1", "student", "")
	if err == nil || err != evaluation.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized for student role, got: %v", err)
	}

	// Test admin role can list evaluations without group filter
	summaries, err := svc.ListEvaluations("admin-1", "admin", "")
	if err != nil {
		t.Fatalf("unexpected error listing evaluations: %v", err)
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}

	// Test admin role listing with group filter
	summariesGrp, err := svc.ListEvaluations("admin-1", "admin", "group-123")
	if err != nil {
		t.Fatalf("unexpected error listing evaluations with group filter: %v", err)
	}
	if len(summariesGrp) != 1 {
		t.Errorf("expected 1 summary for group-123, got %d", len(summariesGrp))
	}

	// Test get quiz evaluation with group filter
	quizID := uuid.New().String()
	quizEval, err := svc.GetQuizEvaluation(quizID, "admin-1", "admin", "group-123")
	if err != nil {
		t.Fatalf("unexpected error getting quiz evaluation with group filter: %v", err)
	}
	if len(quizEval.Students) != 2 {
		t.Errorf("expected 2 students, got %d", len(quizEval.Students))
	}
	if quizEval.Students[0].GroupID == nil || *quizEval.Students[0].GroupID != "group-123" {
		t.Errorf("expected student groupID to be group-123, got: %v", quizEval.Students[0].GroupID)
	}

	// Test grading registered user as admin
	studentID := uuid.New().String()
	teacherID := uuid.New().String()

	gradeRes, err := svc.GradeStudent(quizID, studentID, teacherID, "admin", 9.5)
	if err != nil {
		t.Fatalf("unexpected error grading student: %v", err)
	}
	if gradeRes.FinalGrade != 9.5 || !gradeRes.IsGraded {
		t.Errorf("unexpected grade response: %+v", gradeRes)
	}

	// Test grading anonymous player (player UUID) as admin
	anonPlayerID := uuid.New().String()
	anonGradeRes, err := svc.GradeStudent(quizID, anonPlayerID, teacherID, "admin", 7.25)
	if err != nil {
		t.Fatalf("unexpected error grading anonymous player: %v", err)
	}
	if anonGradeRes.FinalGrade != 7.25 || !anonGradeRes.IsGraded {
		t.Errorf("unexpected grade response for anonymous player: %+v", anonGradeRes)
	}

	// Test invalid grade (< 0 or > 10)
	_, err = svc.GradeStudent(quizID, studentID, teacherID, "admin", 11.0)
	if err == nil || err != evaluation.ErrInvalidGrade {
		t.Errorf("expected ErrInvalidGrade, got: %v", err)
	}
}
