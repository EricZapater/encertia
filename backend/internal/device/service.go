package device

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
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

	log.Printf("[DEVICE VERIFY] Sol·licitud de verificació rebuda per a hash %s", req.HardwareUUIDHash)
	dev, err := s.repo.GetDeviceByHardwareHash(ctx, req.HardwareUUIDHash)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			log.Printf("[DEVICE VERIFY WARNING] Hash no trobat a la base de dades: %s", req.HardwareUUIDHash)
			return nil, shared.NewAppError(
				http.StatusForbidden,
				shared.ErrCodeDeviceNotAuthorized,
				"Aquest equip no està autoritzat per a la docència.",
				nil,
				nil,
			)
		}
		log.Printf("[DEVICE VERIFY ERROR] Error consultant hash %s: %v", req.HardwareUUIDHash, err)
		return nil, shared.ErrInternal(err)
	}

	if !dev.IsActive {
		log.Printf("[DEVICE VERIFY WARNING] Dispositiu %s (%q) marcat com a inactiu/revocat", dev.ID, dev.Name)
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
			log.Printf("[DEVICE VERIFY WARNING] Docent ID %s no trobat per al dispositiu %s", dev.TeacherID, dev.ID)
			return nil, shared.NewAppError(
				http.StatusForbidden,
				shared.ErrCodeDeviceNotAuthorized,
				"Aquest equip no està autoritzat per a la docència.",
				nil,
				nil,
			)
		}
		log.Printf("[DEVICE VERIFY ERROR] Error consultant docent ID %s: %v", dev.TeacherID, err)
		return nil, shared.ErrInternal(err)
	}

	if !teacher.IsActive || (teacher.Role != "teacher" && teacher.Role != "admin") {
		log.Printf("[DEVICE VERIFY WARNING] Docent %s (%s) no té rol docent/admin o està desactivat (rol: %s, actiu: %t)",
			teacher.Name, teacher.Email, teacher.Role, teacher.IsActive)
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
		log.Printf("[DEVICE VERIFY ERROR] Error generant bootstrap token: %v", err)
		return nil, shared.ErrInternal(err)
	}

	token := &DeviceBootstrapToken{
		Token:     tokenStr,
		DeviceID:  dev.ID,
		TeacherID: dev.TeacherID,
		ExpiresAt: time.Now().UTC().Add(60 * time.Second),
	}

	if err := s.repo.CreateBootstrapToken(ctx, token); err != nil {
		log.Printf("[DEVICE VERIFY ERROR] Error registrant bootstrap token a BD: %v", err)
		return nil, shared.ErrInternal(err)
	}

	log.Printf("[DEVICE VERIFY SUCCESS] Dispositiu verificat: %s (%q), Docent: %s (%s). Bootstrap token generat (TTL: 60s).",
		dev.ID, dev.Name, teacher.Name, teacher.Email)

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
	log.Printf("[DEVICE REGISTER] Iniciant procés de registre: AdminID=%s, Rol=%s", actorID, actorRole)

	if actorRole != "admin" {
		log.Printf("[DEVICE REGISTER WARNING] Intent denegat: usuari %s no té rol admin (rol: %s)", actorID, actorRole)
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Accés prohibit.")
	}

	req.HardwareUUIDHash = strings.TrimSpace(strings.ToLower(req.HardwareUUIDHash))
	req.Name = strings.TrimSpace(req.Name)
	req.TeacherEmail = strings.TrimSpace(strings.ToLower(req.TeacherEmail))

	log.Printf("[DEVICE REGISTER] Dades de la sol·licitud: Nom=%q, Docent=%q, Hash=%s", req.Name, req.TeacherEmail, req.HardwareUUIDHash)

	if req.HardwareUUIDHash == "" || len(req.HardwareUUIDHash) != 64 || !isHexString(req.HardwareUUIDHash) {
		log.Printf("[DEVICE REGISTER WARNING] Hash invàlid rebut: %q", req.HardwareUUIDHash)
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El hash del UUID de maquinari ha de ser un SHA-256 hexadecimal de 64 caràcters.",
			map[string]interface{}{"field": "hardwareUuidHash"},
		)
	}

	if req.Name == "" {
		log.Printf("[DEVICE REGISTER WARNING] Nom buit")
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El nom descriptiu de l'ordinador és obligatori.",
			map[string]interface{}{"field": "name"},
		)
	}
	if len(req.Name) > 100 {
		log.Printf("[DEVICE REGISTER WARNING] Nom massa llarg (%d caràcters)", len(req.Name))
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El nom descriptiu de l'ordinador no pot superar els 100 caràcters.",
			map[string]interface{}{"field": "name"},
		)
	}

	if req.TeacherEmail == "" {
		log.Printf("[DEVICE REGISTER WARNING] Email de professor buit")
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El correu electrònic del professor és obligatori.",
			map[string]interface{}{"field": "teacherEmail"},
		)
	}
	if _, err := mail.ParseAddress(req.TeacherEmail); err != nil {
		log.Printf("[DEVICE REGISTER WARNING] Format d'email invàlid: %q", req.TeacherEmail)
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"El format del correu electrònic del professor no és vàlid.",
			map[string]interface{}{"field": "teacherEmail"},
		)
	}

	log.Printf("[DEVICE REGISTER] Cercant docent a la base de dades amb email: %s", req.TeacherEmail)
	teacher, err := s.repo.GetTeacherByEmail(ctx, req.TeacherEmail)
	if err != nil {
		if errors.Is(err, ErrTeacherNotFound) {
			log.Printf("[DEVICE REGISTER WARNING] El professor %q no existeix a la base de dades", req.TeacherEmail)
			return nil, shared.ErrBadRequest(
				shared.ErrCodeValidation,
				"El professor especificat no existeix a la plataforma.",
				map[string]interface{}{"field": "teacherEmail"},
			)
		}
		log.Printf("[DEVICE REGISTER ERROR] Error cercant professor %s a la base de dades: %v", req.TeacherEmail, err)
		return nil, shared.ErrInternal(err)
	}

	log.Printf("[DEVICE REGISTER] Docent trobat: ID=%s, Nom=%q, Rol=%s, Actiu=%t",
		teacher.ID, teacher.Name, teacher.Role, teacher.IsActive)

	if teacher.Role != "teacher" && teacher.Role != "admin" {
		log.Printf("[DEVICE REGISTER WARNING] L'usuari %s té rol %q (es requereix 'teacher' o 'admin')", teacher.Email, teacher.Role)
		return nil, shared.ErrBadRequest(
			shared.ErrCodeValidation,
			"L'usuari especificat no té el rol docent o administrador requerit.",
			map[string]interface{}{"field": "teacherEmail"},
		)
	}

	if !teacher.IsActive {
		log.Printf("[DEVICE REGISTER WARNING] El compte de %s està desactivat", teacher.Email)
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

	log.Printf("[DEVICE REGISTER] Guardant/actualitzant dispositiu a la base de dades (Hash=%s, Nom=%q, TeacherID=%s)...",
		dev.HardwareUUIDHash, dev.Name, dev.TeacherID)

	saved, err := s.repo.UpsertDevice(ctx, dev)
	if err != nil {
		log.Printf("[DEVICE REGISTER ERROR] UpsertDevice ha fallat (hash %s, teacherID %s): %v",
			dev.HardwareUUIDHash, dev.TeacherID, err)
		return nil, shared.ErrInternal(err)
	}

	log.Printf("[DEVICE REGISTER SUCCESS] Dispositiu desat correctament: ID=%s, Nom=%q, Docent=%s (%s)",
		saved.ID, saved.Name, teacher.Name, teacher.Email)

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
