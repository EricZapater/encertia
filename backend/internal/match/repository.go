package match

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/encertia/backend/internal/quiz"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Repository defines all database operations for the match module.
type Repository interface {
	GenerateUniquePIN(ctx context.Context) (string, error)
	CreateMatch(ctx context.Context, match *Match) error
	GetMatchByID(ctx context.Context, id uuid.UUID) (*Match, error)
	GetMatchByPIN(ctx context.Context, pin string) (*Match, error)
	GetMatchWithQuiz(ctx context.Context, pin string) (*Match, *quiz.QuizDetail, error)
	GetMatchWithQuizByID(ctx context.Context, id uuid.UUID) (*Match, *quiz.QuizDetail, error)
	UpdateMatchStatus(ctx context.Context, matchID uuid.UUID, status MatchStatus, currentQuestionIndex int, startedAt *time.Time) error
	AddOrUpdatePlayer(ctx context.Context, player *MatchPlayer) error
	GetPlayerByMatchAndUser(ctx context.Context, matchID, userID uuid.UUID) (*MatchPlayer, error)
	GetPlayerByID(ctx context.Context, playerID uuid.UUID) (*MatchPlayer, error)
	GetPlayersByMatch(ctx context.Context, matchID uuid.UUID) ([]MatchPlayer, error)
	UpdatePlayerConnection(ctx context.Context, playerID uuid.UUID, isConnected bool) error
	KickPlayer(ctx context.Context, matchID, playerID uuid.UUID) error
	RecordAnswer(ctx context.Context, answer *MatchAnswer) error
	GetAnswersForQuestion(ctx context.Context, matchID, questionID uuid.UUID) ([]MatchAnswer, error)
	GetLeaderboard(ctx context.Context, matchID uuid.UUID) ([]PlayerScoreItem, error)
	GetMatchSummary(ctx context.Context, matchID uuid.UUID) (*MatchSummaryResponse, error)
	GetPublicInfoByPIN(ctx context.Context, pin string) (*MatchPublicInfo, error)
}

type sqlRepository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL implementation of Repository.
func NewRepository(db *sql.DB) Repository {
	return &sqlRepository{db: db}
}

// GenerateUniquePIN creates a random 6-digit PIN not currently in use by an active match.
func (r *sqlRepository) GenerateUniquePIN(ctx context.Context) (string, error) {
	const maxAttempts = 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		n, err := rand.Int(rand.Reader, big.NewInt(1000000))
		if err != nil {
			return "", fmt.Errorf("error generant pin aleatori: %w", err)
		}
		pin := fmt.Sprintf("%06d", n.Int64())

		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM matches WHERE pin = $1 AND status != 'finished' AND deleted_at IS NULL)`
		if err := r.db.QueryRowContext(ctx, query, pin).Scan(&exists); err != nil {
			return "", fmt.Errorf("error verificant unicitat de pin: %w", err)
		}

		if !exists {
			return pin, nil
		}
	}
	return "", errors.New("no s'ha pogut generar un PIN únic després de múltiples intents")
}

// CreateMatch inserts a new match record into the database.
func (r *sqlRepository) CreateMatch(ctx context.Context, m *Match) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.Status == "" {
		m.Status = StatusLobby
	}

	query := `
		INSERT INTO matches (id, quiz_id, host_id, pin, status, current_question_index, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.QuizID, m.HostID, m.PIN, m.Status, m.CurrentQuestionIndex, m.CreatedAt, m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error creant partida a la bd: %w", err)
	}
	return nil
}

// GetMatchByID fetches a match by its primary key.
func (r *sqlRepository) GetMatchByID(ctx context.Context, id uuid.UUID) (*Match, error) {
	query := `
		SELECT m.id, m.quiz_id, m.host_id, m.pin, m.status, m.current_question_index, m.question_started_at,
		       m.created_at, m.updated_at, m.deleted_at, q.title, TRIM(CONCAT(u.first_name, ' ', u.last_name)),
		       (SELECT COUNT(*) FROM match_players mp WHERE mp.match_id = m.id AND mp.is_kicked = FALSE) as player_count
		FROM matches m
		JOIN quizzes q ON m.quiz_id = q.id
		JOIN users u ON m.host_id = u.id
		WHERE m.id = $1 AND m.deleted_at IS NULL
	`
	var m Match
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.QuizID, &m.HostID, &m.PIN, &m.Status, &m.CurrentQuestionIndex, &m.QuestionStartedAt,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.QuizTitle, &m.HostName, &m.PlayerCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error cercant partida per id: %w", err)
	}
	return &m, nil
}

