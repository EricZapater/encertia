package quiz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository defines data access methods for the quiz domain.
type Repository interface {
	ListQuizzes(ctx context.Context, filter QuizListFilters) ([]Quiz, int, error)
	GetQuizByID(ctx context.Context, id uuid.UUID) (*QuizDetail, error)
	CreateQuiz(ctx context.Context, creatorID uuid.UUID, input CreateQuizInput) (*QuizDetail, error)
	UpdateQuiz(ctx context.Context, id uuid.UUID, input UpdateQuizInput) (*QuizDetail, error)
	DeleteQuiz(ctx context.Context, id uuid.UUID) error
	DuplicateQuiz(ctx context.Context, originalID, newCreatorID uuid.UUID, newTitle string, includeAnswers bool) (*QuizDetail, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL quiz repository instance.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

// ListQuizzes retrieves a paginated list of quizzes matching the filter.
func (r *postgresRepository) ListQuizzes(ctx context.Context, filter QuizListFilters) ([]Quiz, int, error) {
	if r.db == nil {
		return nil, 0, errors.New("connexió a la base de dades no inicialitzada")
	}

	whereClauses := []string{"q.deleted_at IS NULL"}
	args := []interface{}{}
	argIdx := 1

	if filter.CreatorID != nil && *filter.CreatorID != uuid.Nil {
		whereClauses = append(whereClauses, fmt.Sprintf("q.creator_id = $%d", argIdx))
		args = append(args, *filter.CreatorID)
		argIdx++
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("q.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Tag != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("$%d = ANY(q.tags)", argIdx))
		args = append(args, filter.Tag)
		argIdx++
	}

	if filter.Search != "" {
		searchTerm := "%" + strings.TrimSpace(filter.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(q.title ILIKE $%d OR COALESCE(q.description, '') ILIKE $%d)", argIdx, argIdx))
		args = append(args, searchTerm)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM quizzes q WHERE %s", whereSQL)
	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("error comptant qüestionaris: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 12
	}
	offset := (page - 1) * pageSize

	selectQuery := fmt.Sprintf(`
		SELECT 
			q.id,
			q.creator_id,
			TRIM(CONCAT(u.first_name, ' ', u.last_name)) AS creator_name,
			q.title,
			q.description,
			q.cover_image_url,
			q.status,
			q.tags,
			(SELECT COUNT(*) FROM quiz_questions qq WHERE qq.quiz_id = q.id) AS question_count,
			q.created_at,
			q.updated_at
		FROM quizzes q
		LEFT JOIN users u ON u.id = q.creator_id
		WHERE %s
		ORDER BY q.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)

	queryArgs := append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("error consultant llistat de qüestionaris: %w", err)
	}
	defer rows.Close()

	items := make([]Quiz, 0)
	for rows.Next() {
		var q Quiz
		var desc sql.NullString
		var cover sql.NullString
		var creatorName sql.NullString

		err := rows.Scan(
			&q.ID,
			&q.CreatorID,
			&creatorName,
			&q.Title,
			&desc,
			&cover,
			&q.Status,
			pq.Array(&q.Tags),
			&q.QuestionCount,
			&q.CreatedAt,
			&q.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error escanejant fila de qüestionari: %w", err)
		}

		if desc.Valid {
			q.Description = &desc.String
		}
		if cover.Valid {
			q.CoverImageURL = &cover.String
		}
		if creatorName.Valid {
			q.CreatorName = creatorName.String
		}
		if q.Tags == nil {
			q.Tags = []string{}
		}

		items = append(items, q)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterant files de qüestionaris: %w", err)
	}

	return items, totalCount, nil
}

// GetQuizByID retrieves the full quiz entity with questions and answers.
func (r *postgresRepository) GetQuizByID(ctx context.Context, id uuid.UUID) (*QuizDetail, error) {
	if r.db == nil {
		return nil, errors.New("connexió a la base de dades no inicialitzada")
	}

	queryQuiz := `
		SELECT 
			q.id,
			q.creator_id,
			TRIM(CONCAT(u.first_name, ' ', u.last_name)) AS creator_name,
			q.title,
			q.description,
			q.cover_image_url,
			q.status,
			q.tags,
			q.created_at,
			q.updated_at
		FROM quizzes q
		LEFT JOIN users u ON u.id = q.creator_id
		WHERE q.id = $1 AND q.deleted_at IS NULL
	`

	var q QuizDetail
	var desc sql.NullString
	var cover sql.NullString
	var creatorName sql.NullString

	err := r.db.QueryRowContext(ctx, queryQuiz, id).Scan(
		&q.ID,
		&q.CreatorID,
		&creatorName,
		&q.Title,
		&desc,
		&cover,
		&q.Status,
		pq.Array(&q.Tags),
		&q.CreatedAt,
		&q.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error consultant qüestionari per ID: %w", err)
	}

	if desc.Valid {
		q.Description = &desc.String
	}
	if cover.Valid {
		q.CoverImageURL = &cover.String
	}
	if creatorName.Valid {
		q.CreatorName = creatorName.String
	}
	if q.Tags == nil {
		q.Tags = []string{}
	}

	// Fetch Questions
	queryQuestions := `
		SELECT 
			id,
			quiz_id,
			text,
			image_url,
			question_type,
			time_limit_seconds,
			order_index,
			created_at,
			updated_at
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY order_index ASC, created_at ASC
	`
	rowsQ, err := r.db.QueryContext(ctx, queryQuestions, id)
	if err != nil {
		return nil, fmt.Errorf("error consultant preguntes del qüestionari: %w", err)
	}
	defer rowsQ.Close()

	questions := make([]QuizQuestion, 0)
	questionIDs := make([]uuid.UUID, 0)
	questionIndexMap := make(map[uuid.UUID]int)

	for rowsQ.Next() {
		var qq QuizQuestion
		var img sql.NullString
		err := rowsQ.Scan(
			&qq.ID,
			&qq.QuizID,
			&qq.Text,
			&img,
			&qq.QuestionType,
			&qq.TimeLimitSeconds,
			&qq.OrderIndex,
			&qq.CreatedAt,
			&qq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escanejant pregunta: %w", err)
		}
		if img.Valid {
			qq.ImageURL = &img.String
		}
		qq.Answers = make([]QuizAnswer, 0)

		questionIndexMap[qq.ID] = len(questions)
		questionIDs = append(questionIDs, qq.ID)
		questions = append(questions, qq)
	}

	if err = rowsQ.Err(); err != nil {
		return nil, fmt.Errorf("error iterant preguntes: %w", err)
	}

	q.QuestionCount = len(questions)

	// Fetch Answers if there are questions
	if len(questionIDs) > 0 {
		queryAnswers := `
			SELECT 
				id,
				question_id,
				text,
				is_correct,
				order_index,
				created_at
			FROM quiz_answers
			WHERE question_id = ANY($1)
			ORDER BY order_index ASC, created_at ASC
		`
		rowsA, err := r.db.QueryContext(ctx, queryAnswers, pq.Array(questionIDs))
		if err != nil {
			return nil, fmt.Errorf("error consultant respostes: %w", err)
		}
		defer rowsA.Close()

		for rowsA.Next() {
			var ans QuizAnswer
			err := rowsA.Scan(
				&ans.ID,
				&ans.QuestionID,
				&ans.Text,
				&ans.IsCorrect,
				&ans.OrderIndex,
				&ans.CreatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("error escanejant resposta: %w", err)
			}

			if idx, ok := questionIndexMap[ans.QuestionID]; ok {
				questions[idx].Answers = append(questions[idx].Answers, ans)
			}
		}

		if err = rowsA.Err(); err != nil {
			return nil, fmt.Errorf("error iterant respostes: %w", err)
		}
	}

	q.Questions = questions
	return &q, nil
}

// CreateQuiz creates a new quiz and its questions/answers atomically.
func (r *postgresRepository) CreateQuiz(ctx context.Context, creatorID uuid.UUID, input CreateQuizInput) (*QuizDetail, error) {
	if r.db == nil {
		return nil, errors.New("connexió a la base de dades no inicialitzada")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error iniciant transacció: %w", err)
	}
	defer tx.Rollback()

	quizID := uuid.New()
	status := input.Status
	if status == "" {
		status = StatusDraft
	}
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}

	insertQuizQuery := `
		INSERT INTO quizzes (
			id, creator_id, title, description, cover_image_url, status, tags, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err = tx.ExecContext(ctx, insertQuizQuery,
		quizID,
		creatorID,
		input.Title,
		input.Description,
		input.CoverImageURL,
		status,
		pq.Array(tags),
	)
	if err != nil {
		return nil, fmt.Errorf("error inserint qüestionari: %w", err)
	}

	for qIdx, qInput := range input.Questions {
		qID := ParseOrGenerateUUID(qInput.ID)
		timeLimit := qInput.TimeLimitSeconds
		if timeLimit == 0 {
			timeLimit = 20
		}
		qType := qInput.QuestionType
		if qType == "" {
			qType = QuestionTypeSingle
		}
		orderIdx := qInput.OrderIndex
		if orderIdx == 0 && qIdx > 0 {
			orderIdx = qIdx
		}

		insertQuestionQuery := `
			INSERT INTO quiz_questions (
				id, quiz_id, text, image_url, question_type, time_limit_seconds, order_index, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`
		_, err = tx.ExecContext(ctx, insertQuestionQuery,
			qID,
			quizID,
			qInput.Text,
			qInput.ImageURL,
			qType,
			timeLimit,
			orderIdx,
		)
		if err != nil {
			return nil, fmt.Errorf("error inserint pregunta: %w", err)
		}

		for aIdx, aInput := range qInput.Answers {
			aID := ParseOrGenerateUUID(aInput.ID)
			ansOrderIdx := aInput.OrderIndex
			if ansOrderIdx == 0 && aIdx > 0 {
				ansOrderIdx = aIdx
			}

			insertAnswerQuery := `
				INSERT INTO quiz_answers (
					id, question_id, text, is_correct, order_index, created_at
				) VALUES ($1, $2, $3, $4, $5, NOW())
			`
			_, err = tx.ExecContext(ctx, insertAnswerQuery,
				aID,
				qID,
				aInput.Text,
				aInput.IsCorrect,
				ansOrderIdx,
			)
			if err != nil {
				return nil, fmt.Errorf("error inserint resposta: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error confirmant transacció de creació: %w", err)
	}

	return r.GetQuizByID(ctx, quizID)
}

// UpdateQuiz updates metadata and replaces questions and answers atomically in a transaction.
func (r *postgresRepository) UpdateQuiz(ctx context.Context, id uuid.UUID, input UpdateQuizInput) (*QuizDetail, error) {
	if r.db == nil {
		return nil, errors.New("connexió a la base de dades no inicialitzada")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error iniciant transacció: %w", err)
	}
	defer tx.Rollback()

	// Update quiz
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}

	updateQuizQuery := `
		UPDATE quizzes
		SET 
			title = $1,
			description = $2,
			cover_image_url = $3,
			status = COALESCE($4, status),
			tags = $5,
			updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`
	res, err := tx.ExecContext(ctx, updateQuizQuery,
		input.Title,
		input.Description,
		input.CoverImageURL,
		input.Status,
		pq.Array(tags),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("error actualitzant dades del qüestionari: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	if input.Questions != nil {
		// Delete old questions (cascade deletes answers)
		deleteQuestionsQuery := `DELETE FROM quiz_questions WHERE quiz_id = $1`
		if _, err := tx.ExecContext(ctx, deleteQuestionsQuery, id); err != nil {
			return nil, fmt.Errorf("error eliminant preguntes anteriors: %w", err)
		}

		// Insert updated questions and answers
		for qIdx, qInput := range input.Questions {
			qID := ParseOrGenerateUUID(qInput.ID)
			timeLimit := qInput.TimeLimitSeconds
			if timeLimit == 0 {
				timeLimit = 20
			}
			qType := qInput.QuestionType
			if qType == "" {
				qType = QuestionTypeSingle
			}
			orderIdx := qInput.OrderIndex
			if orderIdx == 0 && qIdx > 0 {
				orderIdx = qIdx
			}

			insertQuestionQuery := `
				INSERT INTO quiz_questions (
					id, quiz_id, text, image_url, question_type, time_limit_seconds, order_index, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
			`
			_, err = tx.ExecContext(ctx, insertQuestionQuery,
				qID,
				id,
				qInput.Text,
				qInput.ImageURL,
				qType,
				timeLimit,
				orderIdx,
			)
			if err != nil {
				return nil, fmt.Errorf("error re-inserint pregunta: %w", err)
			}

			for aIdx, aInput := range qInput.Answers {
				aID := ParseOrGenerateUUID(aInput.ID)
				ansOrderIdx := aInput.OrderIndex
				if ansOrderIdx == 0 && aIdx > 0 {
					ansOrderIdx = aIdx
				}

				insertAnswerQuery := `
					INSERT INTO quiz_answers (
						id, question_id, text, is_correct, order_index, created_at
					) VALUES ($1, $2, $3, $4, $5, NOW())
				`
				_, err = tx.ExecContext(ctx, insertAnswerQuery,
					aID,
					qID,
					aInput.Text,
					aInput.IsCorrect,
					ansOrderIdx,
				)
				if err != nil {
					return nil, fmt.Errorf("error re-inserint resposta: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error confirmant transacció d'actualització: %w", err)
	}

	return r.GetQuizByID(ctx, id)
}

// DeleteQuiz marks a quiz as soft-deleted.
func (r *postgresRepository) DeleteQuiz(ctx context.Context, id uuid.UUID) error {
	if r.db == nil {
		return errors.New("connexió a la base de dades no inicialitzada")
	}

	query := `UPDATE quizzes SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error marcant qüestionari com a eliminat: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error verificant files afectades: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DuplicateQuiz creates a copy of a quiz and its questions (and optionally answers).
func (r *postgresRepository) DuplicateQuiz(ctx context.Context, originalID, newCreatorID uuid.UUID, newTitle string, includeAnswers bool) (*QuizDetail, error) {
	original, err := r.GetQuizByID(ctx, originalID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, sql.ErrNoRows
	}

	createInput := CreateQuizInput{
		Title:         newTitle,
		Description:   original.Description,
		CoverImageURL: original.CoverImageURL,
		Status:        StatusDraft,
		Tags:          original.Tags,
		Questions:     make([]SaveQuestionInput, 0, len(original.Questions)),
	}

	for _, q := range original.Questions {
		saveQ := SaveQuestionInput{
			Text:             q.Text,
			ImageURL:         q.ImageURL,
			QuestionType:     q.QuestionType,
			TimeLimitSeconds: q.TimeLimitSeconds,
			OrderIndex:       q.OrderIndex,
			Answers:          make([]SaveAnswerInput, 0),
		}

		if includeAnswers {
			for _, a := range q.Answers {
				saveQ.Answers = append(saveQ.Answers, SaveAnswerInput{
					Text:       a.Text,
					IsCorrect:  a.IsCorrect,
					OrderIndex: a.OrderIndex,
				})
			}
		}

		createInput.Questions = append(createInput.Questions, saveQ)
	}

	return r.CreateQuiz(ctx, newCreatorID, createInput)
}
