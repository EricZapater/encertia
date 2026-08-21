package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/encertia/backend/internal/shared"
	"github.com/encertia/backend/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestRouter(svc user.Service, actorID uuid.UUID, actorRole string, actorEmail string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := user.NewHandler(svc)

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

func TestHTTP_ListUsers(t *testing.T) {
	svc, _ := setupUserService()
	adminID := uuid.New()
	router := setupTestRouter(svc, adminID, "admin", "admin@encertia.cat")

	ctx := t.Context()
	_, _ = svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "u1@encertia.cat",
		Password:  "Password123!",
		FirstName: "Marc",
		LastName:  "Vila",
		Role:      user.RoleStudent,
	})

	req, _ := http.NewRequest(http.MethodGet, "/users?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res user.UserListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Pagination.TotalCount != 1 {
		t.Errorf("expected 1 user, got %d", res.Pagination.TotalCount)
	}
}

func TestHTTP_CreateUser(t *testing.T) {
	svc, _ := setupUserService()
	adminID := uuid.New()
	router := setupTestRouter(svc, adminID, "admin", "admin@encertia.cat")

	body, _ := json.Marshal(user.CreateUserInput{
		Email:     "nou.usuari@encertia.cat",
		Password:  "Password123!",
		FirstName: "Laia",
		LastName:  "Sole",
		Role:      user.RoleTeacher,
	})

	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var res user.UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if res.User.Email != "nou.usuari@encertia.cat" {
		t.Errorf("expected nou.usuari@encertia.cat, got %s", res.User.Email)
	}
}

func TestHTTP_BatchCreateUsers(t *testing.T) {
	svc, _ := setupUserService()
	teacherID := uuid.New()
	router := setupTestRouter(svc, teacherID, "teacher", "teacher@encertia.cat")

	body, _ := json.Marshal(user.BatchCreateUsersRequest{
		Users: []user.BatchUserItem{
			{Email: "alumne1@encertia.cat", FirstName: "Alumne", LastName: "Un", Role: user.RoleStudent},
			{Email: "alumne2@encertia.cat", FirstName: "Alumne", LastName: "Dos", Role: user.RoleStudent},
		},
	})

	req, _ := http.NewRequest(http.MethodPost, "/users/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for batch, got %d: %s", w.Code, w.Body.String())
	}

	var res user.BatchCreateUsersResponse
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.CreatedCount != 2 {
		t.Errorf("expected 2 created users, got %d", res.CreatedCount)
	}
}

func TestHTTP_GetUserByID(t *testing.T) {
	svc, _ := setupUserService()
	adminID := uuid.New()
	router := setupTestRouter(svc, adminID, "admin", "admin@encertia.cat")

	ctx := t.Context()
	created, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "target@encertia.cat",
		Password:  "Password123!",
		FirstName: "Target",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	// 1. Success 200
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", created.User.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	// 2. Invalid UUID 400
	reqInvalid, _ := http.NewRequest(http.MethodGet, "/users/invalid-uuid", nil)
	wInvalid := httptest.NewRecorder()
	router.ServeHTTP(wInvalid, reqInvalid)

	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", wInvalid.Code)
	}

	// 3. Not found 404
	reqNotFound, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", uuid.New()), nil)
	wNotFound := httptest.NewRecorder()
	router.ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent user, got %d", wNotFound.Code)
	}
}

func TestHTTP_UpdateUser(t *testing.T) {
	svc, _ := setupUserService()
	ctx := t.Context()
	created, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "self@encertia.cat",
		Password:  "Password123!",
		FirstName: "Self",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	// Setup router with student actor matching created.User.ID
	routerSelf := setupTestRouter(svc, created.User.ID, "student", "self@encertia.cat")

	newName := "UpdatedSelf"
	body, _ := json.Marshal(user.UpdateUserInput{
		FirstName: &newName,
	})

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%s", created.User.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	routerSelf.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for update, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTP_ResetPassword(t *testing.T) {
	svc, _ := setupUserService()
	adminID := uuid.New()
	router := setupTestRouter(svc, adminID, "admin", "admin@encertia.cat")

	ctx := t.Context()
	created, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "resetme@encertia.cat",
		Password:  "Password123!",
		FirstName: "Reset",
		LastName:  "Me",
		Role:      user.RoleStudent,
	})

	body, _ := json.Marshal(user.ResetPasswordInput{
		NewPassword: "BrandNewSecurePassword123!",
	})

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/users/%s/password", created.User.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for reset password, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTP_DeleteUser(t *testing.T) {
	svc, _ := setupUserService()
	adminID := uuid.New()
	router := setupTestRouter(svc, adminID, "admin", "admin@encertia.cat")

	ctx := t.Context()
	created, _ := svc.CreateUser(ctx, "admin", user.CreateUserInput{
		Email:     "deleteme@encertia.cat",
		Password:  "Password123!",
		FirstName: "Delete",
		LastName:  "Me",
		Role:      user.RoleStudent,
	})

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s", created.User.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for delete, got %d: %s", w.Code, w.Body.String())
	}
}
