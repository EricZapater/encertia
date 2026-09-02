package course

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
	ErrCourseNotFound     = errors.New("course not found")
	ErrCourseCodeConflict = errors.New("course code already exists")
	ErrUnitNotFound       = errors.New("course unit not found")
)

// Repository defines data access operations for the Course domain.
type Repository interface {
	CreateCourse(ctx context.Context, c *Course) error
	GetCourseByID(ctx context.Context, id uuid.UUID) (*Course, error)
	GetCourseByCode(ctx context.Context, code string) (*Course, error)
	UpdateCourse(ctx context.Context, c *Course) error
	DeleteCourse(ctx context.Context, id uuid.UUID) error
	ListCourses(ctx context.Context, actorID uuid.UUID, role string, filters CourseListFilters) ([]Course, int, error)

	EnrollStudents(ctx context.Context, courseID uuid.UUID, studentIDs []uuid.UUID) error
	UnenrollStudent(ctx context.Context, courseID, studentID uuid.UUID) error
	GetCourseStudents(ctx context.Context, courseID uuid.UUID) ([]EnrolledStudentResponse, error)
	IsStudentEnrolled(ctx context.Context, courseID, studentID uuid.UUID) (bool, error)

	CreateUnit(ctx context.Context, unit *CourseUnit) error
	GetUnitByID(ctx context.Context, unitID uuid.UUID) (*CourseUnit, error)
	ListUnitsByCourseID(ctx context.Context, courseID uuid.UUID) ([]CourseUnit, error)
	UpdateUnit(ctx context.Context, unit *CourseUnit) error
	DeleteUnit(ctx context.Context, unitID uuid.UUID) error
	ReorderUnits(ctx context.Context, courseID uuid.UUID, unitIDs []uuid.UUID) error

	LinkQuiz(ctx context.Context, unitID, quizID uuid.UUID) error
	UnlinkQuiz(ctx context.Context, unitID, quizID uuid.UUID) error
	GetLinkedQuizzes(ctx context.Context, unitID uuid.UUID) ([]LinkedQuiz, error)

	GetScriptBlocks(ctx context.Context, unitID uuid.UUID) ([]ScriptBlock, error)
	ReplaceScriptBlocks(ctx context.Context, unitID uuid.UUID, blocks []ScriptBlock) ([]ScriptBlock, error)
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new instance of Course Repository.
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateCourse(ctx context.Context, c *Course) error {
	query := `
		INSERT INTO courses (id, title, code, description, status, start_date, end_date, teacher_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		c.ID,
		c.Title,
		c.Code,
		c.Description,
		c.Status,
		c.StartDate,
		c.EndDate,
		c.TeacherID,
		c.CreatedAt,
		c.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "idx_courses_code_active") || strings.Contains(err.Error(), "duplicate key") {
			return ErrCourseCodeConflict
		}
		return err
	}
	return nil
}

func (r *repository) GetCourseByID(ctx context.Context, id uuid.UUID) (*Course, error) {
	query := `
		SELECT 
			c.id, c.title, c.code, c.description, c.status, c.start_date, c.end_date, c.teacher_id,
			CONCAT(u.first_name, ' ', u.last_name) AS teacher_name,
			(SELECT COUNT(*) FROM course_enrollments WHERE course_id = c.id) AS enrolled_students_count,
			(SELECT COUNT(*) FROM course_units WHERE course_id = c.id AND deleted_at IS NULL) AS units_count,
			c.created_at, c.updated_at
		FROM courses c
		LEFT JOIN users u ON c.teacher_id = u.id
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`
	var c Course
	var teacherName sql.NullString
	var desc, startDate, endDate sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID,
		&c.Title,
		&c.Code,
		&desc,
		&c.Status,
		&startDate,
		&endDate,
		&c.TeacherID,
		&teacherName,
		&c.EnrolledStudentsCount,
		&c.UnitsCount,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCourseNotFound
		}
		return nil, err
	}

	if desc.Valid {
		c.Description = &desc.String
	}
	if startDate.Valid {
		c.StartDate = &startDate.String
	}
	if endDate.Valid {
		c.EndDate = &endDate.String
	}
	if teacherName.Valid {
		c.TeacherName = &teacherName.String
	}

	return &c, nil
}

func (r *repository) GetCourseByCode(ctx context.Context, code string) (*Course, error) {
	query := `
		SELECT id, title, code, description, status, teacher_id, created_at, updated_at
		FROM courses
		WHERE LOWER(code) = LOWER($1) AND deleted_at IS NULL
	`
	var c Course
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&c.ID,
		&c.Title,
		&c.Code,
		&desc,
		&c.Status,
		&c.TeacherID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCourseNotFound
		}
		return nil, err
	}
	if desc.Valid {
		c.Description = &desc.String
	}
	return &c, nil
}

func (r *repository) UpdateCourse(ctx context.Context, c *Course) error {
	query := `
		UPDATE courses
		SET title = $1, code = $2, description = $3, status = $4, start_date = $5, end_date = $6, updated_at = NOW()
		WHERE id = $7 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query,
		c.Title,
		c.Code,
		c.Description,
		c.Status,
		c.StartDate,
		c.EndDate,
		c.ID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "idx_courses_code_active") || strings.Contains(err.Error(), "duplicate key") {
			return ErrCourseCodeConflict
		}
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func (r *repository) DeleteCourse(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE courses SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCourseNotFound
	}
	return nil
}

