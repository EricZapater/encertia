package device

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrDeviceNotFound        = errors.New("dispositiu no trobat")
	ErrDeviceInactive        = errors.New("dispositiu desactivat")
	ErrTeacherNotFound       = errors.New("professor no trobat o rol invàlid")
	ErrTeacherInactive       = errors.New("el compte del professor està desactivat")
	ErrInvalidBootstrapToken = errors.New("token d'arrencada invàlid o expirat")
)

// Repository interface for device database operations.
type Repository interface {
	GetDeviceByHardwareHash(ctx context.Context, hash string) (*AuthorizedDevice, error)
	GetDeviceByID(ctx context.Context, id uuid.UUID) (*AuthorizedDevice, error)
	GetTeacherByEmail(ctx context.Context, email string) (*TeacherInfo, error)
	GetTeacherByID(ctx context.Context, id uuid.UUID) (*TeacherInfo, error)
	UpsertDevice(ctx context.Context, dev *AuthorizedDevice) (*AuthorizedDevice, error)
	ListDevices(ctx context.Context) ([]DeviceSummary, error)
	RevokeDevice(ctx context.Context, id uuid.UUID) error
	CreateBootstrapToken(ctx context.Context, token *DeviceBootstrapToken) error
	ConsumeBootstrapToken(ctx context.Context, token string, teacherID uuid.UUID) error
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new postgres repository instance for devices.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetDeviceByHardwareHash(ctx context.Context, hash string) (*AuthorizedDevice, error) {
	query := `
		SELECT id, hardware_uuid_hash, name, teacher_id, is_active, last_used_at, registered_by, created_at, updated_at
		FROM authorized_devices
		WHERE hardware_uuid_hash = $1
	`
	var dev AuthorizedDevice
	var lastUsedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&dev.ID,
		&dev.HardwareUUIDHash,
		&dev.Name,
		&dev.TeacherID,
		&dev.IsActive,
		&lastUsedAt,
		&dev.RegisteredBy,
		&dev.CreatedAt,
		&dev.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("error obtenint dispositiu per hash: %w", err)
	}
	if lastUsedAt.Valid {
		dev.LastUsedAt = &lastUsedAt.Time
	}
	return &dev, nil
}

func (r *postgresRepository) GetDeviceByID(ctx context.Context, id uuid.UUID) (*AuthorizedDevice, error) {
	query := `
		SELECT id, hardware_uuid_hash, name, teacher_id, is_active, last_used_at, registered_by, created_at, updated_at
		FROM authorized_devices
		WHERE id = $1
	`
	var dev AuthorizedDevice
	var lastUsedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&dev.ID,
		&dev.HardwareUUIDHash,
		&dev.Name,
		&dev.TeacherID,
		&dev.IsActive,
		&lastUsedAt,
		&dev.RegisteredBy,
		&dev.CreatedAt,
		&dev.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("error obtenint dispositiu per id: %w", err)
	}
	if lastUsedAt.Valid {
		dev.LastUsedAt = &lastUsedAt.Time
	}
	return &dev, nil
}

func (r *postgresRepository) GetTeacherByEmail(ctx context.Context, email string) (*TeacherInfo, error) {
	query := `
		SELECT id, email, TRIM(CONCAT(first_name, ' ', last_name)) AS name, role, is_active
		FROM users
		WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL
	`
	var t TeacherInfo
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&t.ID,
		&t.Email,
		&t.Name,
		&t.Role,
		&t.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeacherNotFound
		}
		return nil, fmt.Errorf("error cercant docent per email: %w", err)
	}
	return &t, nil
}

