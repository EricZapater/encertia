package match_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/encertia/backend/internal/match"
	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockTokenValidator struct {
	userID string
	email  string
	role   string
	err    *shared.AppError
}

func (m *mockTokenValidator) ValidateAccessToken(token string) (string, string, string, *shared.AppError) {
	if m.err != nil {
		return "", "", "", m.err
	}
	return m.userID, m.email, m.role, nil
}

func setupTestMatchRouter(svc match.Service, hub *match.Hub, validator shared.TokenValidator, actorID uuid.UUID, actorRole, actorEmail string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := match.NewHandler(svc, validator, hub)

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

func TestHTTP_CreateMatch(t *testing.T) {
	_, hub, svc, hostID, qd := setupTestService()
	validator := &mockTokenValidator{userID: hostID.String(), role: "teacher", email: "host@encertia.cat"}
	router := setupTestMatchRouter(svc, hub, validator, hostID, "teacher", "host@encertia.cat")

	body, _ := json.Marshal(match.CreateMatchRequest{
		QuizID: qd.ID,
	})

	req, _ := http.NewRequest(http.MethodPost, "/matches", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var res match.MatchCreatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.QuizID != qd.ID || len(res.PIN) != 6 {
		t.Errorf("resposta inesperada: %+v", res)
	}
}

func TestHTTP_GetMatchByPin(t *testing.T) {
	ctx := context.Background()
	_, hub, svc, hostID, qd := setupTestService()
	validator := &mockTokenValidator{userID: hostID.String(), role: "teacher", email: "host@encertia.cat"}
	router := setupTestMatchRouter(svc, hub, validator, hostID, "teacher", "host@encertia.cat")

	created, _ := svc.CreateMatch(ctx, hostID, qd.ID)

	// Valid PIN
	req, _ := http.NewRequest(http.MethodGet, "/matches/"+created.PIN, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	// Invalid PIN length
	reqBad, _ := http.NewRequest(http.MethodGet, "/matches/123", nil)
	wBad := httptest.NewRecorder()
	router.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", wBad.Code)
	}
}

func TestHTTP_JoinMatch(t *testing.T) {
	ctx := context.Background()
	_, hub, svc, hostID, qd := setupTestService()
	playerUserID := uuid.New()
	validator := &mockTokenValidator{userID: playerUserID.String(), role: "student", email: "student@encertia.cat"}
	router := setupTestMatchRouter(svc, hub, validator, playerUserID, "student", "student@encertia.cat")

	created, _ := svc.CreateMatch(ctx, hostID, qd.ID)

	body, _ := json.Marshal(match.JoinMatchRequest{
		Nickname: "Pol",
	})

	req, _ := http.NewRequest(http.MethodPost, "/matches/"+created.PIN+"/join", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res match.JoinMatchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Nickname != "Pol" || res.PIN != created.PIN {
		t.Errorf("resposta inesperada: %+v", res)
	}
}

func TestHTTP_GetMatchSummary(t *testing.T) {
	ctx := context.Background()
	_, hub, svc, hostID, qd := setupTestService()
	validator := &mockTokenValidator{userID: hostID.String(), role: "teacher", email: "host@encertia.cat"}
	router := setupTestMatchRouter(svc, hub, validator, hostID, "teacher", "host@encertia.cat")

	created, _ := svc.CreateMatch(ctx, hostID, qd.ID)

	req, _ := http.NewRequest(http.MethodGet, "/matches/"+created.ID.String()+"/summary", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res match.MatchSummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.MatchID != created.ID || res.TotalQuestions != len(qd.Questions) {
		t.Errorf("resposta de resum inesperada: %+v", res)
	}
}

func TestWebSocket_EndpointValidation(t *testing.T) {
	ctx := context.Background()
	_, hub, svc, hostID, qd := setupTestService()
	validator := &mockTokenValidator{userID: hostID.String(), role: "teacher", email: "host@encertia.cat"}
	router := setupTestMatchRouter(svc, hub, validator, hostID, "teacher", "host@encertia.cat")

	created, _ := svc.CreateMatch(ctx, hostID, qd.ID)

	// 1. Invalid PIN length
	req1, _ := http.NewRequest(http.MethodGet, "/ws/match/123", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w1.Code)
	}

	// 2. Missing token
	req2, _ := http.NewRequest(http.MethodGet, "/ws/match/"+created.PIN, nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized on missing token, got %d", w2.Code)
	}

	// 3. Invalid token
	badValidator := &mockTokenValidator{err: shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "invalid")}
	routerBadVal := setupTestMatchRouter(svc, hub, badValidator, hostID, "teacher", "host@encertia.cat")
	req3, _ := http.NewRequest(http.MethodGet, "/ws/match/"+created.PIN+"?token=bad-token", nil)
	w3 := httptest.NewRecorder()
	routerBadVal.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized on bad token, got %d", w3.Code)
	}

	// 4. Match Not Found
	req4, _ := http.NewRequest(http.MethodGet, "/ws/match/999999?token=valid-token", nil)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found on non-existent match, got %d", w4.Code)
	}

	// 5. Unregistered player Forbidden
	otherUserID := uuid.New()
	studentValidator := &mockTokenValidator{userID: otherUserID.String(), role: "student", email: "student@encertia.cat"}
	routerStudent := setupTestMatchRouter(svc, hub, studentValidator, otherUserID, "student", "student@encertia.cat")
	req5, _ := http.NewRequest(http.MethodGet, "/ws/match/"+created.PIN+"?token=valid-token", nil)
	w5 := httptest.NewRecorder()
	routerStudent.ServeHTTP(w5, req5)
	if w5.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for unregistered player, got %d", w5.Code)
	}
}