// GetMatchByPIN fetches an active match by its 6-digit PIN.
func (r *sqlRepository) GetMatchByPIN(ctx context.Context, pin string) (*Match, error) {
	query := `
		SELECT m.id, m.quiz_id, m.host_id, m.pin, m.status, m.current_question_index, m.question_started_at,
		       m.created_at, m.updated_at, m.deleted_at, q.title, TRIM(CONCAT(u.first_name, ' ', u.last_name)),
		       (SELECT COUNT(*) FROM match_players mp WHERE mp.match_id = m.id AND mp.is_kicked = FALSE) as player_count
		FROM matches m
		JOIN quizzes q ON m.quiz_id = q.id
		JOIN users u ON m.host_id = u.id
		WHERE m.pin = $1 AND m.status != 'finished' AND m.deleted_at IS NULL
	`
	var m Match
	err := r.db.QueryRowContext(ctx, query, pin).Scan(
		&m.ID, &m.QuizID, &m.HostID, &m.PIN, &m.Status, &m.CurrentQuestionIndex, &m.QuestionStartedAt,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.QuizTitle, &m.HostName, &m.PlayerCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error cercant partida per pin: %w", err)
	}
	return &m, nil
}

// GetMatchWithQuiz loads the match along with its full quiz detail (questions and answers).
func (r *sqlRepository) GetMatchWithQuiz(ctx context.Context, pin string) (*Match, *quiz.QuizDetail, error) {
	m, err := r.GetMatchByPIN(ctx, pin)
	if err != nil {
		return nil, nil, err
	}
	if m == nil {
		return nil, nil, nil
	}
	return r.loadQuizDetailForMatch(ctx, m)
}

// GetMatchWithQuizByID loads the match and quiz detail by match UUID.
func (r *sqlRepository) GetMatchWithQuizByID(ctx context.Context, id uuid.UUID) (*Match, *quiz.QuizDetail, error) {
	m, err := r.GetMatchByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if m == nil {
		return nil, nil, nil
	}
	return r.loadQuizDetailForMatch(ctx, m)
}