func (r *postgresRepository) GetTeacherByID(ctx context.Context, id uuid.UUID) (*TeacherInfo, error) {
	query := `
		SELECT id, email, TRIM(CONCAT(first_name, ' ', last_name)) AS name, role, is_active
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var t TeacherInfo
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Email,
		&t.Name,
		&t.Role,
		&t.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTeacherNotFound
		}
		return nil, fmt.Errorf("error cercant docent per id: %w", err)
	}
	return &t, nil
}

func (r *postgresRepository) UpsertDevice(ctx context.Context, dev *AuthorizedDevice) (*AuthorizedDevice, error) {
	query := `
		INSERT INTO authorized_devices (id, hardware_uuid_hash, name, teacher_id, is_active, registered_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, $5, $6, $7)
		ON CONFLICT (hardware_uuid_hash) DO UPDATE
		SET name = EXCLUDED.name,
		    teacher_id = EXCLUDED.teacher_id,
		    is_active = TRUE,
		    registered_by = EXCLUDED.registered_by,
		    updated_at = NOW()
		RETURNING id, hardware_uuid_hash, name, teacher_id, is_active, last_used_at, registered_by, created_at, updated_at
	`
	var result AuthorizedDevice
	var lastUsedAt sql.NullTime
	err := r.db.QueryRowContext(
		ctx,
		query,
		dev.ID,
		dev.HardwareUUIDHash,
		dev.Name,
		dev.TeacherID,
		dev.RegisteredBy,
		dev.CreatedAt,
		dev.UpdatedAt,
	).Scan(
		&result.ID,
		&result.HardwareUUIDHash,
		&result.Name,
		&result.TeacherID,
		&result.IsActive,
		&lastUsedAt,
		&result.RegisteredBy,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error registrant o actualitzant dispositiu: %w", err)
	}
	if lastUsedAt.Valid {
		result.LastUsedAt = &lastUsedAt.Time
	}
	return &result, nil
}

func (r *postgresRepository) ListDevices(ctx context.Context) ([]DeviceSummary, error) {
	query := `
		SELECT d.id, d.name, d.teacher_id, TRIM(CONCAT(u.first_name, ' ', u.last_name)) AS teacher_name, u.email AS teacher_email,
		       d.is_active, d.last_used_at, d.created_at
		FROM authorized_devices d
		JOIN users u ON d.teacher_id = u.id
		ORDER BY d.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error llistant dispositius: %w", err)
	}
	defer rows.Close()

	devices := make([]DeviceSummary, 0)
	for rows.Next() {
		var s DeviceSummary
		var lastUsedAt sql.NullTime
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.TeacherID,
			&s.TeacherName,
			&s.TeacherEmail,
			&s.IsActive,
			&lastUsedAt,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("error llegint fila de dispositiu: %w", err)
		}
		if lastUsedAt.Valid {
			s.LastUsedAt = &lastUsedAt.Time
		}
		devices = append(devices, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterant dispositius: %w", err)
	}
	return devices, nil
}

func (r *postgresRepository) RevokeDevice(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE authorized_devices
		SET is_active = FALSE, updated_at = NOW()
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error revocant dispositiu: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error comprovant files afectades: %w", err)
	}
	if rows == 0 {
		return ErrDeviceNotFound
	}
	return nil
}

func (r *postgresRepository) CreateBootstrapToken(ctx context.Context, token *DeviceBootstrapToken) error {
	query := `
		INSERT INTO device_bootstrap_tokens (token, device_id, teacher_id, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, token.Token, token.DeviceID, token.TeacherID, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("error guardant bootstrap token: %w", err)
	}
	return nil
}

func (r *postgresRepository) ConsumeBootstrapToken(ctx context.Context, token string, teacherID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error iniciant transacció de consum de token: %w", err)
	}
	defer tx.Rollback()

	queryToken := `
		UPDATE device_bootstrap_tokens
		SET used_at = NOW()
		WHERE token = $1 AND teacher_id = $2 AND used_at IS NULL AND expires_at > NOW()
		RETURNING device_id
	`
	var deviceID uuid.UUID
	err = tx.QueryRowContext(ctx, queryToken, token, teacherID).Scan(&deviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidBootstrapToken
		}
		return fmt.Errorf("error validant bootstrap token: %w", err)
	}

	queryDevice := `
		UPDATE authorized_devices
		SET last_used_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND is_active = TRUE
	`
	res, err := tx.ExecContext(ctx, queryDevice, deviceID)
	if err != nil {
		return fmt.Errorf("error actualitzant darrer ús de dispositiu: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error comprovant files afectades del dispositiu: %w", err)
	}
	if rows == 0 {
		return ErrDeviceInactive
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error commit consum de token: %w", err)
	}
	return nil
}
