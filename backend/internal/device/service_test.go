package device_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/encertia/backend/internal/device"
	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

func setupTestService() (device.Service, *mockRepository, *device.TeacherInfo) {
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
	return svc, repo, teacher
}

func TestVerifyDevice_Success(t *testing.T) {
	svc, repo, teacher := setupTestService()
	ctx := context.Background()

	hash := strings.Repeat("a", 64)
	devID := uuid.New()
	dev := &device.AuthorizedDevice{
		ID:               devID,
		HardwareUUIDHash: hash,
		Name:             "Portàtil Aula 3B",
		TeacherID:        teacher.ID,
		IsActive:         true,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repo.devices[hash] = dev
	repo.devicesByID[devID] = dev

	res, appErr := svc.VerifyDevice(ctx, device.VerifyDeviceRequest{
		HardwareUUIDHash: hash,
	})
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}

	if !strings.HasPrefix(res.BootstrapToken, "dbt_") {
		t.Errorf("expected bootstrap token prefix dbt_, got %s", res.BootstrapToken)
	}
	if res.TeacherEmail != teacher.Email {
		t.Errorf("expected teacher email %s, got %s", teacher.Email, res.TeacherEmail)
	}
	if res.DeviceName != "Portàtil Aula 3B" {
		t.Errorf("expected device name Portàtil Aula 3B, got %s", res.DeviceName)
	}
	if res.ExpiresInSeconds != 60 {
		t.Errorf("expected expiresInSeconds 60, got %d", res.ExpiresInSeconds)
	}
}

func TestVerifyDevice_AdminRoleAllowed(t *testing.T) {
	svc, repo, _ := setupTestService()
	ctx := context.Background()

	adminID := uuid.New()
	admin := &device.TeacherInfo{
		ID:       adminID,
		Email:    "admin@encertia.cat",
		Name:     "Admin Encertia",
		Role:     "admin",
		IsActive: true,
	}
	repo.teachersByID[adminID] = admin
	repo.teachersByEmail[admin.Email] = admin

	hash := strings.Repeat("e", 64)
	devID := uuid.New()
	dev := &device.AuthorizedDevice{
		ID:               devID,
		HardwareUUIDHash: hash,
		Name:             "Ordinador de Direcció",
		TeacherID:        admin.ID,
		IsActive:         true,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repo.devices[hash] = dev
	repo.devicesByID[devID] = dev

	res, appErr := svc.VerifyDevice(ctx, device.VerifyDeviceRequest{
		HardwareUUIDHash: hash,
	})
	if appErr != nil {
		t.Fatalf("unexpected error for admin device verification: %v", appErr)
	}

	if res.TeacherEmail != admin.Email {
		t.Errorf("expected email %s, got %s", admin.Email, res.TeacherEmail)
	}
}

func TestVerifyDevice_NotFound(t *testing.T) {
	svc, _, _ := setupTestService()
	ctx := context.Background()

	res, appErr := svc.VerifyDevice(ctx, device.VerifyDeviceRequest{
		HardwareUUIDHash: strings.Repeat("b", 64),
	})
	if res != nil {
		t.Fatalf("expected nil response, got %+v", res)
	}
	if appErr == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", appErr.StatusCode)
	}
	if appErr.Code != shared.ErrCodeDeviceNotAuthorized {
		t.Errorf("expected code DEVICE_NOT_AUTHORIZED, got %s", appErr.Code)
	}
}

func TestVerifyDevice_InactiveDevice(t *testing.T) {
	svc, repo, teacher := setupTestService()
	ctx := context.Background()

	hash := strings.Repeat("c", 64)
	devID := uuid.New()
	dev := &device.AuthorizedDevice{
		ID:               devID,
		HardwareUUIDHash: hash,
		Name:             "Portàtil Desactivat",
		TeacherID:        teacher.ID,
		IsActive:         false,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repo.devices[hash] = dev
	repo.devicesByID[devID] = dev

	res, appErr := svc.VerifyDevice(ctx, device.VerifyDeviceRequest{
		HardwareUUIDHash: hash,
	})
	if res != nil {
		t.Fatalf("expected nil response, got %+v", res)
	}
	if appErr == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", appErr.StatusCode)
	}
	if appErr.Code != shared.ErrCodeDeviceNotAuthorized {
		t.Errorf("expected code DEVICE_NOT_AUTHORIZED, got %s", appErr.Code)
	}
}

