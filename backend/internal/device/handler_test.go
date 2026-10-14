package device_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestHTTP_DownloadAgent(t *testing.T) {
	adminID := uuid.New()
	router, _, _, _ := setupTestRouter("admin", adminID)

	// 1. Plataforma desconeguda -> 400 Bad Request
	req, _ := http.NewRequest(http.MethodGet, "/admin/devices/download/unknown-os", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for unknown platform, got %d: %s", w.Code, w.Body.String())
	}

	var errResp shared.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error.Code != shared.ErrCodeValidation {
		t.Errorf("expected error code %s, got %s", shared.ErrCodeValidation, errResp.Error.Code)
	}

	// 2. Descàrrega vàlida de binaris i certificat
	testDir := t.TempDir()
	t.Setenv("AGENT_DOWNLOADS_DIR", testDir)

	files := map[string]string{
		"encertia-agent-windows.exe": "dummy-windows-binary-content",
		"encertia-agent-darwin":      "dummy-darwin-binary-content",
		"encertia-agent-linux":       "dummy-linux-binary-content",
		"encertia-codesign.cer":      "dummy-cert-content",
	}

	for name, content := range files {
		err := os.WriteFile(filepath.Join(testDir, name), []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to write test file %s: %v", name, err)
		}
	}

	testCases := []struct {
		platform            string
		expectedFilename    string
		expectedContentType string
		expectedBody        string
	}{
		{
			platform:            "windows",
			expectedFilename:    "encertia-agent-windows.zip",
			expectedContentType: "application/zip",
		},
		{
			platform:            "windows-exe",
			expectedFilename:    "encertia-agent-windows.exe",
			expectedContentType: "application/octet-stream",
			expectedBody:        "dummy-windows-binary-content",
		},
		{
			platform:            "darwin",
			expectedFilename:    "encertia-agent-darwin",
			expectedContentType: "application/octet-stream",
			expectedBody:        "dummy-darwin-binary-content",
		},
		{
			platform:            "macos",
			expectedFilename:    "encertia-agent-darwin",
			expectedContentType: "application/octet-stream",
			expectedBody:        "dummy-darwin-binary-content",
		},
		{
			platform:            "linux",
			expectedFilename:    "encertia-agent-linux",
			expectedContentType: "application/octet-stream",
			expectedBody:        "dummy-linux-binary-content",
		},
		{
			platform:            "cert",
			expectedFilename:    "encertia-codesign.cer",
			expectedContentType: "application/x-x509-ca-cert",
			expectedBody:        "dummy-cert-content",
		},
	}

	for _, tc := range testCases {
		t.Run("download_"+tc.platform, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/admin/devices/download/"+tc.platform, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
			}

			expectedDisp := fmt.Sprintf(`attachment; filename="%s"`, tc.expectedFilename)
			if disp := w.Header().Get("Content-Disposition"); disp != expectedDisp {
				t.Errorf("expected Content-Disposition %q, got %q", expectedDisp, disp)
			}

			if ctype := w.Header().Get("Content-Type"); !strings.HasPrefix(ctype, tc.expectedContentType) {
				t.Errorf("expected Content-Type prefix %q, got %q", tc.expectedContentType, ctype)
			}

			if tc.expectedBody != "" && w.Body.String() != tc.expectedBody {
				t.Errorf("expected body %q, got %q", tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestHTTP_DownloadAgent_NotFound(t *testing.T) {
	adminID := uuid.New()
	router, _, _, _ := setupTestRouter("admin", adminID)

	emptyDir := t.TempDir()
	t.Setenv("AGENT_DOWNLOADS_DIR", emptyDir)

	// Canviem temporalment el directori de treball a emptyDir per assegurar que els camins relatius locals no trobin fitxers
	origWd, _ := os.Getwd()
	_ = os.Chdir(emptyDir)
	defer func() { _ = os.Chdir(origWd) }()

	req, _ := http.NewRequest(http.MethodGet, "/admin/devices/download/windows", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found, got %d: %s", w.Code, w.Body.String())
	}

	var errResp shared.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error.Code != "AGENT_FILE_NOT_FOUND" {
		t.Errorf("expected error code AGENT_FILE_NOT_FOUND, got %s", errResp.Error.Code)
	}
}

func TestHTTP_DownloadAgent_ForbiddenForNonAdmin(t *testing.T) {
	teacherID := uuid.New()
	router, _, _, _ := setupTestRouter("teacher", teacherID)

	req, _ := http.NewRequest(http.MethodGet, "/admin/devices/download/windows", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden for teacher, got %d: %s", w.Code, w.Body.String())
	}
}

