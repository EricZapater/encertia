package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserNotFound      = errors.New("usuari no trobat")
	ErrEmailAlreadyInUse = errors.New("el correu ja està en ús")
)

// Repository defines data access methods for the User domain.
type Repository interface {
	ListUsers(ctx context.Context, filter ListUsersFilter) ([]UserDB, int, error)
	CreateUser(ctx context.Context, user *UserDB) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*UserDB, error)
	GetUserByEmail(ctx context.Context, email string) (*UserDB, error)
	UpdateUser(ctx context.Context, user *UserDB) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	SoftDeleteUser(ctx context.Context, id uuid.UUID, deletedAt time.Time) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

type sqlRepository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL User repository.
func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) ListUsers(ctx context.Context, filter ListUsersFilter) ([]UserDB, int, error) {
	var whereClauses []string
	var args []interface{}
	argIdx := 1

	// Always exclude soft-deleted users
	whereClauses = append(whereClauses, "deleted_at IS NULL")

	// Role scoping
	if filter.ActorRole == string(RoleTeacher) {
		whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, string(RoleStudent))
		argIdx++
	} else if filter.Role != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, filter.Role)
		argIdx++
	}

	// Status filter
	switch filter.Status {
	case "inactive":
		whereClauses = append(whereClauses, "is_active = FALSE")
	case "all":
		// No filter on is_active
	default: // "active" or empty
		whereClauses = append(whereClauses, "is_active = TRUE")
	}

	// Search filter
	if strings.TrimSpace(filter.Search) != "" {
		searchTerm := "%" + strings.ToLower(strings.TrimSpace(filter.Search)) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(LOWER(first_name) LIKE $%d OR LOWER(last_name) LIKE $%d OR LOWER(email) LIKE $%d)",
			argIdx, argIdx, argIdx,
		))
		args = append(args, searchTerm)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// 1. Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereSQL)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []UserDB{}, 0, nil
	}

	// 2. Select query
	limit := filter.PageSize
	if limit <= 0 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	selectSQL := fmt.Sprintf(`
		SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at, deleted_at
		FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectSQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]UserDB, 0, limit)
	for rows.Next() {
		var u UserDB
		if err := rows.Scan(
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
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
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

	return r.db.QueryRowContext(
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

func (r *sqlRepository) GetUserByEmail(ctx context.Context, email string) (*UserDB, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL
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

func (r *sqlRepository) UpdateUser(ctx context.Context, user *UserDB) error {
	query := `
		UPDATE users
		SET email = $1, first_name = $2, last_name = $3, role = $4, is_active = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`
	now := time.Now().UTC()
	user.UpdatedAt = now

	res, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *sqlRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, query, passwordHash, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *sqlRepository) SoftDeleteUser(ctx context.Context, id uuid.UUID, deletedAt time.Time) error {
	query := `
		UPDATE users
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, deletedAt, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *sqlRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $1, updated_at = $1
		WHERE user_id = $2 AND revoked_at IS NULL AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, revokedAt, userID)
	return err
}
