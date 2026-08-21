package quiz_test

import (
	"context"
	"strings"
	"testing"

	"github.com/encertia/backend/internal/quiz"
	"github.com/google/uuid"
)

func setupQuizService() (quiz.Service, *mockRepository) {
	repo := newMockRepository()
	svc := quiz.NewService(repo)
	return svc, repo
}

// 1. Creation Tests
func TestCreateQuiz_Success_Draft(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()
	teacherID := uuid.New()

	desc := "Un quiz sobre comarques"
	cover := "https://pub-r2.encertia.cat/covers/cat.jpg"
	input := quiz.CreateQuizInput{
		Title:         "Geografia de Catalunya",
		Description:   &desc,
		CoverImageURL: &cover,
		Status:        quiz.StatusDraft,
		Tags:          []string{"geografia", "catalunya"},
	}

	res, appErr := svc.CreateQuiz(ctx, teacherID, "teacher", input)
	if appErr != nil {
		t.Fatalf("unexpected error creating quiz: %v", appErr)
	}

	if res.Title != "Geografia de Catalunya" {
		t.Errorf("expected title Geografia de Catalunya, got %s", res.Title)
	}
	if res.CreatorID != teacherID {
		t.Errorf("expected creatorID %s, got %s", teacherID, res.CreatorID)
	}
	if res.Status != quiz.StatusDraft {
		t.Errorf("expected draft status, got %s", res.Status)
	}
	if len(res.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(res.Tags))
	}
}

func TestCreateQuiz_Success_PublishedWithValidQuestions(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()
	teacherID := uuid.New()

	input := quiz.CreateQuizInput{
		Title:  "Història Medieval",
		Status: quiz.StatusPublished,
		Tags:   []string{"historia"},
		Questions: []quiz.SaveQuestionInput{
			{
				Text:             "En quin any va morir Jaume I?",
				QuestionType:     quiz.QuestionTypeSingle,
				TimeLimitSeconds: 30,
				OrderIndex:       0,
				Answers: []quiz.SaveAnswerInput{
					{Text: "1276", IsCorrect: true, OrderIndex: 0},
					{Text: "1213", IsCorrect: false, OrderIndex: 1},
					{Text: "1348", IsCorrect: false, OrderIndex: 2},
					{Text: "1492", IsCorrect: false, OrderIndex: 3},
				},
			},
			{
				Text:             "Quins d'aquests van ser comtats catalans?",
				QuestionType:     quiz.QuestionTypeMultiple,
				TimeLimitSeconds: 20,
				OrderIndex:       1,
				Answers: []quiz.SaveAnswerInput{
					{Text: "Comtat d'Urgell", IsCorrect: true, OrderIndex: 0},
					{Text: "Comtat de Barcelona", IsCorrect: true, OrderIndex: 1},
					{Text: "Comtat de Wessex", IsCorrect: false, OrderIndex: 2},
				},
			},
		},
	}

	res, appErr := svc.CreateQuiz(ctx, teacherID, "teacher", input)
	if appErr != nil {
		t.Fatalf("unexpected error creating published quiz: %v", appErr)
	}

	if res.Status != quiz.StatusPublished {
		t.Errorf("expected status published, got %s", res.Status)
	}
	if len(res.Questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(res.Questions))
	}
	if len(res.Questions[0].Answers) != 4 {
		t.Errorf("expected 4 answers in Q1, got %d", len(res.Questions[0].Answers))
	}
}

