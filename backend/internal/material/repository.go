package material

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("material no trobat")
)

// Repository defines data access contract for the Material domain.
type Repository interface {
	Create(ctx context.Context, material *Material) error
	GetByID(ctx context.Context, id uuid.UUID) (*Material, error)
	Update(ctx context.Context, material *Material) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters MaterialListFilters, teacherID *uuid.UUID) ([]Material, int, error)
	RecordView(ctx context.Context, materialID, studentID uuid.UUID) (int, error)
	GetViewsReport(ctx context.Context, materialID uuid.UUID) (*MaterialViewsReportResponse, error)
	ListUnitMaterials(ctx context.Context, unitID uuid.UUID) ([]Material, error)
	LinkMaterialToUnit(ctx context.Context, unitID, materialID uuid.UUID, orderIndex int) error
	UnlinkMaterialFromUnit(ctx context.Context, unitID, materialID uuid.UUID) error
	GetCourseTeacherID(ctx context.Context, courseID uuid.UUID) (uuid.UUID, error)
	GetCourseIDByUnitID(ctx context.Context, unitID uuid.UUID) (uuid.UUID, error)
}

type sqlRepository struct {
	db *sql.DB
}

// NewRepository creates a new SQL-backed Repository instance.
func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, m *Material) error {
	query := `
		INSERT INTO materials (
			id, title, description, material_type, file_url, file_name, file_size_bytes,
			mime_type, page_count, video_url, video_provider, teacher_id, created_at, updated_at
		) VALUES (
			COALESCE(NULLIF($1, '00000000-0000-0000-0000-000000000000'::uuid), gen_random_uuid()),
			$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW()
		)
		RETURNING id, created_at, updated_at;
	`
	err := r.db.QueryRowContext(
		ctx, query,
		m.ID, m.Title, m.Description, m.MaterialType, m.FileURL, m.FileName,
		m.FileSizeBytes, m.MIMEType, m.PageCount, m.VideoURL, m.VideoProvider, m.TeacherID,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error inserting material: %w", err)
	}
	return nil
}

func (r *sqlRepository) GetByID(ctx context.Context, id uuid.UUID) (*Material, error) {
	query := `
		SELECT id, title, description, material_type, file_url, file_name, file_size_bytes,
		       mime_type, page_count, video_url, video_provider, teacher_id, created_at, updated_at
		FROM materials
		WHERE id = $1 AND deleted_at IS NULL;
	`
	var m Material
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Title, &m.Description, &m.MaterialType, &m.FileURL, &m.FileName,
		&m.FileSizeBytes, &m.MIMEType, &m.PageCount, &m.VideoURL, &m.VideoProvider, &m.TeacherID,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error querying material by id: %w", err)
	}
	return &m, nil
}

func (r *sqlRepository) Update(ctx context.Context, m *Material) error {
	query := `
		UPDATE materials
		SET title = $2,
		    description = $3,
		    file_url = $4,
		    file_name = $5,
		    file_size_bytes = $6,
		    mime_type = $7,
		    page_count = $8,
		    video_url = $9,
		    video_provider = $10,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at;
	`
	err := r.db.QueryRowContext(
		ctx, query,
		m.ID, m.Title, m.Description, m.FileURL, m.FileName, m.FileSizeBytes,
		m.MIMEType, m.PageCount, m.VideoURL, m.VideoProvider,
	).Scan(&m.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("error updating material: %w", err)
	}
	return nil
}

func (r *sqlRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE materials
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL;
	`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error soft-deleting material: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlRepository) List(ctx context.Context, filters MaterialListFilters, teacherID *uuid.UUID) ([]Material, int, error) {
	where := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if teacherID != nil {
		where += fmt.Sprintf(" AND teacher_id = $%d", argIdx)
		args = append(args, *teacherID)
		argIdx++
	}

	if filters.Search != "" {
		where += fmt.Sprintf(" AND (LOWER(title) LIKE $%d OR LOWER(description) LIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	if filters.MaterialType != "" {
		where += fmt.Sprintf(" AND material_type = $%d", argIdx)
		args = append(args, filters.MaterialType)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM materials %s;", where)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting materials: %w", err)
	}

	if total == 0 {
		return []Material{}, 0, nil
	}

	limit := filters.PageSize
	if limit <= 0 {
		limit = 10
	}
	offset := (filters.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, title, description, material_type, file_url, file_name, file_size_bytes,
		       mime_type, page_count, video_url, video_provider, teacher_id, created_at, updated_at
		FROM materials
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d;
	`, where, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing materials: %w", err)
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var m Material
		err := rows.Scan(
			&m.ID, &m.Title, &m.Description, &m.MaterialType, &m.FileURL, &m.FileName,
			&m.FileSizeBytes, &m.MIMEType, &m.PageCount, &m.VideoURL, &m.VideoProvider, &m.TeacherID,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning material row: %w", err)
		}
		materials = append(materials, m)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating material rows: %w", err)
	}

	return materials, total, nil
}

