package device

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

// Service defines business logic operations for authorized devices.
type Service interface {
	VerifyDevice(ctx context.Context, req VerifyDeviceRequest) (*VerifyDeviceResponse, *shared.AppError)
	ListDevices(ctx context.Context, actorRole string) ([]DeviceSummary, *shared.AppError)
	RegisterDevice(ctx context.Context, actorID uuid.UUID, actorRole string, req RegisterDeviceRequest) (*DeviceSummary, *shared.AppError)
	RevokeDevice(ctx context.Context, actorRole string, deviceID uuid.UUID) *shared.AppError
	ValidateAndConsumeBootstrapToken(ctx context.Context, token string, teacherID uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService creates a new instance of the device service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) VerifyDevice(ctx context.Context, req VerifyDeviceRequest) (*VerifyDeviceResponse, *shared.AppError) {
	req.HardwareUUIDHash = strings.TrimSpace(strings.ToLower(req.HardwareUUIDHash))
	if req.HardwareUUIDHash == "" || len(req.HardwareUUIDHash) != 64 || !isHexString(req.HardwareUUIDHash) {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El hash del UUID de maquinari ha de ser un SHA-256 hexadecimal de 64 caràcters.",
			map[string]interface{}{"field": "hardwareUuidHash"},
		)
	}

	dev, err := s.repo.GetDeviceByHardwareHash(ctx, req.HardwareUUIDHash)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return nil, shared.NewAppError(
				http.StatusForbidden,
				shared.ErrCodeDeviceNotAuthorized,
				"Aquest equip no està autoritzat per a la docència.",
				nil,
				nil,
			)
		}
		return nil, shared.ErrInternal(err)
	}

	if !dev.IsActive {
		return nil, shared.NewAppError(
			http.StatusForbidden,
			shared.ErrCodeDeviceNotAuthorized,
			"Aquest equip no està autoritzat per a la docència.",
			nil,
			nil,
		)
	}

	teacher, err := s.repo.GetTeacherByID(ctx, dev.TeacherID)
	if err != nil {
		if errors.Is(err, ErrTeacherNotFound) {
			return nil, shared.NewAppError(
				http.StatusForbidden,
				shared.ErrCodeDeviceNotAuthorized,
				"Aquest equip no està autoritzat per a la docència.",
				nil,
				nil,
			)
		}
		return nil, shared.ErrInternal(err)
	}

	if !teacher.IsActive || teacher.Role != "teacher" {
		return nil, shared.NewAppError(
			http.StatusForbidden,
			shared.ErrCodeDeviceNotAuthorized,
			"El docent assignat a aquest equip no està autoritzat o el compte està desactivat.",
			nil,
			nil,
		)
	}

	tokenStr, err := generateBootstrapToken()
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	token := &DeviceBootstrapToken{
		Token:     tokenStr,
		DeviceID:  dev.ID,
		TeacherID: dev.TeacherID,
		ExpiresAt: time.Now().UTC().Add(60 * time.Second),
	}

	if err := s.repo.CreateBootstrapToken(ctx, token); err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &VerifyDeviceResponse{
		BootstrapToken:   tokenStr,
		TeacherEmail:     teacher.Email,
		DeviceName:       dev.Name,
		ExpiresInSeconds: 60,
	}, nil
}

func (s *service) ListDevices(ctx context.Context, actorRole string) ([]DeviceSummary, *shared.AppError) {
	if actorRole != "admin" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Accés prohibit.")
	}

	devices, err := s.repo.ListDevices(ctx)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	if devices == nil {
		devices = make([]DeviceSummary, 0)
	}

	return devices, nil
}

func (s *service) RegisterDevice(ctx context.Context, actorID uuid.UUID, actorRole string, req RegisterDeviceRequest) (*DeviceSummary, *shared.AppError) {
	if actorRole != "admin" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Accés prohibit.")
	}

	req.HardwareUUIDHash = strings.TrimSpace(strings.ToLower(req.HardwareUUIDHash))
	req.Name = strings.TrimSpace(req.Name)
	req.TeacherEmail = strings.TrimSpace(strings.ToLower(req.TeacherEmail))

	if req.HardwareUUIDHash == "" || len(req.HardwareUUIDHash) != 64 || !isHexString(req.HardwareUUIDHash) {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El hash del UUID de maquinari ha de ser un SHA-256 hexadecimal de 64 caràcters.",
			map[string]interface{}{"field": "hardwareUuidHash"},
		)
	}

	if req.Name == "" {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El nom descriptiu de l'ordinador és obligatori.",
			map[string]interface{}{"field": "name"},
		)
	}
	if len(req.Name) > 100 {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El nom descriptiu de l'ordinador no pot superar els 100 caràcters.",
			map[string]interface{}{"field": "name"},
		)
	}

	if req.TeacherEmail == "" {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El correu electrònic del professor és obligatori.",
			map[string]interface{}{"field": "teacherEmail"},
		)
	}
	if _, err := mail.ParseAddress(req.TeacherEmail); err != nil {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El format del correu electrònic del professor no és vàlid.",
			map[string]interface{}{"field": "teacherEmail"},
		)
	}

	teacher, err := s.repo.GetTeacherByEmail(ctx, req.TeacherEmail)
	if err != nil {
		if errors.Is(err, ErrTeacherNotFound) {
			return nil, shared.ErrBadRequest(
				shared.ErrCodeValidation,
				"El professor especificat no existeix o no té rol docent.",
				map[string]interface{}{"field": "teacherEmail"},
			)
		}
		return nil, shared.ErrInternal(err)
	}

	if teacher.Role != "teacher" {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"L'usuari especificat no té el rol docent requerit.",
			map[string]interface{}{"field": "teacherEmail"},
		)
	}

	if !teacher.IsActive {
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El compte del professor especificat està desactivat.",
			map[string]interface{}{"field": "teacherEmail"},
		)
	}

	now := time.Now().UTC()
	dev := &AuthorizedDevice{
		ID:               uuid.New(),
		HardwareUUIDHash: req.HardwareUUIDHash,
		Name:             req.Name,
		TeacherID:        teacher.ID,
		IsActive:         true,
		RegisteredBy:     actorID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	saved, err := s.repo.UpsertDevice(ctx, dev)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &DeviceSummary{
		ID:           saved.ID,
		Name:         saved.Name,
		TeacherID:    teacher.ID,
		TeacherName:  teacher.Name,
		TeacherEmail: teacher.Email,
		IsActive:     saved.IsActive,
		LastUsedAt:   saved.LastUsedAt,
		CreatedAt:    saved.CreatedAt,
	}, nil
}

func (s *service) RevokeDevice(ctx context.Context, actorRole string, deviceID uuid.UUID) *shared.AppError {
	if actorRole != "admin" {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "Accés prohibit.")
	}

	err := s.repo.RevokeDevice(ctx, deviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return shared.ErrNotFound("DEVICE_NOT_FOUND", "Dispositiu no trobat.")
		}
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) ValidateAndConsumeBootstrapToken(ctx context.Context, token string, teacherID uuid.UUID) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrInvalidBootstrapToken
	}
	return s.repo.ConsumeBootstrapToken(ctx, token, teacherID)
}

func generateBootstrapToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "dbt_" + hex.EncodeToString(b), nil
}

func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