func TestCreateQuiz_ValidationFailures(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()
	userID := uuid.New()

	longTitle := strings.Repeat("A", 151)
	longDesc := strings.Repeat("D", 1001)

	tests := []struct {
		name  string
		input quiz.CreateQuizInput
	}{
		{
			name:  "short title",
			input: quiz.CreateQuizInput{Title: "AB"},
		},
		{
			name:  "long title",
			input: quiz.CreateQuizInput{Title: longTitle},
		},
		{
			name:  "long description",
			input: quiz.CreateQuizInput{Title: "Valid Title", Description: &longDesc},
		},
		{
			name:  "invalid status",
			input: quiz.CreateQuizInput{Title: "Valid Title", Status: "unknown_status"},
		},
		{
			name:  "published without questions",
			input: quiz.CreateQuizInput{Title: "Valid Title", Status: quiz.StatusPublished, Questions: []quiz.SaveQuestionInput{}},
		},
		{
			name: "single_choice with 0 correct answers",
			input: quiz.CreateQuizInput{
				Title:  "Valid Title",
				Status: quiz.StatusPublished,
				Questions: []quiz.SaveQuestionInput{
					{
						Text:             "Question text?",
						QuestionType:     quiz.QuestionTypeSingle,
						TimeLimitSeconds: 20,
						Answers: []quiz.SaveAnswerInput{
							{Text: "Ans 1", IsCorrect: false},
							{Text: "Ans 2", IsCorrect: false},
						},
					},
				},
			},
		},
		{
			name: "single_choice with 2 correct answers",
			input: quiz.CreateQuizInput{
				Title:  "Valid Title",
				Status: quiz.StatusPublished,
				Questions: []quiz.SaveQuestionInput{
					{
						Text:             "Question text?",
						QuestionType:     quiz.QuestionTypeSingle,
						TimeLimitSeconds: 20,
						Answers: []quiz.SaveAnswerInput{
							{Text: "Ans 1", IsCorrect: true},
							{Text: "Ans 2", IsCorrect: true},
						},
					},
				},
			},
		},
		{
			name: "multiple_choice with 0 correct answers",
			input: quiz.CreateQuizInput{
				Title:  "Valid Title",
				Status: quiz.StatusPublished,
				Questions: []quiz.SaveQuestionInput{
					{
						Text:             "Question text?",
						QuestionType:     quiz.QuestionTypeMultiple,
						TimeLimitSeconds: 20,
						Answers: []quiz.SaveAnswerInput{
							{Text: "Ans 1", IsCorrect: false},
							{Text: "Ans 2", IsCorrect: false},
						},
					},
				},
			},
		},
		{
			name: "published question with less than 2 answers",
			input: quiz.CreateQuizInput{
				Title:  "Valid Title",
				Status: quiz.StatusPublished,
				Questions: []quiz.SaveQuestionInput{
					{
						Text:             "Question text?",
						QuestionType:     quiz.QuestionTypeSingle,
						TimeLimitSeconds: 20,
						Answers: []quiz.SaveAnswerInput{
							{Text: "Ans 1", IsCorrect: true},
						},
					},
				},
			},
		},
		{
			name: "published question with more than 6 answers",
			input: quiz.CreateQuizInput{
				Title:  "Valid Title",
				Status: quiz.StatusPublished,
				Questions: []quiz.SaveQuestionInput{
					{
						Text:             "Question text?",
						QuestionType:     quiz.QuestionTypeSingle,
						TimeLimitSeconds: 20,
						Answers: []quiz.SaveAnswerInput{
							{Text: "Ans 1", IsCorrect: true},
							{Text: "Ans 2", IsCorrect: false},
							{Text: "Ans 3", IsCorrect: false},
							{Text: "Ans 4", IsCorrect: false},
							{Text: "Ans 5", IsCorrect: false},
							{Text: "Ans 6", IsCorrect: false},
							{Text: "Ans 7", IsCorrect: false},
						},
					},
				},
			},
		},
		{
			name: "invalid time limit",
			input: quiz.CreateQuizInput{
				Title:  "Valid Title",
				Status: quiz.StatusPublished,
				Questions: []quiz.SaveQuestionInput{
					{
						Text:             "Question text?",
						QuestionType:     quiz.QuestionTypeSingle,
						TimeLimitSeconds: 45, // invalid time limit
						Answers: []quiz.SaveAnswerInput{
							{Text: "Ans 1", IsCorrect: true},
							{Text: "Ans 2", IsCorrect: false},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appErr := svc.CreateQuiz(ctx, userID, "teacher", tt.input)
			if appErr == nil {
				t.Fatalf("expected validation error for test case '%s', got nil", tt.name)
			}
			if appErr.StatusCode != 400 {
				t.Errorf("expected status 400, got %d", appErr.StatusCode)
			}
		})
	}
}

// 2. List Quizzes & RBAC Tests
func TestListQuizzes_RBACAndFilters(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()

	adminID := uuid.New()
	teacher1ID := uuid.New()
	teacher2ID := uuid.New()

	// Teacher 1 creates 3 quizzes
	_, _ = svc.CreateQuiz(ctx, teacher1ID, "teacher", quiz.CreateQuizInput{
		Title:  "Matemàtiques 1",
		Status: quiz.StatusDraft,
		Tags:   []string{"mates", "eso"},
	})
	_, _ = svc.CreateQuiz(ctx, teacher1ID, "teacher", quiz.CreateQuizInput{
		Title:  "Matemàtiques 2",
		Status: quiz.StatusPublished,
		Tags:   []string{"mates", "batxillerat"},
		Questions: []quiz.SaveQuestionInput{
			{
				Text:             "Quant és 2+2?",
				TimeLimitSeconds: 10,
				QuestionType:     quiz.QuestionTypeSingle,
				Answers: []quiz.SaveAnswerInput{
					{Text: "4", IsCorrect: true},
					{Text: "5", IsCorrect: false},
				},
			},
		},
	})
	_, _ = svc.CreateQuiz(ctx, teacher1ID, "teacher", quiz.CreateQuizInput{
		Title:  "Física Quàntica",
		Status: quiz.StatusArchived,
		Tags:   []string{"fisica"},
	})

	// Teacher 2 creates 2 quizzes
	_, _ = svc.CreateQuiz(ctx, teacher2ID, "teacher", quiz.CreateQuizInput{
		Title:  "Literatura Catalana",
		Status: quiz.StatusDraft,
		Tags:   []string{"literatura"},
	})
	_, _ = svc.CreateQuiz(ctx, teacher2ID, "teacher", quiz.CreateQuizInput{
		Title:  "Història de l'Art",
		Status: quiz.StatusDraft,
		Tags:   []string{"art"},
	})

	// 1. Teacher 1 only sees their own quizzes (3 items)
	resT1, appErrT1 := svc.ListQuizzes(ctx, teacher1ID, "teacher", quiz.QuizListFilters{})
	if appErrT1 != nil {
		t.Fatalf("unexpected error listing for teacher1: %v", appErrT1)
	}
	if resT1.Pagination.TotalCount != 3 {
		t.Errorf("teacher1 expected 3 quizzes, got %d", resT1.Pagination.TotalCount)
	}

	// 2. Teacher 2 only sees their own quizzes (2 items)
	resT2, appErrT2 := svc.ListQuizzes(ctx, teacher2ID, "teacher", quiz.QuizListFilters{})
	if appErrT2 != nil {
		t.Fatalf("unexpected error listing for teacher2: %v", appErrT2)
	}
	if resT2.Pagination.TotalCount != 2 {
		t.Errorf("teacher2 expected 2 quizzes, got %d", resT2.Pagination.TotalCount)
	}

	// 3. Admin sees all quizzes (5 items)
	resAdmin, appErrAdmin := svc.ListQuizzes(ctx, adminID, "admin", quiz.QuizListFilters{})
	if appErrAdmin != nil {
		t.Fatalf("unexpected error listing for admin: %v", appErrAdmin)
	}
	if resAdmin.Pagination.TotalCount != 5 {
		t.Errorf("admin expected 5 quizzes, got %d", resAdmin.Pagination.TotalCount)
	}

	// 4. Filter by status 'published'
	resStatus, _ := svc.ListQuizzes(ctx, teacher1ID, "teacher", quiz.QuizListFilters{Status: "published"})
	if resStatus.Pagination.TotalCount != 1 {
		t.Errorf("expected 1 published quiz for teacher1, got %d", resStatus.Pagination.TotalCount)
	}

	// 5. Filter by tag 'mates'
	resTag, _ := svc.ListQuizzes(ctx, teacher1ID, "teacher", quiz.QuizListFilters{Tag: "mates"})
	if resTag.Pagination.TotalCount != 2 {
		t.Errorf("expected 2 quizzes with tag mates for teacher1, got %d", resTag.Pagination.TotalCount)
	}

	// 6. Search filter
	resSearch, _ := svc.ListQuizzes(ctx, teacher1ID, "teacher", quiz.QuizListFilters{Search: "Física"})
	if resSearch.Pagination.TotalCount != 1 {
		t.Errorf("expected 1 match for search 'Física', got %d", resSearch.Pagination.TotalCount)
	}
}

// 3. GetQuizByID & Permissions
func TestGetQuizByID_Permissions(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()

	adminID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	created, err := svc.CreateQuiz(ctx, ownerID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz de Prova",
		Status: quiz.StatusDraft,
	})
	if err != nil {
		t.Fatalf("unexpected error creating quiz: %v", err)
	}

	// 1. Owner can view quiz -> 200
	qOwner, appErrOwner := svc.GetQuizByID(ctx, ownerID, "teacher", created.ID)
	if appErrOwner != nil {
		t.Fatalf("owner failed to get quiz: %v", appErrOwner)
	}
	if qOwner.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, qOwner.ID)
	}

	// 2. Admin can view quiz -> 200
	qAdmin, appErrAdmin := svc.GetQuizByID(ctx, adminID, "admin", created.ID)
	if appErrAdmin != nil {
		t.Fatalf("admin failed to get quiz: %v", appErrAdmin)
	}
	if qAdmin.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, qAdmin.ID)
	}

	// 3. Another user cannot view quiz -> 403
	_, appErrOther := svc.GetQuizByID(ctx, otherUserID, "teacher", created.ID)
	if appErrOther == nil || appErrOther.StatusCode != 403 {
		t.Errorf("expected 403 forbidden for other user, got %v", appErrOther)
	}

	// 4. Non-existent quiz -> 404
	_, appErr404 := svc.GetQuizByID(ctx, adminID, "admin", uuid.New())
	if appErr404 == nil || appErr404.StatusCode != 404 {
		t.Errorf("expected 404 for non-existent quiz, got %v", appErr404)
	}
}

