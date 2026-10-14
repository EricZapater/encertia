package device_test

import (
	"context"
	"sync"
	"time"

	"github.com/encertia/backend/internal/device"
	"github.com/google/uuid"
)

type mockRepository struct {
	mu              sync.RWMutex
	devices         map[string]*device.AuthorizedDevice // hardwareHash -> device
	devicesByID     map[uuid.UUID]*device.AuthorizedDevice
	teachersByEmail map[string]*device.TeacherInfo // lower(email) -> teacher
	teachersByID    map[uuid.UUID]*device.TeacherInfo
	bootstrapTokens map[string]*device.DeviceBootstrapToken // token -> tokenObj
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		devices:         make(map[string]*device.AuthorizedDevice),
		devicesByID:     make(map[uuid.UUID]*device.AuthorizedDevice),
		teachersByEmail: make(map[string]*device.TeacherInfo),
		teachersByID:    make(map[uuid.UUID]*device.TeacherInfo),
		bootstrapTokens: make(map[string]*device.DeviceBootstrapToken),
	}
}

func (m *mockRepository) GetDeviceByHardwareHash(ctx context.Context, hash string) (*device.AuthorizedDevice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dev, exists := m.devices[hash]
	if !exists {
		return nil, device.ErrDeviceNotFound
	}
	copied := *dev
	return &copied, nil
}

func (m *mockRepository) GetDeviceByID(ctx context.Context, id uuid.UUID) (*device.AuthorizedDevice, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dev, exists := m.devicesByID[id]
	if !exists {
		return nil, device.ErrDeviceNotFound
	}
	copied := *dev
	return &copied, nil
}

func (m *mockRepository) GetTeacherByEmail(ctx context.Context, email string) (*device.TeacherInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, exists := m.teachersByEmail[email]
	if !exists {
		return nil, device.ErrTeacherNotFound
	}
	copied := *t
	return &copied, nil
}

func (m *mockRepository) GetTeacherByID(ctx context.Context, id uuid.UUID) (*device.TeacherInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, exists := m.teachersByID[id]
	if !exists {
		return nil, device.ErrTeacherNotFound
	}
	copied := *t
	return &copied, nil
}

func (m *mockRepository) UpsertDevice(ctx context.Context, dev *device.AuthorizedDevice) (*device.AuthorizedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.devices[dev.HardwareUUIDHash]
	now := time.Now().UTC()
	if exists {
		existing.Name = dev.Name
		existing.TeacherID = dev.TeacherID
		existing.IsActive = true
		existing.RegisteredBy = dev.RegisteredBy
		existing.UpdatedAt = now
		copied := *existing
		return &copied, nil
	}

	if dev.ID == uuid.Nil {
		dev.ID = uuid.New()
	}
	dev.IsActive = true
	dev.CreatedAt = now
	dev.UpdatedAt = now

	saved := *dev
	m.devices[dev.HardwareUUIDHash] = &saved
	m.devicesByID[dev.ID] = &saved
	return &saved, nil
}

func (m *mockRepository) ListDevices(ctx context.Context) ([]device.DeviceSummary, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summaries := make([]device.DeviceSummary, 0, len(m.devicesByID))
	for _, dev := range m.devicesByID {
		teacherName := ""
		teacherEmail := ""
		if t, ok := m.teachersByID[dev.TeacherID]; ok {
			teacherName = t.Name
			teacherEmail = t.Email
		}
		summaries = append(summaries, device.DeviceSummary{
			ID:           dev.ID,
			Name:         dev.Name,
			TeacherID:    dev.TeacherID,
			TeacherName:  teacherName,
			TeacherEmail: teacherEmail,
			IsActive:     dev.IsActive,
			LastUsedAt:   dev.LastUsedAt,
			CreatedAt:    dev.CreatedAt,
		})
	}
	return summaries, nil
}

func (m *mockRepository) RevokeDevice(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dev, exists := m.devicesByID[id]
	if !exists {
		return device.ErrDeviceNotFound
	}
	dev.IsActive = false
	dev.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *mockRepository) CreateBootstrapToken(ctx context.Context, token *device.DeviceBootstrapToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	saved := *token
	m.bootstrapTokens[token.Token] = &saved
	return nil
}

func (m *mockRepository) ConsumeBootstrapToken(ctx context.Context, token string, teacherID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tok, exists := m.bootstrapTokens[token]
	if !exists {
		return device.ErrInvalidBootstrapToken
	}

	if tok.TeacherID != teacherID {
		return device.ErrInvalidBootstrapToken
	}

	if tok.UsedAt != nil {
		return device.ErrInvalidBootstrapToken
	}

	if time.Now().UTC().After(tok.ExpiresAt) {
		return device.ErrInvalidBootstrapToken
	}

	dev, devExists := m.devicesByID[tok.DeviceID]
	if !devExists || !dev.IsActive {
		return device.ErrDeviceInactive
	}

	now := time.Now().UTC()
	tok.UsedAt = &now
	dev.LastUsedAt = &now
	dev.UpdatedAt = now

	return nil
}