func (r *sqlRepository) loadQuizDetailForMatch(ctx context.Context, m *Match) (*Match, *quiz.QuizDetail, error) {
	// 1. Fetch Quiz metadata
	quizQuery := `
		SELECT q.id, q.creator_id, TRIM(CONCAT(u.first_name, ' ', u.last_name)), q.title, q.description, q.cover_image_url, q.status, q.tags,
		       (SELECT COUNT(*) FROM quiz_questions qq WHERE qq.quiz_id = q.id) as question_count,
		       q.created_at, q.updated_at
		FROM quizzes q
		JOIN users u ON q.creator_id = u.id
		WHERE q.id = $1 AND q.deleted_at IS NULL
	`
	var qDetail quiz.QuizDetail
	var tags pq.StringArray
	err := r.db.QueryRowContext(ctx, quizQuery, m.QuizID).Scan(
		&qDetail.ID, &qDetail.CreatorID, &qDetail.CreatorName, &qDetail.Title, &qDetail.Description,
		&qDetail.CoverImageURL, &qDetail.Status, &tags, &qDetail.QuestionCount, &qDetail.CreatedAt, &qDetail.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error carregant qüestionari de la partida: %w", err)
	}
	qDetail.Tags = []string(tags)
	m.QuizTitle = qDetail.Title

	// 2. Fetch Questions
	questionsQuery := `
		SELECT id, quiz_id, text, image_url, question_type, time_limit_seconds, order_index, created_at, updated_at
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY order_index ASC
	`
	qRows, err := r.db.QueryContext(ctx, questionsQuery, m.QuizID)
	if err != nil {
		return nil, nil, fmt.Errorf("error carregant preguntes del qüestionari: %w", err)
	}
	defer qRows.Close()

	var questions []quiz.QuizQuestion
	var questionIDs []uuid.UUID

	for qRows.Next() {
		var qq quiz.QuizQuestion
		if err := qRows.Scan(
			&qq.ID, &qq.QuizID, &qq.Text, &qq.ImageURL, &qq.QuestionType, &qq.TimeLimitSeconds,
			&qq.OrderIndex, &qq.CreatedAt, &qq.UpdatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("error llegint fila de pregunta: %w", err)
		}
		questions = append(questions, qq)
		questionIDs = append(questionIDs, qq.ID)
	}
	if err := qRows.Err(); err != nil {
		return nil, nil, err
	}

	// 3. Fetch Answers for questions
	if len(questionIDs) > 0 {
		answersQuery := `
			SELECT id, question_id, text, is_correct, order_index, created_at
			FROM quiz_answers
			WHERE question_id = ANY($1)
			ORDER BY order_index ASC
		`
		aRows, err := r.db.QueryContext(ctx, answersQuery, pq.Array(questionIDs))
		if err != nil {
			return nil, nil, fmt.Errorf("error carregant opcions de resposta: %w", err)
		}
		defer aRows.Close()

		answersMap := make(map[uuid.UUID][]quiz.QuizAnswer)
		for aRows.Next() {
			var ans quiz.QuizAnswer
			if err := aRows.Scan(&ans.ID, &ans.QuestionID, &ans.Text, &ans.IsCorrect, &ans.OrderIndex, &ans.CreatedAt); err != nil {
				return nil, nil, fmt.Errorf("error llegint fila de resposta: %w", err)
			}
			answersMap[ans.QuestionID] = append(answersMap[ans.QuestionID], ans)
		}

		for i := range questions {
			questions[i].Answers = answersMap[questions[i].ID]
		}
	}

	qDetail.Questions = questions
	return m, &qDetail, nil
}

// UpdateMatchStatus modifies status, question index and timer of the match.
func (r *sqlRepository) UpdateMatchStatus(ctx context.Context, matchID uuid.UUID, status MatchStatus, currentQuestionIndex int, startedAt *time.Time) error {
	query := `
		UPDATE matches
		SET status = $1, current_question_index = $2, question_started_at = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, status, currentQuestionIndex, startedAt, matchID)
	if err != nil {
		return fmt.Errorf("error actualitzant estat de la partida: %w", err)
	}
	return nil
}

// AddOrUpdatePlayer registers or updates a player in the match.
func (r *sqlRepository) AddOrUpdatePlayer(ctx context.Context, p *MatchPlayer) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now().UTC()
	p.JoinedAt = now
	p.UpdatedAt = now

	query := `
		INSERT INTO match_players (id, match_id, user_id, nickname, score, is_connected, is_kicked, joined_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (match_id, user_id) DO UPDATE
		SET nickname = EXCLUDED.nickname, is_connected = EXCLUDED.is_connected, updated_at = NOW()
		RETURNING id, score, is_kicked, joined_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, p.ID, p.MatchID, p.UserID, p.Nickname, p.Score, p.IsConnected, p.IsKicked, p.JoinedAt, p.UpdatedAt).
		Scan(&p.ID, &p.Score, &p.IsKicked, &p.JoinedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("error registrant/actualitzant jugador: %w", err)
	}
	return nil
}

// GetPlayerByMatchAndUser gets a player record by match ID and user ID.
func (r *sqlRepository) GetPlayerByMatchAndUser(ctx context.Context, matchID, userID uuid.UUID) (*MatchPlayer, error) {
	query := `
		SELECT id, match_id, user_id, nickname, score, is_connected, is_kicked, joined_at, updated_at
		FROM match_players
		WHERE match_id = $1 AND user_id = $2
	`
	var p MatchPlayer
	err := r.db.QueryRowContext(ctx, query, matchID, userID).Scan(
		&p.ID, &p.MatchID, &p.UserID, &p.Nickname, &p.Score, &p.IsConnected, &p.IsKicked, &p.JoinedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error cercant jugador per usuari: %w", err)
	}
	return &p, nil
}