// 4. UpdateQuiz Tests
func TestUpdateQuiz_PermissionsAndModifications(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()

	adminID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	created, _ := svc.CreateQuiz(ctx, ownerID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz Inicial",
		Status: quiz.StatusDraft,
	})

	// 1. Other user updating -> 403
	newTitle := "Títol No Autoritzat"
	_, appErrOther := svc.UpdateQuiz(ctx, otherUserID, "teacher", created.ID, quiz.UpdateQuizInput{
		Title: newTitle,
	})
	if appErrOther == nil || appErrOther.StatusCode != 403 {
		t.Errorf("expected 403 when other user updates quiz, got %v", appErrOther)
	}

	// 2. Owner updates quiz and publishes with valid question -> 200
	newStatus := quiz.StatusPublished
	updatedTitle := "Quiz Actualitzat"
	updated, appErrOwner := svc.UpdateQuiz(ctx, ownerID, "teacher", created.ID, quiz.UpdateQuizInput{
		Title:  updatedTitle,
		Status: &newStatus,
		Questions: []quiz.SaveQuestionInput{
			{
				Text:             "Nova pregunta?",
				TimeLimitSeconds: 20,
				QuestionType:     quiz.QuestionTypeSingle,
				Answers: []quiz.SaveAnswerInput{
					{Text: "Correcta", IsCorrect: true},
					{Text: "Incorrecta", IsCorrect: false},
				},
			},
		},
	})
	if appErrOwner != nil {
		t.Fatalf("unexpected error when owner updates quiz: %v", appErrOwner)
	}
	if updated.Title != "Quiz Actualitzat" {
		t.Errorf("expected updated title, got %s", updated.Title)
	}
	if updated.Status != quiz.StatusPublished {
		t.Errorf("expected published status, got %s", updated.Status)
	}
	if len(updated.Questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(updated.Questions))
	}

	// 3. Admin can also update quiz -> 200
	adminTitle := "Quiz Modificat per Admin"
	updatedAdmin, appErrAdmin := svc.UpdateQuiz(ctx, adminID, "admin", created.ID, quiz.UpdateQuizInput{
		Title: adminTitle,
	})
	if appErrAdmin != nil {
		t.Fatalf("unexpected error when admin updates quiz: %v", appErrAdmin)
	}
	if updatedAdmin.Title != adminTitle {
		t.Errorf("expected admin title, got %s", updatedAdmin.Title)
	}
}

