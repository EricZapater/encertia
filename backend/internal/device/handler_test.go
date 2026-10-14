package device_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/encertia/backend/internal/device"
	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestRouter(actorRole string, actorID uuid.UUID) (*gin.Engine, device.Service, *mockRepository, *device.TeacherInfo) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := newMockRepository()

	teacherID := uuid.New()
	teacher := &device.TeacherInfo{
		ID:       teacherID,
		Email:    "joan.docent@encertia.cat",
		Name:     "Joan Docent",
		Role:     "teacher",
		IsActive: true,
	}
	repo.teachersByID[teacherID] = teacher
	repo.teachersByEmail[teacher.Email] = teacher

	svc := device.NewService(repo)
	handler := device.NewHandler(svc)

	fakeAuthMiddleware := func(c *gin.Context) {
		if actorRole != "" {
			c.Set(shared.CtxKeyUserID, actorID.String())
			c.Set(shared.CtxKeyUserRole, actorRole)
		}
		c.Next()
	}

	handler.RegisterRoutes(router.Group(""), fakeAuthMiddleware)
	return router, svc, repo, teacher
}

func TestHTTP_VerifyDevice(t *testing.T) {
	adminID := uuid.New()
	router, _, repo, teacher := setupTestRouter("admin", adminID)

	hash := strings.Repeat("a", 64)
	devID := uuid.New()
	repo.devices[hash] = &device.AuthorizedDevice{
		ID:               devID,
		HardwareUUIDHash: hash,
		Name:             "Portàtil Aula 1",
		TeacherID:        teacher.ID,
		IsActive:         true,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repo.devicesByID[devID] = repo.devices[hash]

	// 1. Success
	body, _ := json.Marshal(device.VerifyDeviceRequest{HardwareUUIDHash: hash})
	req, _ := http.NewRequest(http.MethodPost, "/devices/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var res device.VerifyDeviceResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !strings.HasPrefix(res.BootstrapToken, "dbt_") {
		t.Errorf("expected dbt_ prefix, got %s", res.BootstrapToken)
	}
	if res.TeacherEmail != teacher.Email {
		t.Errorf("expected email %s, got %s", teacher.Email, res.TeacherEmail)
	}

	// 2. Unregistered device -> 403 Forbidden
	unregisteredHash := strings.Repeat("b", 64)
	body, _ = json.Marshal(device.VerifyDeviceRequest{HardwareUUIDHash: unregisteredHash})
	req, _ = http.NewRequest(http.MethodPost, "/devices/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTP_AdminDevices_CRUD(t *testing.T) {
	adminID := uuid.New()
	router, _, _, teacher := setupTestRouter("admin", adminID)

	hash := strings.Repeat("c", 64)

	// 1. Register device
	regBody, _ := json.Marshal(device.RegisterDeviceRequest{
		HardwareUUIDHash: hash,
		Name:             "Equip Administració",
		TeacherEmail:     teacher.Email,
	})
	req, _ := http.NewRequest(http.MethodPost, "/admin/devices", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var summary device.DeviceSummary
	if err := json.Unmarshal(w.Body.Bytes(), &summary); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if summary.Name != "Equip Administració" {
		t.Errorf("expected name Equip Administració, got %s", summary.Name)
	}

	// 2. List devices
	req, _ = http.NewRequest(http.MethodGet, "/admin/devices", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var list []device.DeviceSummary
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 device in list, got %d", len(list))
	}

	// 3. Revoke device
	req, _ = http.NewRequest(http.MethodDelete, "/admin/devices/"+summary.ID.String(), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d: %s", w.Code, w.Body.String())
	}
}
