package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/encertia/backend/internal/auth"
	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, auth.Service) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := newMockRepository()
	cfg := auth.Config{
		JWTSecret:            "test-secret-key-32-characters-long!",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "encertia-test",
	}
	svc := auth.NewService(repo, cfg)
	handler := auth.NewHandler(svc)
	authMiddleware := shared.AuthMiddleware(handler.TokenValidatorAdapter())

	handler.RegisterRoutes(router.Group(""), authMiddleware)
	return router, svc
}

func TestHTTP_Register_And_Me(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Register
	body, _ := json.Marshal(auth.RegisterRequest{
		Email:     "anna.sole@encertia.cat",
		Password:  "StrongPassword123!",
		FirstName: "Anna",
		LastName:  "Sole",
		Role:      auth.RoleStudent,
	})

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var authRes auth.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &authRes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if authRes.User.Email != "anna.sole@encertia.cat" {
		t.Errorf("expected email anna.sole@encertia.cat, got %s", authRes.User.Email)
	}
	if authRes.Tokens.AccessToken == "" {
		t.Fatal("expected access token in response")
	}

	// 2. GET /auth/me with Bearer token
	reqMe, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+authRes.Tokens.AccessToken)
	wMe := httptest.NewRecorder()
	router.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for /auth/me, got %d: %s", wMe.Code, wMe.Body.String())
	}

	var userRes auth.UserResponse
	if err := json.Unmarshal(wMe.Body.Bytes(), &userRes); err != nil {
		t.Fatalf("failed to decode me response: %v", err)
	}

	if userRes.User.ID != authRes.User.ID {
		t.Errorf("expected user ID %s, got %s", authRes.User.ID, userRes.User.ID)
	}
}

func TestHTTP_Login_And_Refresh_And_Logout(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Pre-register user
	regBody, _ := json.Marshal(auth.RegisterRequest{
		Email:     "jordi.vidal@encertia.cat",
		Password:  "TeacherPass123!",
		FirstName: "Jordi",
		LastName:  "Vidal",
		Role:      auth.RoleTeacher,
	})
	reqReg, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(regBody))
	reqReg.Header.Set("Content-Type", "application/json")
	wReg := httptest.NewRecorder()
	router.ServeHTTP(wReg, reqReg)

	// 2. Login
	loginBody, _ := json.Marshal(auth.LoginRequest{
		Email:    "jordi.vidal@encertia.cat",
		Password: "TeacherPass123!",
	})
	reqLogin, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for login, got %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var loginRes auth.AuthResponse
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginRes)

	// 3. Refresh
	refreshBody, _ := json.Marshal(auth.RefreshTokenRequest{
		RefreshToken: loginRes.Tokens.RefreshToken,
	})
	reqRefresh, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(refreshBody))
	reqRefresh.Header.Set("Content-Type", "application/json")
	wRefresh := httptest.NewRecorder()
	router.ServeHTTP(wRefresh, reqRefresh)

	if wRefresh.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for refresh, got %d: %s", wRefresh.Code, wRefresh.Body.String())
	}

	var newTokens auth.TokenPair
	_ = json.Unmarshal(wRefresh.Body.Bytes(), &newTokens)

	// 4. Logout
	logoutBody, _ := json.Marshal(auth.LogoutRequest{
		RefreshToken: &newTokens.RefreshToken,
	})
	reqLogout, _ := http.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(logoutBody))
	reqLogout.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)
	reqLogout.Header.Set("Content-Type", "application/json")
	wLogout := httptest.NewRecorder()
	router.ServeHTTP(wLogout, reqLogout)

	if wLogout.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for logout, got %d: %s", wLogout.Code, wLogout.Body.String())
	}

	var msgRes auth.MessageResponse
	_ = json.Unmarshal(wLogout.Body.Bytes(), &msgRes)
	if msgRes.Message != "Sessió tancada correctament." {
		t.Errorf("expected logout message 'Sessió tancada correctament.', got '%s'", msgRes.Message)
	}
}

func TestHTTP_Unauthorized_Me(t *testing.T) {
	router, _ := setupTestRouter()

	reqMe, _ := http.NewRequest(http.MethodGet, "/auth/me", nil)
	wMe := httptest.NewRecorder()
	router.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d", wMe.Code)
	}

	var errRes shared.ErrorResponse
	_ = json.Unmarshal(wMe.Body.Bytes(), &errRes)
	if errRes.Error.Code != shared.ErrCodeUnauthorized {
		t.Errorf("expected code UNAUTHORIZED, got %s", errRes.Error.Code)
	}
}