// 5. DeleteQuiz (Soft-Delete) Tests
func TestDeleteQuiz_PermissionsAndSoftDelete(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()

	adminID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	created, _ := svc.CreateQuiz(ctx, ownerID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz per Eliminar",
		Status: quiz.StatusDraft,
	})

	// 1. Other user trying to delete -> 403
	_, appErrOther := svc.DeleteQuiz(ctx, otherUserID, "teacher", created.ID)
	if appErrOther == nil || appErrOther.StatusCode != 403 {
		t.Errorf("expected 403 when other user deletes quiz, got %v", appErrOther)
	}

	// 2. Owner deletes quiz -> 200
	msg, appErrOwner := svc.DeleteQuiz(ctx, ownerID, "teacher", created.ID)
	if appErrOwner != nil {
		t.Fatalf("unexpected error when owner deletes quiz: %v", appErrOwner)
	}
	if msg.Message != "Qüestionari eliminat correctament" {
		t.Errorf("expected success message, got %s", msg.Message)
	}

	// 3. Quiz is now not found (404)
	_, appErrGet := svc.GetQuizByID(ctx, ownerID, "teacher", created.ID)
	if appErrGet == nil || appErrGet.StatusCode != 404 {
		t.Errorf("expected 404 after deletion, got %v", appErrGet)
	}

	// 4. Admin deleting non-existent / already deleted quiz -> 404
	_, appErrAdmin := svc.DeleteQuiz(ctx, adminID, "admin", created.ID)
	if appErrAdmin == nil || appErrAdmin.StatusCode != 404 {
		t.Errorf("expected 404 deleting already deleted quiz, got %v", appErrAdmin)
	}
}