func (r *sqlRepository) RecordView(ctx context.Context, materialID, studentID uuid.UUID) (int, error) {
	insertQuery := `
		INSERT INTO material_views (material_id, student_id, created_at)
		VALUES ($1, $2, NOW());
	`
	_, err := r.db.ExecContext(ctx, insertQuery, materialID, studentID)
	if err != nil {
		return 0, fmt.Errorf("error recording material view: %w", err)
	}

	countQuery := `
		SELECT COUNT(*) FROM material_views
		WHERE material_id = $1 AND student_id = $2;
	`
	var count int
	err = r.db.QueryRowContext(ctx, countQuery, materialID, studentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting material views: %w", err)
	}
	return count, nil
}

func (r *sqlRepository) GetViewsReport(ctx context.Context, materialID uuid.UUID) (*MaterialViewsReportResponse, error) {
	var report MaterialViewsReportResponse
	report.MaterialID = materialID
	report.StudentViews = []StudentViewRecord{}

	totalQuery := `SELECT COUNT(*) FROM material_views WHERE material_id = $1;`
	err := r.db.QueryRowContext(ctx, totalQuery, materialID).Scan(&report.TotalViews)
	if err != nil {
		return nil, fmt.Errorf("error counting total views: %w", err)
	}

	studentsQuery := `SELECT COUNT(DISTINCT student_id) FROM material_views WHERE material_id = $1;`
	err = r.db.QueryRowContext(ctx, studentsQuery, materialID).Scan(&report.TotalStudentsViewed)
	if err != nil {
		return nil, fmt.Errorf("error counting distinct students viewed: %w", err)
	}

	listQuery := `
		SELECT 
			v.student_id, 
			COALESCE(TRIM(u.first_name || ' ' || u.last_name), u.email) as student_name,
			u.email as student_email,
			COUNT(*)::int as view_count,
			MAX(v.created_at) as last_viewed_at
		FROM material_views v
		JOIN users u ON u.id = v.student_id
		WHERE v.material_id = $1
		GROUP BY v.student_id, u.first_name, u.last_name, u.email
		ORDER BY last_viewed_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, listQuery, materialID)
	if err != nil {
		return nil, fmt.Errorf("error querying student views report: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rec StudentViewRecord
		err := rows.Scan(&rec.StudentID, &rec.StudentName, &rec.StudentEmail, &rec.ViewCount, &rec.LastViewedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning student view record: %w", err)
		}
		report.StudentViews = append(report.StudentViews, rec)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating student view records: %w", err)
	}

	return &report, nil
}

func (r *sqlRepository) ListUnitMaterials(ctx context.Context, unitID uuid.UUID) ([]Material, error) {
	query := `
		SELECT m.id, m.title, m.description, m.material_type, m.file_url, m.file_name, m.file_size_bytes,
		       m.mime_type, m.page_count, m.video_url, m.video_provider, m.teacher_id, m.created_at, m.updated_at
		FROM materials m
		JOIN unit_materials um ON um.material_id = m.id
		WHERE um.unit_id = $1 AND m.deleted_at IS NULL
		ORDER BY um.order_index ASC, um.created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("error querying unit materials: %w", err)
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var m Material
		err := rows.Scan(
			&m.ID, &m.Title, &m.Description, &m.MaterialType, &m.FileURL, &m.FileName,
			&m.FileSizeBytes, &m.MIMEType, &m.PageCount, &m.VideoURL, &m.VideoProvider, &m.TeacherID,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning unit material row: %w", err)
		}
		materials = append(materials, m)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unit material rows: %w", err)
	}

	return materials, nil
}

func (r *sqlRepository) LinkMaterialToUnit(ctx context.Context, unitID, materialID uuid.UUID, orderIndex int) error {
	query := `
		INSERT INTO unit_materials (id, unit_id, material_id, order_index, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (unit_id, material_id) DO UPDATE
		SET order_index = EXCLUDED.order_index;
	`
	_, err := r.db.ExecContext(ctx, query, unitID, materialID, orderIndex)
	if err != nil {
		return fmt.Errorf("error linking material to unit: %w", err)
	}
	return nil
}

func (r *sqlRepository) UnlinkMaterialFromUnit(ctx context.Context, unitID, materialID uuid.UUID) error {
	query := `
		DELETE FROM unit_materials
		WHERE unit_id = $1 AND material_id = $2;
	`
	res, err := r.db.ExecContext(ctx, query, unitID, materialID)
	if err != nil {
		return fmt.Errorf("error unlinking material from unit: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlRepository) GetCourseTeacherID(ctx context.Context, courseID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT teacher_id FROM courses WHERE id = $1 AND deleted_at IS NULL;`
	var teacherID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, courseID).Scan(&teacherID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("error querying course teacher id: %w", err)
	}
	return teacherID, nil
}

func (r *sqlRepository) GetCourseIDByUnitID(ctx context.Context, unitID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT course_id FROM course_units WHERE id = $1 AND deleted_at IS NULL;`
	var courseID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, unitID).Scan(&courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("error querying course id by unit id: %w", err)
	}
	return courseID, nil
}
