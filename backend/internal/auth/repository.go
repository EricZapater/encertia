package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("usuari no trobat")
	ErrTokenNotFound     = errors.New("refresh token no trobat")
	ErrEmailAlreadyInUse = errors.New("el correu ja està en ús")
)

// Repository defines data access methods for the auth domain.
type Repository interface {
	CreateUser(ctx context.Context, user *UserDB) error
	GetUserByEmail(ctx context.Context, email string) (*UserDB, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserDB, error)
	SaveRefreshToken(ctx context.Context, rt *RefreshTokenDB) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenDB, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string, revokedAt time.Time) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

type sqlRepository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL auth repository.
func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) CreateUser(ctx context.Context, user *UserDB) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, is_active, created_at, updated_at
	`
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	return err
}

func (r *sqlRepository) GetUserByEmail(ctx context.Context, email string) (*UserDB, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	var u UserDB
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *sqlRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*UserDB, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	var u UserDB
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *sqlRepository) SaveRefreshToken(ctx context.Context, rt *RefreshTokenDB) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	now := time.Now().UTC()
	rt.CreatedAt = now
	rt.UpdatedAt = now

	return r.db.QueryRowContext(
		ctx,
		query,
		rt.ID,
		rt.UserID,
		rt.TokenHash,
		rt.ExpiresAt,
		rt.CreatedAt,
		rt.UpdatedAt,
	).Scan(&rt.ID, &rt.CreatedAt, &rt.UpdatedAt)
}

func (r *sqlRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshTokenDB, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at, updated_at, deleted_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND deleted_at IS NULL
	`
	var rt RefreshTokenDB
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.ExpiresAt,
		&rt.RevokedAt,
		&rt.CreatedAt,
		&rt.UpdatedAt,
		&rt.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *sqlRepository) RevokeRefreshToken(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1, updated_at = $1
		WHERE token_hash = $2 AND revoked_at IS NULL AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, revokedAt, tokenHash)
	return err
}

func (r *sqlRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1, updated_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, revokedAt, userID)
	return err
}