func (r *repository) ListCourses(ctx context.Context, actorID uuid.UUID, role string, filters CourseListFilters) ([]Course, int, error) {
	whereClauses := []string{"c.deleted_at IS NULL"}
	args := []interface{}{}
	argIdx := 1

	if role == "teacher" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.teacher_id = $%d", argIdx))
		args = append(args, actorID)
		argIdx++
	} else if role == "student" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.id IN (SELECT course_id FROM course_enrollments WHERE student_id = $%d)", argIdx))
		args = append(args, actorID)
		argIdx++
	}

	if filters.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("c.status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}

	if filters.Search != "" {
		pattern := "%" + strings.ToLower(filters.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(c.title) LIKE $%d OR LOWER(c.code) LIKE $%d)", argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM courses c WHERE %s", whereSQL)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []Course{}, 0, nil
	}

	limit := filters.PageSize
	if limit <= 0 {
		limit = 10
	}
	page := filters.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	selectQuery := fmt.Sprintf(`
		SELECT 
			c.id, c.title, c.code, c.description, c.status, c.start_date, c.end_date, c.teacher_id,
			CONCAT(u.first_name, ' ', u.last_name) AS teacher_name,
			(SELECT COUNT(*) FROM course_enrollments WHERE course_id = c.id) AS enrolled_students_count,
			(SELECT COUNT(*) FROM course_units WHERE course_id = c.id AND deleted_at IS NULL) AS units_count,
			c.created_at, c.updated_at
		FROM courses c
		LEFT JOIN users u ON c.teacher_id = u.id
		WHERE %s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	courses := make([]Course, 0)
	for rows.Next() {
		var c Course
		var teacherName sql.NullString
		var desc, startDate, endDate sql.NullString

		err := rows.Scan(
			&c.ID,
			&c.Title,
			&c.Code,
			&desc,
			&c.Status,
			&startDate,
			&endDate,
			&c.TeacherID,
			&teacherName,
			&c.EnrolledStudentsCount,
			&c.UnitsCount,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if desc.Valid {
			c.Description = &desc.String
		}
		if startDate.Valid {
			c.StartDate = &startDate.String
		}
		if endDate.Valid {
			c.EndDate = &endDate.String
		}
		if teacherName.Valid {
			c.TeacherName = &teacherName.String
		}

		courses = append(courses, c)
	}

	return courses, total, nil
}

func (r *repository) EnrollStudents(ctx context.Context, courseID uuid.UUID, studentIDs []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO course_enrollments (course_id, student_id) VALUES ($1, $2) ON CONFLICT (course_id, student_id) DO NOTHING`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, sID := range studentIDs {
		_, err := stmt.ExecContext(ctx, courseID, sID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) UnenrollStudent(ctx context.Context, courseID, studentID uuid.UUID) error {
	query := `DELETE FROM course_enrollments WHERE course_id = $1 AND student_id = $2`
	res, err := r.db.ExecContext(ctx, query, courseID, studentID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("student enrollment not found")
	}
	return nil
}

func (r *repository) GetCourseStudents(ctx context.Context, courseID uuid.UUID) ([]EnrolledStudentResponse, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email, ce.enrolled_at
		FROM course_enrollments ce
		JOIN users u ON ce.student_id = u.id
		WHERE ce.course_id = $1
		ORDER BY ce.enrolled_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]EnrolledStudentResponse, 0)
	for rows.Next() {
		var s EnrolledStudentResponse
		if err := rows.Scan(&s.ID, &s.FirstName, &s.LastName, &s.Email, &s.EnrolledAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}

	return students, nil
}

func (r *repository) IsStudentEnrolled(ctx context.Context, courseID, studentID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM course_enrollments WHERE course_id = $1 AND student_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, courseID, studentID).Scan(&exists)
	return exists, err
}

