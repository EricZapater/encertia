package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
)

func setupTestRouter(svc Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Mock Auth Middleware for testing endpoints
	authMiddleware := func(c *gin.Context) {
		role := c.GetHeader("X-Test-Role")
		if role == "" {
			role = "admin"
		}
		c.Set(shared.CtxKeyUserID, "user-123")
		c.Set(shared.CtxKeyUserEmail, "admin@encertia.cat")
		c.Set(shared.CtxKeyUserRole, role)
		c.Next()
	}

	handler := NewHandler(svc)
	handler.RegisterRoutes(r, authMiddleware)
	return r
}

func TestMetricsService(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewService(mockRepo, nil)
	ctx := context.Background()

	userID := "user-uuid-1"
	userEmail := "admin@encertia.cat"
	userRole := "admin"
	ip := "127.0.0.1"

	t.Run("record audit log and get summary", func(t *testing.T) {
		entry := shared.AuditLogInput{
			UserID:     &userID,
			UserEmail:  &userEmail,
			UserRole:   &userRole,
			Action:     "GET /api/quizzes",
			Module:     "quizzes",
			Endpoint:   "/api/quizzes",
			Method:     "GET",
			StatusCode: 200,
			DurationMs: 50,
			IPAddress:  &ip,
		}

		err := svc.RecordAuditLog(ctx, entry)
		if err != nil {
			t.Fatalf("expected no error recording log, got %v", err)
		}

		summary, appErr := svc.GetSummary(ctx)
		if appErr != nil {
			t.Fatalf("expected no app error, got %v", appErr)
		}

		if summary.TotalRequestsToday != 1 {
			t.Errorf("expected 1 total request today, got %d", summary.TotalRequestsToday)
		}
		if summary.ActiveUsersToday != 1 {
			t.Errorf("expected 1 active user today, got %d", summary.ActiveUsersToday)
		}
	})

	t.Run("get api latency metrics", func(t *testing.T) {
		latencyResp, appErr := svc.GetApiLatency(ctx)
		if appErr != nil {
			t.Fatalf("expected no app error, got %v", appErr)
		}

		if len(latencyResp.Endpoints) != 1 {
			t.Fatalf("expected 1 endpoint metric, got %d", len(latencyResp.Endpoints))
		}

		item := latencyResp.Endpoints[0]
		if item.Endpoint != "/api/quizzes" || item.Method != "GET" {
			t.Errorf("unexpected endpoint item: %+v", item)
		}
	})

	t.Run("get audit logs paginated", func(t *testing.T) {
		resp, appErr := svc.GetAuditLogs(ctx, AuditLogFilter{Page: 1, PageSize: 10, Module: "quizzes"})
		if appErr != nil {
			t.Fatalf("expected no app error, got %v", appErr)
		}

		if resp.Total != 1 || len(resp.Items) != 1 {
			t.Errorf("expected 1 audit log item, got total=%d, items=%d", resp.Total, len(resp.Items))
		}
	})

	t.Run("export audit logs CSV", func(t *testing.T) {
		csvBytes, appErr := svc.ExportAuditLogsCSV(ctx)
		if appErr != nil {
			t.Fatalf("expected no app error, got %v", appErr)
		}

		if len(csvBytes) == 0 {
			t.Errorf("expected non-empty CSV output")
		}
	})
}

func TestMetricsEndpoints(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewService(mockRepo, nil)
	router := setupTestRouter(svc)

	// Seed some audit log
	uid := "admin-1"
	email := "admin@encertia.cat"
	role := "admin"
	_ = svc.RecordAuditLog(context.Background(), shared.AuditLogInput{
		UserID:     &uid,
		UserEmail:  &email,
		UserRole:   &role,
		Action:     "GET /metrics/summary",
		Module:     "metrics",
		Endpoint:   "/metrics/summary",
		Method:     "GET",
		StatusCode: 200,
		DurationMs: 15,
	})

	t.Run("GET /metrics/summary success for admin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metrics/summary", nil)
		req.Header.Set("X-Test-Role", "admin")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var summary MetricsSummary
		if err := json.Unmarshal(w.Body.Bytes(), &summary); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if summary.TotalRequestsToday < 1 {
			t.Errorf("expected at least 1 total request today, got %d", summary.TotalRequestsToday)
		}
	})

	t.Run("GET /metrics/summary forbidden for teacher", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metrics/summary", nil)
		req.Header.Set("X-Test-Role", "teacher")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403 Forbidden, got %d", w.Code)
		}
	})

	t.Run("GET /metrics/api-latency", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metrics/api-latency", nil)
		req.Header.Set("X-Test-Role", "admin")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /metrics/audit-logs", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metrics/audit-logs?page=1&pageSize=10", nil)
		req.Header.Set("X-Test-Role", "admin")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("GET /metrics/audit-logs/export", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metrics/audit-logs/export", nil)
		req.Header.Set("X-Test-Role", "admin")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); contentType != "text/csv; charset=utf-8" {
			t.Errorf("expected content type 'text/csv; charset=utf-8', got '%s'", contentType)
		}
	})
}
