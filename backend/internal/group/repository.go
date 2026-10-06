package group

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrGroupNotFound = errors.New("grup no trobat")
)

// Repository interface for group database operations.
type Repository interface {
	Create(ctx context.Context, g *Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*Group, error)
	Update(ctx context.Context, g *Group) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters GroupListFilters, actorID uuid.UUID, actorRole string) ([]Group, int, error)
	ListStudents(ctx context.Context, groupID uuid.UUID) ([]GroupStudent, error)
	AddStudents(ctx context.Context, groupID uuid.UUID, studentIDs []uuid.UUID) error
	RemoveStudent(ctx context.Context, groupID uuid.UUID, studentID uuid.UUID) error
	IsStudentInGroup(ctx context.Context, groupID uuid.UUID, studentID uuid.UUID) (bool, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository initializes a new postgres repository for group module.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, g *Group) error {
	query := `
		INSERT INTO groups (id, name, academic_year, course_id, teacher_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		g.ID,
		g.Name,
		g.AcademicYear,
		g.CourseID,
		g.TeacherID,
		g.CreatedAt,
		g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creant grup: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Group, error) {
	query := `
		SELECT g.id, g.name, g.academic_year, g.course_id, c.title AS course_title, g.teacher_id,
		       (SELECT COUNT(*) FROM group_students gs WHERE gs.group_id = g.id) AS student_count,
		       g.created_at, g.updated_at
		FROM groups g
		LEFT JOIN courses c ON g.course_id = c.id AND c.deleted_at IS NULL
		WHERE g.id = $1 AND g.deleted_at IS NULL
	`

	var g Group
	var courseTitle sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&g.ID,
		&g.Name,
		&g.AcademicYear,
		&g.CourseID,
		&courseTitle,
		&g.TeacherID,
		&g.StudentCount,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("error obtenint grup: %w", err)
	}

	if courseTitle.Valid {
		g.CourseTitle = &courseTitle.String
	}

	return &g, nil
}

func (r *postgresRepository) Update(ctx context.Context, g *Group) error {
	query := `
		UPDATE groups
		SET name = $1, academic_year = $2, course_id = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, g.Name, g.AcademicYear, g.CourseID, g.UpdatedAt, g.ID)
	if err != nil {
		return fmt.Errorf("error actualitzant grup: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error comprovant files afectades: %w", err)
	}
	if rows == 0 {
		return ErrGroupNotFound
	}

	return nil
}

func (r *postgresRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE groups
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error esborrant grup (soft-delete): %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error comprovant files afectades: %w", err)
	}
	if rows == 0 {
		return ErrGroupNotFound
	}

	return nil
}

func (r *postgresRepository) List(ctx context.Context, filters GroupListFilters, actorID uuid.UUID, actorRole string) ([]Group, int, error) {
	whereClauses := []string{"g.deleted_at IS NULL"}
	args := []interface{}{}
	paramIdx := 1

	if actorRole == "student" {
		whereClauses = append(whereClauses, fmt.Sprintf("g.id IN (SELECT gs.group_id FROM group_students gs WHERE gs.student_id = $%d)", paramIdx))
		args = append(args, actorID)
		paramIdx++
	} else if actorRole == "teacher" {
		whereClauses = append(whereClauses, fmt.Sprintf("g.teacher_id = $%d", paramIdx))
		args = append(args, actorID)
		paramIdx++
	}

	if strings.TrimSpace(filters.Search) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(g.name) LIKE $%d", paramIdx))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(filters.Search))+"%")
		paramIdx++
	}

	if strings.TrimSpace(filters.AcademicYear) != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("g.academic_year = $%d", paramIdx))
		args = append(args, strings.TrimSpace(filters.AcademicYear))
		paramIdx++
	}

	if filters.CourseID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("g.course_id = $%d", paramIdx))
		args = append(args, *filters.CourseID)
		paramIdx++
	}

	whereStmt := "WHERE " + strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM groups g %s", whereStmt)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error comptant grups: %w", err)
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	selectQuery := fmt.Sprintf(`
		SELECT g.id, g.name, g.academic_year, g.course_id, c.title AS course_title, g.teacher_id,
		       (SELECT COUNT(*) FROM group_students gs WHERE gs.group_id = g.id) AS student_count,
		       g.created_at, g.updated_at
		FROM groups g
		LEFT JOIN courses c ON g.course_id = c.id AND c.deleted_at IS NULL
		%s
		ORDER BY g.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereStmt, paramIdx, paramIdx+1)

	queryArgs := append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("error llistant grups: %w", err)
	}
	defer rows.Close()

	groups := make([]Group, 0)
	for rows.Next() {
		var g Group
		var courseTitle sql.NullString

		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.AcademicYear,
			&g.CourseID,
			&courseTitle,
			&g.TeacherID,
			&g.StudentCount,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("error llegint grup: %w", err)
		}

		if courseTitle.Valid {
			g.CourseTitle = &courseTitle.String
		}

		groups = append(groups, g)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error en iterar grups: %w", err)
	}

	return groups, total, nil
}

func (r *postgresRepository) ListStudents(ctx context.Context, groupID uuid.UUID) ([]GroupStudent, error) {
	query := `
		SELECT u.id, u.email, TRIM(u.first_name || ' ' || u.last_name) AS full_name, gs.assigned_at
		FROM group_students gs
		JOIN users u ON gs.student_id = u.id
		WHERE gs.group_id = $1 AND u.deleted_at IS NULL
		ORDER BY u.last_name, u.first_name
	`

	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("error llistant alumnes del grup: %w", err)
	}
	defer rows.Close()

	students := make([]GroupStudent, 0)
	for rows.Next() {
		var s GroupStudent
		if err := rows.Scan(&s.ID, &s.Email, &s.FullName, &s.AssignedAt); err != nil {
			return nil, fmt.Errorf("error llegint alumne del grup: %w", err)
		}
		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error en iterar alumnes del grup: %w", err)
	}

	return students, nil
}

func (r *postgresRepository) AddStudents(ctx context.Context, groupID uuid.UUID, studentIDs []uuid.UUID) error {
	if len(studentIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error iniciant transacció per assignar alumnes: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO group_students (group_id, student_id, assigned_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (group_id, student_id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("error preparant instrucció SQL: %w", err)
	}
	defer stmt.Close()

	for _, sID := range studentIDs {
		if _, err := stmt.ExecContext(ctx, groupID, sID); err != nil {
			return fmt.Errorf("error afegint alumne %s al grup: %w", sID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error en fer commit de l'assignació d'alumnes: %w", err)
	}

	return nil
}

func (r *postgresRepository) RemoveStudent(ctx context.Context, groupID uuid.UUID, studentID uuid.UUID) error {
	query := `DELETE FROM group_students WHERE group_id = $1 AND student_id = $2`
	res, err := r.db.ExecContext(ctx, query, groupID, studentID)
	if err != nil {
		return fmt.Errorf("error desassignant alumne del grup: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error comprovant files afectades: %w", err)
	}
	if rows == 0 {
		return ErrGroupNotFound
	}

	return nil
}

func (r *postgresRepository) IsStudentInGroup(ctx context.Context, groupID uuid.UUID, studentID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM group_students WHERE group_id = $1 AND student_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, groupID, studentID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error comprovant pertinença d'alumne al grup: %w", err)
	}
	return exists, nil
}