// GetPlayerByID fetches a player record by its ID.
func (r *sqlRepository) GetPlayerByID(ctx context.Context, playerID uuid.UUID) (*MatchPlayer, error) {
	query := `
		SELECT id, match_id, user_id, nickname, score, is_connected, is_kicked, joined_at, updated_at
		FROM match_players
		WHERE id = $1
	`
	var p MatchPlayer
	err := r.db.QueryRowContext(ctx, query, playerID).Scan(
		&p.ID, &p.MatchID, &p.UserID, &p.Nickname, &p.Score, &p.IsConnected, &p.IsKicked, &p.JoinedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error cercant jugador per id: %w", err)
	}
	return &p, nil
}

// GetPlayersByMatch fetches all non-kicked players for a match.
func (r *sqlRepository) GetPlayersByMatch(ctx context.Context, matchID uuid.UUID) ([]MatchPlayer, error) {
	query := `
		SELECT id, match_id, user_id, nickname, score, is_connected, is_kicked, joined_at, updated_at
		FROM match_players
		WHERE match_id = $1 AND is_kicked = FALSE
		ORDER BY joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("error cercant jugadors de la partida: %w", err)
	}
	defer rows.Close()

	var players []MatchPlayer
	for rows.Next() {
		var p MatchPlayer
		if err := rows.Scan(&p.ID, &p.MatchID, &p.UserID, &p.Nickname, &p.Score, &p.IsConnected, &p.IsKicked, &p.JoinedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error llegint fila de jugador: %w", err)
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// UpdatePlayerConnection updates the is_connected status of a player.
func (r *sqlRepository) UpdatePlayerConnection(ctx context.Context, playerID uuid.UUID, isConnected bool) error {
	query := `UPDATE match_players SET is_connected = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, isConnected, playerID)
	if err != nil {
		return fmt.Errorf("error actualitzant connexió del jugador: %w", err)
	}
	return nil
}

