package course

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	rg := r.Group("")

	dummyAuthMiddleware := func(c *gin.Context) {
		userID := c.GetHeader("X-Test-User-ID")
		role := c.GetHeader("X-Test-User-Role")
		if userID != "" {
			c.Set(shared.CtxKeyUserID, userID)
		}
		if role != "" {
			c.Set(shared.CtxKeyUserRole, role)
		}
		c.Next()
	}

	handler.RegisterRoutes(rg, dummyAuthMiddleware)
	return r
}

func TestHandler_CoursesEndpoints(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewService(mockRepo)
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	teacherID := uuid.New().String()
	studentID := uuid.New().String()

	var createdCourseID string

	t.Run("POST /courses - Teacher creates course", func(t *testing.T) {
		body := CreateCourseRequest{
			Title: "Curs de Programació Go",
			Code:  "GOLANG101",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/courses", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", teacherID)
		req.Header.Set("X-Test-User-Role", "teacher")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201 Created, got %d. Body: %s", w.Code, w.Body.String())
		}

		var res CourseResponse
		json.Unmarshal(w.Body.Bytes(), &res)
		if res.Title != "Curs de Programació Go" {
			t.Errorf("unexpected course title: %s", res.Title)
		}
		createdCourseID = res.ID.String()
	})

	t.Run("POST /courses - Student gets forbidden by RequireRole", func(t *testing.T) {
		body := CreateCourseRequest{
			Title: "Curs de Python",
			Code:  "PY101",
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/courses", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", studentID)
		req.Header.Set("X-Test-User-Role", "student")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403 Forbidden, got %d", w.Code)
		}
	})

	t.Run("GET /courses - Teacher lists courses", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/courses", nil)
		req.Header.Set("X-Test-User-ID", teacherID)
		req.Header.Set("X-Test-User-Role", "teacher")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", w.Code)
		}

		var res CourseListResponse
		json.Unmarshal(w.Body.Bytes(), &res)
		if res.Total != 1 {
			t.Errorf("expected 1 course, got %d", res.Total)
		}
	})

	t.Run("POST /courses/:id/students - Enroll student", func(t *testing.T) {
		body := EnrollStudentsRequest{
			StudentIDs: []uuid.UUID{uuid.MustParse(studentID)},
		}
		jsonBody, _ := json.Marshal(body)

		url := "/courses/" + createdCourseID + "/students"
		req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User-ID", teacherID)
		req.Header.Set("X-Test-User-Role", "teacher")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("DELETE /courses/:id - Delete course", func(t *testing.T) {
		url := "/courses/" + createdCourseID
		req, _ := http.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("X-Test-User-ID", teacherID)
		req.Header.Set("X-Test-User-Role", "teacher")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204 No Content, got %d", w.Code)
		}
	})
}