func TestVerifyDevice_InvalidHash(t *testing.T) {
	svc, _, _ := setupTestService()
	ctx := context.Background()

	testCases := []string{
		"",
		"short",
		strings.Repeat("z", 64), // not hex
	}

	for _, tc := range testCases {
		_, appErr := svc.VerifyDevice(ctx, device.VerifyDeviceRequest{HardwareUUIDHash: tc})
		if appErr == nil {
			t.Errorf("expected error for hash '%s', got nil", tc)
		} else if appErr.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400 for hash '%s', got %d", tc, appErr.StatusCode)
		}
	}
}

func TestRegisterDevice_Success(t *testing.T) {
	svc, _, teacher := setupTestService()
	ctx := context.Background()
	adminID := uuid.New()

	hash := strings.Repeat("d", 64)
	res, appErr := svc.RegisterDevice(ctx, adminID, "admin", device.RegisterDeviceRequest{
		HardwareUUIDHash: hash,
		Name:             "Ordinador Aula Informàtica",
		TeacherEmail:     teacher.Email,
	})
	if appErr != nil {
		t.Fatalf("unexpected error registering device: %v", appErr)
	}

	if res.Name != "Ordinador Aula Informàtica" {
		t.Errorf("expected name 'Ordinador Aula Informàtica', got '%s'", res.Name)
	}
	if res.TeacherEmail != teacher.Email {
		t.Errorf("expected teacher email '%s', got '%s'", teacher.Email, res.TeacherEmail)
	}
	if !res.IsActive {
		t.Error("expected isActive to be true")
	}
}

func TestRegisterDevice_ForbiddenForNonAdmin(t *testing.T) {
	svc, _, teacher := setupTestService()
	ctx := context.Background()

	_, appErr := svc.RegisterDevice(ctx, teacher.ID, "teacher", device.RegisterDeviceRequest{
		HardwareUUIDHash: strings.Repeat("e", 64),
		Name:             "Ordinador Aula",
		TeacherEmail:     teacher.Email,
	})
	if appErr == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", appErr.StatusCode)
	}
}

func TestRegisterDevice_TeacherNotFound(t *testing.T) {
	svc, _, _ := setupTestService()
	ctx := context.Background()
	adminID := uuid.New()

	_, appErr := svc.RegisterDevice(ctx, adminID, "admin", device.RegisterDeviceRequest{
		HardwareUUIDHash: strings.Repeat("f", 64),
		Name:             "Ordinador Aula",
		TeacherEmail:     "inexistent@encertia.cat",
	})
	if appErr == nil {
		t.Fatal("expected error for non-existent teacher, got nil")
	}
	if appErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", appErr.StatusCode)
	}
}

func TestRegisterDevice_AdminRoleAllowed(t *testing.T) {
	svc, repo, _ := setupTestService()
	ctx := context.Background()
	adminID := uuid.New()
	adminUser := &device.TeacherInfo{
		ID:       adminID,
		Email:    "admin@encertia.cat",
		Name:     "Admin Encertia",
		Role:     "admin",
		IsActive: true,
	}
	repo.teachersByID[adminID] = adminUser
	repo.teachersByEmail[adminUser.Email] = adminUser

	res, appErr := svc.RegisterDevice(ctx, adminID, "admin", device.RegisterDeviceRequest{
		HardwareUUIDHash: strings.Repeat("a", 64),
		Name:             "Ordinador de l'Admin",
		TeacherEmail:     adminUser.Email,
	})
	if appErr != nil {
		t.Fatalf("expected success for admin assignment, got error: %v", appErr)
	}
	if res.TeacherEmail != adminUser.Email {
		t.Errorf("expected email %s, got %s", adminUser.Email, res.TeacherEmail)
	}
}

func TestRegisterDevice_StudentRoleRejected(t *testing.T) {
	svc, repo, _ := setupTestService()
	ctx := context.Background()
	adminID := uuid.New()
	studentID := uuid.New()
	studentUser := &device.TeacherInfo{
		ID:       studentID,
		Email:    "student@encertia.cat",
		Name:     "Alumne Prova",
		Role:     "student",
		IsActive: true,
	}
	repo.teachersByID[studentID] = studentUser
	repo.teachersByEmail[studentUser.Email] = studentUser

	_, appErr := svc.RegisterDevice(ctx, adminID, "admin", device.RegisterDeviceRequest{
		HardwareUUIDHash: strings.Repeat("b", 64),
		Name:             "Ordinador Alumne",
		TeacherEmail:     studentUser.Email,
	})
	if appErr == nil {
		t.Fatal("expected error when assigning device to student, got nil")
	}
	if appErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", appErr.StatusCode)
	}
}

