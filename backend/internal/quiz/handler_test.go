package quiz_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/encertia/backend/internal/quiz"
	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestQuizRouter(svc quiz.Service, storageSvc shared.StorageService, actorID uuid.UUID, actorRole, actorEmail string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := quiz.NewHandler(svc, storageSvc)

	authMiddleware := func(c *gin.Context) {
		if actorID != uuid.Nil {
			c.Set(shared.CtxKeyUserID, actorID.String())
			c.Set(shared.CtxKeyUserEmail, actorEmail)
			c.Set(shared.CtxKeyUserRole, actorRole)
		}
		c.Next()
	}

	handler.RegisterRoutes(router.Group(""), authMiddleware)
	return router
}

func TestHTTP_ListQuizzes(t *testing.T) {
	svc, _ := setupQuizService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestQuizRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	ctx := t.Context()
	_, _ = svc.CreateQuiz(ctx, teacherID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz 1",
		Status: quiz.StatusDraft,
	})

	req, _ := http.NewRequest(http.MethodGet, "/quizzes?page=1&pageSize=12", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res quiz.QuizListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if res.Pagination.TotalCount != 1 {
		t.Errorf("expected 1 quiz, got %d", res.Pagination.TotalCount)
	}
}

func TestHTTP_CreateQuiz(t *testing.T) {
	svc, _ := setupQuizService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestQuizRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	body, _ := json.Marshal(quiz.CreateQuizInput{
		Title:  "Nou Quiz Creat",
		Status: quiz.StatusDraft,
		Tags:   []string{"cat", "prova"},
	})

	req, _ := http.NewRequest(http.MethodPost, "/quizzes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var res quiz.QuizDetail
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode quiz detail: %v", err)
	}
	if res.Title != "Nou Quiz Creat" {
		t.Errorf("expected title Nou Quiz Creat, got %s", res.Title)
	}
}

func TestHTTP_GetQuizByID(t *testing.T) {
	svc, _ := setupQuizService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestQuizRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	ctx := t.Context()
	created, _ := svc.CreateQuiz(ctx, teacherID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz Detall",
		Status: quiz.StatusDraft,
	})

	// 1. Success 200
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/quizzes/%s", created.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// 2. Invalid UUID 400
	reqInv, _ := http.NewRequest(http.MethodGet, "/quizzes/not-a-uuid", nil)
	wInv := httptest.NewRecorder()
	router.ServeHTTP(wInv, reqInv)
	if wInv.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", wInv.Code)
	}

	// 3. Not Found 404
	req404, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/quizzes/%s", uuid.New()), nil)
	w404 := httptest.NewRecorder()
	router.ServeHTTP(w404, req404)
	if w404.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing quiz, got %d", w404.Code)
	}
}

func TestHTTP_UpdateQuiz(t *testing.T) {
	svc, _ := setupQuizService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestQuizRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	ctx := t.Context()
	created, _ := svc.CreateQuiz(ctx, teacherID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz per Actualitzar",
		Status: quiz.StatusDraft,
	})

	body, _ := json.Marshal(quiz.UpdateQuizInput{
		Title: "Quiz Amb Títol Modificat",
	})

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/quizzes/%s", created.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for update, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTP_DeleteQuiz(t *testing.T) {
	svc, _ := setupQuizService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestQuizRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	ctx := t.Context()
	created, _ := svc.CreateQuiz(ctx, teacherID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz per Eliminar HTTP",
		Status: quiz.StatusDraft,
	})

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/quizzes/%s", created.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for delete, got %d", w.Code)
	}
}

func TestHTTP_DuplicateQuiz(t *testing.T) {
	svc, _ := setupQuizService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestQuizRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	ctx := t.Context()
	created, _ := svc.CreateQuiz(ctx, teacherID, "teacher", quiz.CreateQuizInput{
		Title:  "Quiz Original",
		Status: quiz.StatusDraft,
	})

	// 1. Duplicate without body
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/quizzes/%s/duplicate", created.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for duplicate, got %d: %s", w.Code, w.Body.String())
	}

	var res quiz.QuizDetail
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.Title != "[Còpia] Quiz Original" {
		t.Errorf("expected [Còpia] Quiz Original, got %s", res.Title)
	}

	// 2. Duplicate with custom title
	customTitle := "Quiz Duplicat Personalitzat"
	body, _ := json.Marshal(quiz.DuplicateQuizInput{
		Title:          &customTitle,
		IncludeAnswers: true,
	})
	req2, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/quizzes/%s/duplicate", created.ID), bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for custom duplicate, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestHTTP_UploadImage(t *testing.T) {
	svc, _ := setupQuizService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestQuizRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	// Create multipart form with a valid PNG
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.png")
	if err != nil {
		t.Fatalf("error creating form file: %v", err)
	}
	// 8-byte PNG signature
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	_, _ = part.Write(pngHeader)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/uploads/images", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for image upload, got %d: %s", w.Code, w.Body.String())
	}

	var res shared.UploadResult
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode upload response: %v", err)
	}
	if res.URL == "" || res.Key == "" {
		t.Errorf("expected non-empty url and key, got url=%s, key=%s", res.URL, res.Key)
	}
}