// 6. DuplicateQuiz Tests (with and without answers, default vs custom title)
func TestDuplicateQuiz_DefaultWithoutAnswers(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()
	ownerID := uuid.New()

	original, _ := svc.CreateQuiz(ctx, ownerID, "teacher", quiz.CreateQuizInput{
		Title:  "Capitales d'Europa",
		Status: quiz.StatusPublished,
		Tags:   []string{"geografia", "europa"},
		Questions: []quiz.SaveQuestionInput{
			{
				Text:             "Capital de França?",
				QuestionType:     quiz.QuestionTypeSingle,
				TimeLimitSeconds: 20,
				OrderIndex:       0,
				Answers: []quiz.SaveAnswerInput{
					{Text: "París", IsCorrect: true, OrderIndex: 0},
					{Text: "Lió", IsCorrect: false, OrderIndex: 1},
				},
			},
			{
				Text:             "Capital d'Itàlia?",
				QuestionType:     quiz.QuestionTypeSingle,
				TimeLimitSeconds: 30,
				OrderIndex:       1,
				Answers: []quiz.SaveAnswerInput{
					{Text: "Roma", IsCorrect: true, OrderIndex: 0},
					{Text: "Milà", IsCorrect: false, OrderIndex: 1},
				},
			},
		},
	})

	// Duplicate with includeAnswers = false (default)
	duplicated, appErr := svc.DuplicateQuiz(ctx, ownerID, "teacher", original.ID, quiz.DuplicateQuizInput{
		IncludeAnswers: false,
	})
	if appErr != nil {
		t.Fatalf("unexpected error duplicating quiz: %v", appErr)
	}

	if duplicated.ID == original.ID {
		t.Error("expected duplicated quiz to have a new ID")
	}
	if duplicated.Title != "[Còpia] Capitales d'Europa" {
		t.Errorf("expected title '[Còpia] Capitales d'Europa', got '%s'", duplicated.Title)
	}
	if duplicated.Status != quiz.StatusDraft {
		t.Errorf("expected duplicated quiz status to be draft, got %s", duplicated.Status)
	}
	if len(duplicated.Questions) != 2 {
		t.Fatalf("expected 2 duplicated questions, got %d", len(duplicated.Questions))
	}

	// CRITICAL: When includeAnswers = false, questions must have NO answers!
	if len(duplicated.Questions[0].Answers) != 0 {
		t.Errorf("expected Q1 to have 0 answers when includeAnswers=false, got %d", len(duplicated.Questions[0].Answers))
	}
	if len(duplicated.Questions[1].Answers) != 0 {
		t.Errorf("expected Q2 to have 0 answers when includeAnswers=false, got %d", len(duplicated.Questions[1].Answers))
	}
	if duplicated.Questions[0].Text != "Capital de França?" {
		t.Errorf("expected question text preserved, got %s", duplicated.Questions[0].Text)
	}
}