func TestListDevices_Success(t *testing.T) {
	svc, repo, teacher := setupTestService()
	ctx := context.Background()

	hash := strings.Repeat("1", 64)
	devID := uuid.New()
	dev := &device.AuthorizedDevice{
		ID:               devID,
		HardwareUUIDHash: hash,
		Name:             "Equip 1",
		TeacherID:        teacher.ID,
		IsActive:         true,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repo.devices[hash] = dev
	repo.devicesByID[devID] = dev

	devices, appErr := svc.ListDevices(ctx, "admin")
	if appErr != nil {
		t.Fatalf("unexpected error listing devices: %v", appErr)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Name != "Equip 1" {
		t.Errorf("expected name Equip 1, got %s", devices[0].Name)
	}
}

func TestListDevices_Forbidden(t *testing.T) {
	svc, _, _ := setupTestService()
	ctx := context.Background()

	_, appErr := svc.ListDevices(ctx, "teacher")
	if appErr == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", appErr.StatusCode)
	}
}

func TestRevokeDevice_Success(t *testing.T) {
	svc, repo, teacher := setupTestService()
	ctx := context.Background()

	hash := strings.Repeat("2", 64)
	devID := uuid.New()
	dev := &device.AuthorizedDevice{
		ID:               devID,
		HardwareUUIDHash: hash,
		Name:             "Equip a Revocar",
		TeacherID:        teacher.ID,
		IsActive:         true,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
	repo.devices[hash] = dev
	repo.devicesByID[devID] = dev

	appErr := svc.RevokeDevice(ctx, "admin", devID)
	if appErr != nil {
		t.Fatalf("unexpected error revoking device: %v", appErr)
	}

	if repo.devicesByID[devID].IsActive {
		t.Error("expected device to be deactivated")
	}
}

func TestRevokeDevice_NotFound(t *testing.T) {
	svc, _, _ := setupTestService()
	ctx := context.Background()

	appErr := svc.RevokeDevice(ctx, "admin", uuid.New())
	if appErr == nil {
		t.Fatal("expected error, got nil")
	}
	if appErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", appErr.StatusCode)
	}
}

func TestValidateAndConsumeBootstrapToken(t *testing.T) {
	svc, repo, teacher := setupTestService()
	ctx := context.Background()

	devID := uuid.New()
	dev := &device.AuthorizedDevice{
		ID:               devID,
		HardwareUUIDHash: strings.Repeat("3", 64),
		Name:             "Equip Test Token",
		TeacherID:        teacher.ID,
		IsActive:         true,
	}
	repo.devicesByID[devID] = dev

	token := "dbt_valid12345"
	repo.bootstrapTokens[token] = &device.DeviceBootstrapToken{
		Token:     token,
		DeviceID:  devID,
		TeacherID: teacher.ID,
		ExpiresAt: time.Now().UTC().Add(60 * time.Second),
	}

	// 1. Wrong teacher ID fails
	wrongTeacherID := uuid.New()
	if err := svc.ValidateAndConsumeBootstrapToken(ctx, token, wrongTeacherID); err == nil {
		t.Error("expected error for wrong teacher ID, got nil")
	}

	// 2. Correct teacher ID succeeds
	if err := svc.ValidateAndConsumeBootstrapToken(ctx, token, teacher.ID); err != nil {
		t.Fatalf("unexpected error consuming token: %v", err)
	}

	// 3. Re-consuming already used token fails
	if err := svc.ValidateAndConsumeBootstrapToken(ctx, token, teacher.ID); err == nil {
		t.Error("expected error for already used token, got nil")
	}

	// 4. Expired token fails
	expiredToken := "dbt_expired12345"
	repo.bootstrapTokens[expiredToken] = &device.DeviceBootstrapToken{
		Token:     expiredToken,
		DeviceID:  devID,
		TeacherID: teacher.ID,
		ExpiresAt: time.Now().UTC().Add(-10 * time.Second),
	}
	if err := svc.ValidateAndConsumeBootstrapToken(ctx, expiredToken, teacher.ID); err == nil {
		t.Error("expected error for expired token, got nil")
	}
}
