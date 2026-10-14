package device

import (
	"time"

	"github.com/google/uuid"
)

// AuthorizedDevice represents a physical computer authorized for teaching.
type AuthorizedDevice struct {
	ID               uuid.UUID  `json:"id"`
	HardwareUUIDHash string     `json:"hardwareUuidHash"`
	Name             string     `json:"name"`
	TeacherID        uuid.UUID  `json:"teacherId"`
	IsActive         bool       `json:"isActive"`
	LastUsedAt       *time.Time `json:"lastUsedAt,omitempty"`
	RegisteredBy     uuid.UUID  `json:"registeredBy"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// DeviceBootstrapToken represents a short-lived single-use bootstrap token.
type DeviceBootstrapToken struct {
	Token     string     `json:"token"`
	DeviceID  uuid.UUID  `json:"deviceId"`
	TeacherID uuid.UUID  `json:"teacherId"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
}

// VerifyDeviceRequest defines the request body for POST /api/v1/devices/verify.
type VerifyDeviceRequest struct {
	HardwareUUIDHash string `json:"hardwareUuidHash"`
}

// VerifyDeviceResponse defines the response body for POST /api/v1/devices/verify.
type VerifyDeviceResponse struct {
	BootstrapToken   string `json:"bootstrapToken"`
	TeacherEmail     string `json:"teacherEmail"`
	DeviceName       string `json:"deviceName"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

// RegisterDeviceRequest defines the request body for POST /api/v1/admin/devices.
type RegisterDeviceRequest struct {
	HardwareUUIDHash string `json:"hardwareUuidHash"`
	Name             string `json:"name"`
	TeacherEmail     string `json:"teacherEmail"`
}

// DeviceSummary defines the summary view of an authorized device for admin endpoints.
type DeviceSummary struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	TeacherID    uuid.UUID  `json:"teacherId"`
	TeacherName  string     `json:"teacherName"`
	TeacherEmail string     `json:"teacherEmail"`
	IsActive     bool       `json:"isActive"`
	LastUsedAt   *time.Time `json:"lastUsedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// TeacherInfo represents basic teacher information retrieved from the users table.
type TeacherInfo struct {
	ID       uuid.UUID
	Email    string
	Name     string
	Role     string
	IsActive bool
}
