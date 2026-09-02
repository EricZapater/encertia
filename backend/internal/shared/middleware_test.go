package shared_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validator := shared.TokenValidatorFunc(func(ctx context.Context, tokenString string) (string, string, string, *shared.AppError) {
		if tokenString == "valid-token" {
			return "user-123", "user@encertia.cat", "teacher", nil
		}
		if tokenString == "expired-token" {
			return "", "", "", shared.ErrUnauthorized(shared.ErrCodeTokenExpired, "Token expirat.")
		}
		return "", "", "", shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Token invàlid.")
	})

	router := gin.New()
	router.Use(shared.AuthMiddleware(validator))
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString(shared.CtxKeyUserID)
		email := c.GetString(shared.CtxKeyUserEmail)
		role := c.GetString(shared.CtxKeyUserRole)
		c.JSON(http.StatusOK, gin.H{
			"userId": userID,
			"email":  email,
			"role":   role,
		})
	})

	// Test 1: No Authorization Header
	req1, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing header, got %d", w1.Code)
	}

	// Test 2: Invalid Header Format
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid format, got %d", w2.Code)
	}

	// Test 3: Invalid Token
	req3, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer bad-token")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad token, got %d", w3.Code)
	}

	// Test 4: Valid Token
	req4, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req4.Header.Set("Authorization", "Bearer valid-token")
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Errorf("expected 200 for valid token, got %d: %s", w4.Code, w4.Body.String())
	}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(shared.CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// OPTIONS preflight
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for OPTIONS, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