func (r *repository) CreateUnit(ctx context.Context, unit *CourseUnit) error {
	query := `
		INSERT INTO course_units (id, course_id, title, description, order_index, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if unit.ID == uuid.Nil {
		unit.ID = uuid.New()
	}
	now := time.Now().UTC()
	unit.CreatedAt = now
	unit.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		unit.ID,
		unit.CourseID,
		unit.Title,
		unit.Description,
		unit.OrderIndex,
		unit.CreatedAt,
		unit.UpdatedAt,
	)
	return err
}

func (r *repository) GetUnitByID(ctx context.Context, unitID uuid.UUID) (*CourseUnit, error) {
	query := `
		SELECT 
			u.id, u.course_id, u.title, u.description, u.order_index,
			(SELECT COUNT(*) FROM unit_quizzes WHERE unit_id = u.id) AS quizzes_count,
			(SELECT COUNT(*) FROM script_blocks WHERE unit_id = u.id) AS blocks_count,
			u.created_at, u.updated_at
		FROM course_units u
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`
	var u CourseUnit
	var desc sql.NullString

	err := r.db.QueryRowContext(ctx, query, unitID).Scan(
		&u.ID,
		&u.CourseID,
		&u.Title,
		&desc,
		&u.OrderIndex,
		&u.QuizzesCount,
		&u.BlocksCount,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUnitNotFound
		}
		return nil, err
	}

	if desc.Valid {
		u.Description = &desc.String
	}

	return &u, nil
}

func (r *repository) ListUnitsByCourseID(ctx context.Context, courseID uuid.UUID) ([]CourseUnit, error) {
	query := `
		SELECT 
			u.id, u.course_id, u.title, u.description, u.order_index,
			(SELECT COUNT(*) FROM unit_quizzes WHERE unit_id = u.id) AS quizzes_count,
			(SELECT COUNT(*) FROM script_blocks WHERE unit_id = u.id) AS blocks_count,
			u.created_at, u.updated_at
		FROM course_units u
		WHERE u.course_id = $1 AND u.deleted_at IS NULL
		ORDER BY u.order_index ASC, u.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	units := make([]CourseUnit, 0)
	for rows.Next() {
		var u CourseUnit
		var desc sql.NullString
		err := rows.Scan(
			&u.ID,
			&u.CourseID,
			&u.Title,
			&desc,
			&u.OrderIndex,
			&u.QuizzesCount,
			&u.BlocksCount,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if desc.Valid {
			u.Description = &desc.String
		}
		units = append(units, u)
	}

	return units, nil
}

func (r *repository) UpdateUnit(ctx context.Context, unit *CourseUnit) error {
	query := `
		UPDATE course_units
		SET title = $1, description = $2, order_index = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`
	res, err := r.db.ExecContext(ctx, query, unit.Title, unit.Description, unit.OrderIndex, unit.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUnitNotFound
	}
	return nil
}

func (r *repository) DeleteUnit(ctx context.Context, unitID uuid.UUID) error {
	query := `UPDATE course_units SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, unitID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUnitNotFound
	}
	return nil
}

func (r *repository) ReorderUnits(ctx context.Context, courseID uuid.UUID, unitIDs []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE course_units SET order_index = $1, updated_at = NOW() WHERE id = $2 AND course_id = $3 AND deleted_at IS NULL`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for idx, uID := range unitIDs {
		_, err := stmt.ExecContext(ctx, idx, uID, courseID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repository) LinkQuiz(ctx context.Context, unitID, quizID uuid.UUID) error {
	query := `INSERT INTO unit_quizzes (unit_id, quiz_id) VALUES ($1, $2) ON CONFLICT (unit_id, quiz_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, unitID, quizID)
	return err
}

func (r *repository) UnlinkQuiz(ctx context.Context, unitID, quizID uuid.UUID) error {
	query := `DELETE FROM unit_quizzes WHERE unit_id = $1 AND quiz_id = $2`
	_, err := r.db.ExecContext(ctx, query, unitID, quizID)
	return err
}

func (r *repository) GetLinkedQuizzes(ctx context.Context, unitID uuid.UUID) ([]LinkedQuiz, error) {
	query := `
		SELECT q.id, q.title, (SELECT COUNT(*) FROM quiz_questions WHERE quiz_id = q.id) AS questions_count
		FROM unit_quizzes uq
		JOIN quizzes q ON uq.quiz_id = q.id
		WHERE uq.unit_id = $1 AND q.deleted_at IS NULL
		ORDER BY uq.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := make([]LinkedQuiz, 0)
	for rows.Next() {
		var lq LinkedQuiz
		if err := rows.Scan(&lq.ID, &lq.Title, &lq.QuestionsCount); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, lq)
	}

	return quizzes, nil
}

func (r *repository) GetScriptBlocks(ctx context.Context, unitID uuid.UUID) ([]ScriptBlock, error) {
	query := `
		SELECT sb.id, sb.unit_id, sb.block_type, sb.order_index, sb.title, sb.description,
		       sb.material_id, sb.pdf_url, sb.start_page, sb.end_page, sb.quiz_id, q.title as quiz_title,
		       sb.duration_minutes, sb.created_at
		FROM script_blocks sb
		LEFT JOIN quizzes q ON sb.quiz_id = q.id
		WHERE sb.unit_id = $1
		ORDER BY sb.order_index ASC, sb.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocks := make([]ScriptBlock, 0)
	for rows.Next() {
		var sb ScriptBlock
		var desc, pdfURL, quizTitle sql.NullString
		var matID, quizID uuid.NullUUID
		var startPage, endPage, durationMinutes sql.NullInt32

		err := rows.Scan(
			&sb.ID,
			&sb.UnitID,
			&sb.BlockType,
			&sb.OrderIndex,
			&sb.Title,
			&desc,
			&matID,
			&pdfURL,
			&startPage,
			&endPage,
			&quizID,
			&quizTitle,
			&durationMinutes,
			&sb.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if desc.Valid {
			sb.Description = &desc.String
		}
		if pdfURL.Valid {
			sb.PdfURL = &pdfURL.String
		}
		if quizTitle.Valid {
			sb.QuizTitle = &quizTitle.String
		}
		if matID.Valid {
			sb.MaterialID = &matID.UUID
		}
		if quizID.Valid {
			sb.QuizID = &quizID.UUID
		}
		if startPage.Valid {
			sp := int(startPage.Int32)
			sb.StartPage = &sp
		}
		if endPage.Valid {
			ep := int(endPage.Int32)
			sb.EndPage = &ep
		}
		if durationMinutes.Valid {
			dm := int(durationMinutes.Int32)
			sb.DurationMinutes = &dm
		}

		blocks = append(blocks, sb)
	}

	return blocks, nil
}

func (r *repository) ReplaceScriptBlocks(ctx context.Context, unitID uuid.UUID, blocks []ScriptBlock) ([]ScriptBlock, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	deleteQuery := `DELETE FROM script_blocks WHERE unit_id = $1`
	if _, err := tx.ExecContext(ctx, deleteQuery, unitID); err != nil {
		return nil, err
	}

	insertQuery := `
		INSERT INTO script_blocks (
			id, unit_id, block_type, order_index, title, description, material_id, pdf_url, start_page, end_page, quiz_id, duration_minutes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	now := time.Now().UTC()
	inserted := make([]ScriptBlock, 0, len(blocks))

	for idx, b := range blocks {
		if b.ID == uuid.Nil {
			b.ID = uuid.New()
		}
		b.UnitID = unitID
		b.OrderIndex = idx
		b.CreatedAt = now
		b.UpdatedAt = now

		_, err := stmt.ExecContext(ctx,
			b.ID,
			b.UnitID,
			b.BlockType,
			b.OrderIndex,
			b.Title,
			b.Description,
			b.MaterialID,
			b.PdfURL,
			b.StartPage,
			b.EndPage,
			b.QuizID,
			b.DurationMinutes,
			b.CreatedAt,
			b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		inserted = append(inserted, b)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Fetch again with quiz title join
	return r.GetScriptBlocks(ctx, unitID)
}