func TestDuplicateQuiz_WithAnswersAndCustomTitle(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()
	ownerID := uuid.New()

	original, _ := svc.CreateQuiz(ctx, ownerID, "teacher", quiz.CreateQuizInput{
		Title:  "Química Orgànica",
		Status: quiz.StatusPublished,
		Tags:   []string{"quimica"},
		Questions: []quiz.SaveQuestionInput{
			{
				Text:             "Fórmula del metà?",
				QuestionType:     quiz.QuestionTypeSingle,
				TimeLimitSeconds: 20,
				OrderIndex:       0,
				Answers: []quiz.SaveAnswerInput{
					{Text: "CH4", IsCorrect: true, OrderIndex: 0},
					{Text: "C2H6", IsCorrect: false, OrderIndex: 1},
				},
			},
		},
	})

	customTitle := "Química Orgànica (Grup B)"
	duplicated, appErr := svc.DuplicateQuiz(ctx, ownerID, "teacher", original.ID, quiz.DuplicateQuizInput{
		IncludeAnswers: true,
		Title:          &customTitle,
	})
	if appErr != nil {
		t.Fatalf("unexpected error duplicating quiz with answers: %v", appErr)
	}

	if duplicated.Title != customTitle {
		t.Errorf("expected title '%s', got '%s'", customTitle, duplicated.Title)
	}
	if len(duplicated.Questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(duplicated.Questions))
	}
	if len(duplicated.Questions[0].Answers) != 2 {
		t.Fatalf("expected 2 answers when includeAnswers=true, got %d", len(duplicated.Questions[0].Answers))
	}
	if !duplicated.Questions[0].Answers[0].IsCorrect || duplicated.Questions[0].Answers[0].Text != "CH4" {
		t.Errorf("expected first answer to be CH4 and correct")
	}
}

func TestDuplicateQuiz_Permissions(t *testing.T) {
	svc, _ := setupQuizService()
	ctx := context.Background()

	adminID := uuid.New()
	ownerID := uuid.New()
	otherUserID := uuid.New()

	original, _ := svc.CreateQuiz(ctx, ownerID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz de Prova",
		Status: quiz.StatusDraft,
	})

	// 1. Other user duplicate attempt -> 403
	_, appErrOther := svc.DuplicateQuiz(ctx, otherUserID, "teacher", original.ID, quiz.DuplicateQuizInput{})
	if appErrOther == nil || appErrOther.StatusCode != 403 {
		t.Errorf("expected 403 when unauthorized user duplicates quiz, got %v", appErrOther)
	}

	// 2. Admin duplicates other user's quiz -> 201
	resAdmin, appErrAdmin := svc.DuplicateQuiz(ctx, adminID, "admin", original.ID, quiz.DuplicateQuizInput{})
	if appErrAdmin != nil {
		t.Fatalf("unexpected error when admin duplicates quiz: %v", appErrAdmin)
	}
	if resAdmin.CreatorID != adminID {
		t.Errorf("expected duplicated quiz creator to be admin (%s), got %s", adminID, resAdmin.CreatorID)
	}
}
