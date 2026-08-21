package evaluation

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/lib/pq"
)

type Repository interface {
	ListEvaluations(teacherID string, isAdmin bool) ([]EvaluationQuizSummary, error)
	GetQuizEvaluation(quizID string) (*QuizEvaluationResponse, error)
	GetStudentEvaluation(quizID, studentID string) (*StudentEvaluationDetail, error)
	GradeStudent(quizID, studentID, teacherID string, finalGrade float64) (*GradeResponse, error)
	UpsertCalculatedGradeForMatch(matchID string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListEvaluations(teacherID string, isAdmin bool) ([]EvaluationQuizSummary, error) {
	query := `
		SELECT 
			q.id AS quiz_id,
			q.title AS quiz_title,
			COUNT(DISTINCT m.id) AS total_matches,
			COUNT(DISTINCT mp.user_id) AS total_students,
			COUNT(DISTINCT CASE WHEN e.is_graded = true THEN e.student_id END) AS graded_count,
			MAX(m.updated_at) AS last_match_at
		FROM quizzes q
		INNER JOIN matches m ON m.quiz_id = q.id AND m.status = 'finished' AND m.deleted_at IS NULL
		INNER JOIN match_players mp ON mp.match_id = m.id
		LEFT JOIN evaluations e ON e.quiz_id = q.id AND e.student_id = mp.user_id
		WHERE q.deleted_at IS NULL
	`
	args := []interface{}{}
	if !isAdmin {
		query += ` AND q.creator_id = $1`
		args = append(args, teacherID)
	}
	query += ` GROUP BY q.id, q.title ORDER BY last_match_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying evaluations list: %w", err)
	}
	defer rows.Close()

	summaries := []EvaluationQuizSummary{}
	for rows.Next() {
		var s EvaluationQuizSummary
		if err := rows.Scan(&s.QuizID, &s.QuizTitle, &s.TotalMatches, &s.TotalStudents, &s.GradedCount, &s.LastMatchAt); err != nil {
			return nil, fmt.Errorf("error scanning evaluation summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func (r *repository) GetQuizEvaluation(quizID string) (*QuizEvaluationResponse, error) {
	var resp QuizEvaluationResponse
	err := r.db.QueryRow(`
		SELECT q.id, q.title, COUNT(DISTINCT m.id)
		FROM quizzes q
		LEFT JOIN matches m ON m.quiz_id = q.id AND m.status = 'finished' AND m.deleted_at IS NULL
		WHERE q.id = $1 AND q.deleted_at IS NULL
		GROUP BY q.id, q.title
	`, quizID).Scan(&resp.QuizID, &resp.QuizTitle, &resp.TotalMatches)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("quiz not found")
		}
		return nil, err
	}

	// Question Stats
	qRows, err := r.db.Query(`
		SELECT id, text, order_index
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY order_index ASC
	`, quizID)
	if err != nil {
		return nil, fmt.Errorf("error querying quiz questions: %w", err)
	}
	defer qRows.Close()

	stats := []QuestionStats{}
	for qRows.Next() {
		var qid, qtext string
		var qidx int
		if err := qRows.Scan(&qid, &qtext, &qidx); err != nil {
			return nil, err
		}

		// Calculate stats for this question across all finished matches of this quiz
		var totalAnswers, correctAnswers, sumTimeMs, noAnswerCount int
		err := r.db.QueryRow(`
			SELECT 
				COUNT(ma.id) AS total_answers,
				COUNT(CASE WHEN ma.is_correct = true THEN 1 END) AS correct_answers,
				COALESCE(SUM(ma.response_time_ms), 0) AS sum_time,
				(
					SELECT COUNT(DISTINCT mp.id)
					FROM match_players mp
					INNER JOIN matches m ON m.id = mp.match_id
					WHERE m.quiz_id = $1 AND m.status = 'finished' AND m.deleted_at IS NULL
					  AND mp.id NOT IN (
						SELECT player_id FROM match_answers WHERE question_id = $2
					  )
				) AS no_answer_count
			FROM match_answers ma
			INNER JOIN matches m ON m.id = ma.match_id
			WHERE m.quiz_id = $1 AND m.status = 'finished' AND m.deleted_at IS NULL
			  AND ma.question_id = $2
		`, quizID, qid).Scan(&totalAnswers, &correctAnswers, &sumTimeMs, &noAnswerCount)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		hitRate := 0.0
		if totalAnswers > 0 {
			hitRate = math.Round((float64(correctAnswers)/float64(totalAnswers))*100) / 100
		}
		avgTimeMs := 0
		if totalAnswers > 0 {
			avgTimeMs = sumTimeMs / totalAnswers
		}

		// Options distribution
		optRows, err := r.db.Query(`
			SELECT id, text, is_correct
			FROM quiz_answers
			WHERE question_id = $1
			ORDER BY order_index ASC
		`, qid)
		if err != nil {
			return nil, err
		}

		distItems := []AnswerDistributionItem{}
		for optRows.Next() {
			var optId, optText string
			var isCorr bool
			if err := optRows.Scan(&optId, &optText, &isCorr); err != nil {
				optRows.Close()
				return nil, err
			}

			var count int
			_ = r.db.QueryRow(`
				SELECT COUNT(*)
				FROM match_answers ma
				INNER JOIN matches m ON m.id = ma.match_id
				WHERE m.quiz_id = $1 AND m.status = 'finished' AND m.deleted_at IS NULL
				  AND ma.question_id = $2
				  AND $3 = ANY(ma.selected_answer_ids)
			`, quizID, qid, optId).Scan(&count)

			pct := 0.0
			if totalAnswers > 0 {
				pct = math.Round((float64(count)/float64(totalAnswers))*100) / 100
			}

			distItems = append(distItems, AnswerDistributionItem{
				AnswerID:   optId,
				AnswerText: optText,
				IsCorrect:  isCorr,
				Count:      count,
				Percentage: pct,
			})
		}
		optRows.Close()

		stats = append(stats, QuestionStats{
			QuestionID:         qid,
			QuestionText:       qtext,
			QuestionIndex:      qidx,
			HitRate:            hitRate,
			AvgResponseTimeMs:  avgTimeMs,
			AnswerDistribution: distItems,
			NoAnswerCount:      noAnswerCount,
		})
	}
	resp.Stats = stats

	// Students list
	stRows, err := r.db.Query(`
		SELECT 
			u.id AS student_id,
			u.name AS student_name,
			COUNT(DISTINCT m.id) AS matches_count,
			COALESCE(e.calculated_grade, 0.00) AS calculated_grade,
			e.final_grade,
			COALESCE(e.is_graded, false) AS is_graded
		FROM users u
		INNER JOIN match_players mp ON mp.user_id = u.id
		INNER JOIN matches m ON m.id = mp.match_id AND m.status = 'finished' AND m.deleted_at IS NULL
		LEFT JOIN evaluations e ON e.quiz_id = m.quiz_id AND e.student_id = u.id
		WHERE m.quiz_id = $1
		GROUP BY u.id, u.name, e.calculated_grade, e.final_grade, e.is_graded
		ORDER BY u.name ASC
	`, quizID)
	if err != nil {
		return nil, fmt.Errorf("error querying students list: %w", err)
	}
	defer stRows.Close()

	students := []StudentEvaluationSummary{}
	for stRows.Next() {
		var s StudentEvaluationSummary
		var finalGrade sql.NullFloat64
		if err := stRows.Scan(&s.StudentID, &s.StudentName, &s.MatchesCount, &s.CalculatedGrade, &finalGrade, &s.IsGraded); err != nil {
			return nil, err
		}
		if finalGrade.Valid {
			fg := finalGrade.Float64
			s.FinalGrade = &fg
		}
		students = append(students, s)
	}
	resp.Students = students

	return &resp, nil
}

func (r *repository) GetStudentEvaluation(quizID, studentID string) (*StudentEvaluationDetail, error) {
	var detail StudentEvaluationDetail
	detail.StudentID = studentID

	err := r.db.QueryRow("SELECT name FROM users WHERE id = $1", studentID).Scan(&detail.StudentName)
	if err != nil {
		return nil, fmt.Errorf("student not found")
	}

	var evalID sql.NullString
	var calcGrade sql.NullFloat64
	var finalGrade sql.NullFloat64
	var isGraded sql.NullBool
	var gradedBy sql.NullString
	var gradedAt pq.NullTime

	err = r.db.QueryRow(`
		SELECT id, calculated_grade, final_grade, is_graded, graded_by, graded_at
		FROM evaluations
		WHERE quiz_id = $1 AND student_id = $2
	`, quizID, studentID).Scan(&evalID, &calcGrade, &finalGrade, &isGraded, &gradedBy, &gradedAt)
	if err == nil && evalID.Valid {
		detail.EvaluationID = evalID.String
		detail.CalculatedGrade = calcGrade.Float64
		if finalGrade.Valid {
			fg := finalGrade.Float64
			detail.FinalGrade = &fg
		}
		detail.IsGraded = isGraded.Bool
		if gradedBy.Valid {
			gb := gradedBy.String
			detail.GradedBy = &gb
		}
		if gradedAt.Valid {
			gt := gradedAt.Time
			detail.GradedAt = &gt
		}
	} else {
		detail.EvaluationID = ""
		detail.CalculatedGrade = 0.00
		detail.IsGraded = false
	}

	// Fetch matches
	mRows, err := r.db.Query(`
		SELECT m.id, m.created_at, mp.score,
		       (SELECT COUNT(*) FROM quiz_questions WHERE quiz_id = $1) AS total_questions
		FROM matches m
		INNER JOIN match_players mp ON mp.match_id = m.id
		WHERE m.quiz_id = $1 AND mp.user_id = $2 AND m.status = 'finished' AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
	`, quizID, studentID)
	if err != nil {
		return nil, err
	}
	defer mRows.Close()

	matchResults := []StudentMatchResult{}
	for mRows.Next() {
		var mr StudentMatchResult
		if err := mRows.Scan(&mr.MatchID, &mr.MatchDate, &mr.Score, &mr.TotalQuestions); err != nil {
			return nil, err
		}

		// Fetch answers for this match & questions
		qRows, err := r.db.Query(`
			SELECT qq.id, qq.text, qq.order_index,
			       COALESCE(ma.selected_answer_ids, '{}') AS selected,
				   COALESCE(ma.is_correct, false) AS is_correct,
				   COALESCE(ma.response_time_ms, 0) AS response_time
			FROM quiz_questions qq
			INNER JOIN match_players mp ON mp.match_id = $1 AND mp.user_id = $2
			LEFT JOIN match_answers ma ON ma.match_id = $1 AND ma.question_id = qq.id AND ma.player_id = mp.id
			WHERE qq.quiz_id = $3
			ORDER BY qq.order_index ASC
		`, mr.MatchID, studentID, quizID)
		if err != nil {
			return nil, err
		}

		answers := []StudentAnswerDetail{}
		for qRows.Next() {
			var ad StudentAnswerDetail
			var selectedArray []string
			if err := qRows.Scan(&ad.QuestionID, &ad.QuestionText, &ad.QuestionIndex, pq.Array(&selectedArray), &ad.IsCorrect, &ad.ResponseTimeMs); err != nil {
				qRows.Close()
				return nil, err
			}
			ad.SelectedAnswerIDs = selectedArray

			// Get correct answers
			caRows, _ := r.db.Query("SELECT id FROM quiz_answers WHERE question_id = $1 AND is_correct = true", ad.QuestionID)
			caIDs := []string{}
			if caRows != nil {
				for caRows.Next() {
					var caID string
					_ = caRows.Scan(&caID)
					caIDs = append(caIDs, caID)
				}
				caRows.Close()
			}
			ad.CorrectAnswerIDs = caIDs
			answers = append(answers, ad)
		}
		qRows.Close()
		mr.Answers = answers
		matchResults = append(matchResults, mr)
	}

	detail.Matches = matchResults
	return &detail, nil
}

func (r *repository) GradeStudent(quizID, studentID, teacherID string, finalGrade float64) (*GradeResponse, error) {
	now := time.Now()
	var evalID string
	var calcGrade float64

	err := r.db.QueryRow(`
		INSERT INTO evaluations (quiz_id, student_id, calculated_grade, final_grade, is_graded, graded_by, graded_at, created_at, updated_at)
		VALUES ($1, $2, 0.00, $3, true, $4, $5, $5, $5)
		ON CONFLICT (quiz_id, student_id)
		DO UPDATE SET
			final_grade = EXCLUDED.final_grade,
			is_graded = true,
			graded_by = EXCLUDED.graded_by,
			graded_at = EXCLUDED.graded_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, calculated_grade
	`, quizID, studentID, finalGrade, teacherID, now).Scan(&evalID, &calcGrade)
	if err != nil {
		return nil, fmt.Errorf("error updating student grade: %w", err)
	}

	return &GradeResponse{
		EvaluationID:    evalID,
		CalculatedGrade: calcGrade,
		FinalGrade:      finalGrade,
		IsGraded:        true,
		GradedBy:        teacherID,
		GradedAt:        now,
	}, nil
}

func (r *repository) UpsertCalculatedGradeForMatch(matchID string) error {
	var quizID string
	err := r.db.QueryRow("SELECT quiz_id FROM matches WHERE id = $1", matchID).Scan(&quizID)
	if err != nil {
		return err
	}

	var totalQuestions int
	err = r.db.QueryRow("SELECT COUNT(*) FROM quiz_questions WHERE quiz_id = $1", quizID).Scan(&totalQuestions)
	if err != nil || totalQuestions == 0 {
		return nil
	}

	rows, err := r.db.Query(`
		SELECT mp.user_id, mp.score
		FROM match_players mp
		WHERE mp.match_id = $1
	`, matchID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		var score int
		if err := rows.Scan(&userID, &score); err != nil {
			continue
		}

		grade := math.Round((float64(score)/float64(totalQuestions))*1000) / 100
		if grade > 10.0 {
			grade = 10.0
		}

		now := time.Now()
		_, _ = r.db.Exec(`
			INSERT INTO evaluations (quiz_id, student_id, calculated_grade, is_graded, created_at, updated_at)
			VALUES ($1, $2, $3, false, $4, $4)
			ON CONFLICT (quiz_id, student_id)
			DO UPDATE SET
				calculated_grade = EXCLUDED.calculated_grade,
				updated_at = EXCLUDED.updated_at
		`, quizID, userID, grade, now)
	}

	return nil
}