// KickPlayer marks a player as kicked.
func (r *sqlRepository) KickPlayer(ctx context.Context, matchID, playerID uuid.UUID) error {
	query := `UPDATE match_players SET is_kicked = TRUE, is_connected = FALSE, updated_at = NOW() WHERE id = $1 AND match_id = $2`
	res, err := r.db.ExecContext(ctx, query, playerID, matchID)
	if err != nil {
		return fmt.Errorf("error expulsant jugador: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("jugador no trobat a la partida")
	}
	return nil
}

// RecordAnswer inserts a player's answer and updates player's cumulative score atomically.
func (r *sqlRepository) RecordAnswer(ctx context.Context, a *MatchAnswer) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	a.AnsweredAt = time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error iniciant transacció de resposta: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	insertQuery := `
		INSERT INTO match_answers (id, match_id, question_id, player_id, selected_answer_ids, is_correct, score_awarded, response_time_ms, answered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.ExecContext(ctx, insertQuery, a.ID, a.MatchID, a.QuestionID, a.PlayerID, pq.Array(a.SelectedAnswerIDs), a.IsCorrect, a.ScoreAwarded, a.ResponseTimeMS, a.AnsweredAt)
	if err != nil {
		return fmt.Errorf("error registrant resposta a la bd: %w", err)
	}

	if a.ScoreAwarded > 0 {
		updateScoreQuery := `
			UPDATE match_players
			SET score = score + $1, updated_at = NOW()
			WHERE id = $2
		`
		_, err = tx.ExecContext(ctx, updateScoreQuery, a.ScoreAwarded, a.PlayerID)
		if err != nil {
			return fmt.Errorf("error actualitzant puntuació de jugador: %w", err)
		}
	}

	return tx.Commit()
}

// GetAnswersForQuestion retrieves all answers given for a specific question.
func (r *sqlRepository) GetAnswersForQuestion(ctx context.Context, matchID, questionID uuid.UUID) ([]MatchAnswer, error) {
	query := `
		SELECT id, match_id, question_id, player_id, selected_answer_ids, is_correct, score_awarded, response_time_ms, answered_at
		FROM match_answers
		WHERE match_id = $1 AND question_id = $2
	`
	rows, err := r.db.QueryContext(ctx, query, matchID, questionID)
	if err != nil {
		return nil, fmt.Errorf("error obtenint respostes per la pregunta: %w", err)
	}
	defer rows.Close()

	var answers []MatchAnswer
	for rows.Next() {
		var a MatchAnswer
		var ids []string
		if err := rows.Scan(&a.ID, &a.MatchID, &a.QuestionID, &a.PlayerID, pq.Array(&ids), &a.IsCorrect, &a.ScoreAwarded, &a.ResponseTimeMS, &a.AnsweredAt); err != nil {
			return nil, fmt.Errorf("error llegint resposta: %w", err)
		}
		for _, idStr := range ids {
			if parsed, err := uuid.Parse(idStr); err == nil {
				a.SelectedAnswerIDs = append(a.SelectedAnswerIDs, parsed)
			}
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

// GetLeaderboard calculates ranks, correct counts and total answers for all active players in a match.
func (r *sqlRepository) GetLeaderboard(ctx context.Context, matchID uuid.UUID) ([]PlayerScoreItem, error) {
	query := `
		SELECT mp.id, mp.user_id, mp.nickname, mp.score,
		       DENSE_RANK() OVER (ORDER BY mp.score DESC, mp.joined_at ASC) as rank,
		       COALESCE(SUM(CASE WHEN ma.is_correct = TRUE THEN 1 ELSE 0 END), 0) as correct_count,
		       COALESCE(COUNT(ma.id), 0) as total_answered
		FROM match_players mp
		LEFT JOIN match_answers ma ON mp.id = ma.player_id AND ma.match_id = mp.match_id
		WHERE mp.match_id = $1 AND mp.is_kicked = FALSE
		GROUP BY mp.id, mp.user_id, mp.nickname, mp.score, mp.joined_at
		ORDER BY rank ASC, mp.joined_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("error calculant classificació: %w", err)
	}
	defer rows.Close()

	var items []PlayerScoreItem
	for rows.Next() {
		var item PlayerScoreItem
		if err := rows.Scan(&item.PlayerID, &item.UserID, &item.Nickname, &item.Score, &item.Rank, &item.CorrectCount, &item.TotalAnswered); err != nil {
			return nil, fmt.Errorf("error llegint classificació: %w", err)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []PlayerScoreItem{}
	}
	return items, rows.Err()
}

// GetMatchSummary returns the full podium, complete leaderboard and stats.
func (r *sqlRepository) GetMatchSummary(ctx context.Context, matchID uuid.UUID) (*MatchSummaryResponse, error) {
	// 1. Fetch match and quiz info
	m, qDetail, err := r.GetMatchWithQuizByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if m == nil || qDetail == nil {
		return nil, nil
	}

	// 2. Fetch Leaderboard
	leaderboard, err := r.GetLeaderboard(ctx, matchID)
	if err != nil {
		return nil, err
	}

	// 3. Compute Podium (Top 3)
	podium := make([]PlayerScoreItem, 0, 3)
	for i := 0; i < len(leaderboard) && i < 3; i++ {
		podium = append(podium, leaderboard[i])
	}

	return &MatchSummaryResponse{
		MatchID:        m.ID,
		QuizTitle:      qDetail.Title,
		TotalQuestions: len(qDetail.Questions),
		TotalPlayers:   len(leaderboard),
		Podium:         podium,
		Leaderboard:    leaderboard,
	}, nil
}

// GetPublicInfoByPIN gets basic public info about a match.
func (r *sqlRepository) GetPublicInfoByPIN(ctx context.Context, pin string) (*MatchPublicInfo, error) {
	m, err := r.GetMatchByPIN(ctx, pin)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, nil
	}

	return &MatchPublicInfo{
		ID:          m.ID,
		PIN:         m.PIN,
		QuizTitle:   m.QuizTitle,
		HostName:    m.HostName,
		Status:      m.Status,
		PlayerCount: m.PlayerCount,
	}, nil
}
