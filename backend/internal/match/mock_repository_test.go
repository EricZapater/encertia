package match_test

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/encertia/backend/internal/match"
	"github.com/encertia/backend/internal/quiz"
	"github.com/google/uuid"
)

type mockRepository struct {
	mu         sync.RWMutex
	matches    map[uuid.UUID]*match.Match
	quizzes    map[uuid.UUID]*quiz.QuizDetail
	players    map[uuid.UUID]*match.MatchPlayer // keyed by PlayerID
	answers    []match.MatchAnswer
	nextPINSeq int
	forceError bool
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		matches:    make(map[uuid.UUID]*match.Match),
		quizzes:    make(map[uuid.UUID]*quiz.QuizDetail),
		players:    make(map[uuid.UUID]*match.MatchPlayer),
		answers:    make([]match.MatchAnswer, 0),
		nextPINSeq: 100000,
	}
}

func (m *mockRepository) GenerateUniquePIN(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.forceError {
		return "", errors.New("db error")
	}

	m.nextPINSeq++
	return fmt.Sprintf("%06d", m.nextPINSeq), nil
}

func (m *mockRepository) CreateMatch(ctx context.Context, matchObj *match.Match) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.forceError {
		return errors.New("db error")
	}

	if matchObj.ID == uuid.Nil {
		matchObj.ID = uuid.New()
	}
	matchObj.CreatedAt = time.Now()
	matchObj.UpdatedAt = time.Now()
	m.matches[matchObj.ID] = matchObj
	return nil
}

func (m *mockRepository) GetMatchByID(ctx context.Context, id uuid.UUID) (*match.Match, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, errors.New("db error")
	}

	if matchObj, ok := m.matches[id]; ok && matchObj.DeletedAt == nil {
		copy := *matchObj
		return &copy, nil
	}
	return nil, nil
}

func (m *mockRepository) GetMatchByPIN(ctx context.Context, pin string) (*match.Match, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, errors.New("db error")
	}

	for _, matchObj := range m.matches {
		if matchObj.PIN == pin && matchObj.Status != match.StatusFinished && matchObj.DeletedAt == nil {
			copy := *matchObj
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) GetMatchWithQuiz(ctx context.Context, pin string) (*match.Match, *quiz.QuizDetail, error) {
	matchObj, err := m.GetMatchByPIN(ctx, pin)
	if err != nil || matchObj == nil {
		return nil, nil, err
	}
	return m.GetMatchWithQuizByID(ctx, matchObj.ID)
}

func (m *mockRepository) GetMatchWithQuizByID(ctx context.Context, id uuid.UUID) (*match.Match, *quiz.QuizDetail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, nil, errors.New("db error")
	}

	matchObj, ok := m.matches[id]
	if !ok || matchObj.DeletedAt != nil {
		return nil, nil, nil
	}

	qd, ok := m.quizzes[matchObj.QuizID]
	if !ok {
		return nil, nil, errors.New("quiz not found")
	}

	matchCopy := *matchObj
	matchCopy.QuizTitle = qd.Title
	qdCopy := *qd
	return &matchCopy, &qdCopy, nil
}

func (m *mockRepository) UpdateMatchStatus(ctx context.Context, matchID uuid.UUID, status match.MatchStatus, currentQuestionIndex int, startedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.forceError {
		return errors.New("db error")
	}

	if matchObj, ok := m.matches[matchID]; ok {
		matchObj.Status = status
		matchObj.CurrentQuestionIndex = currentQuestionIndex
		matchObj.QuestionStartedAt = startedAt
		matchObj.UpdatedAt = time.Now()
		return nil
	}
	return errors.New("match not found")
}

func (m *mockRepository) AddOrUpdatePlayer(ctx context.Context, player *match.MatchPlayer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.forceError {
		return errors.New("db error")
	}

	// Check if exists by match_id and user_id
	for _, p := range m.players {
		if p.MatchID == player.MatchID && p.UserID == player.UserID {
			p.Nickname = player.Nickname
			p.IsConnected = player.IsConnected
			p.UpdatedAt = time.Now()
			player.ID = p.ID
			player.Score = p.Score
			player.IsKicked = p.IsKicked
			player.JoinedAt = p.JoinedAt
			return nil
		}
	}

	if player.ID == uuid.Nil {
		player.ID = uuid.New()
	}
	player.JoinedAt = time.Now()
	player.UpdatedAt = time.Now()
	m.players[player.ID] = player
	return nil
}

func (m *mockRepository) GetPlayerByMatchAndUser(ctx context.Context, matchID, userID uuid.UUID) (*match.MatchPlayer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, errors.New("db error")
	}

	for _, p := range m.players {
		if p.MatchID == matchID && p.UserID == userID {
			copy := *p
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *mockRepository) GetPlayerByID(ctx context.Context, playerID uuid.UUID) (*match.MatchPlayer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, errors.New("db error")
	}

	if p, ok := m.players[playerID]; ok {
		copy := *p
		return &copy, nil
	}
	return nil, nil
}

func (m *mockRepository) GetPlayersByMatch(ctx context.Context, matchID uuid.UUID) ([]match.MatchPlayer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, errors.New("db error")
	}

	var result []match.MatchPlayer
	for _, p := range m.players {
		if p.MatchID == matchID && !p.IsKicked {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (m *mockRepository) UpdatePlayerConnection(ctx context.Context, playerID uuid.UUID, isConnected bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, ok := m.players[playerID]; ok {
		p.IsConnected = isConnected
		p.UpdatedAt = time.Now()
		return nil
	}
	return errors.New("player not found")
}

func (m *mockRepository) KickPlayer(ctx context.Context, matchID, playerID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, ok := m.players[playerID]; ok && p.MatchID == matchID {
		p.IsKicked = true
		p.IsConnected = false
		p.UpdatedAt = time.Now()
		return nil
	}
	return errors.New("player not found")
}

func (m *mockRepository) RecordAnswer(ctx context.Context, answer *match.MatchAnswer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.forceError {
		return errors.New("db error")
	}

	if answer.ID == uuid.Nil {
		answer.ID = uuid.New()
	}
	answer.AnsweredAt = time.Now()
	m.answers = append(m.answers, *answer)

	if answer.ScoreAwarded > 0 {
		if p, ok := m.players[answer.PlayerID]; ok {
			p.Score += answer.ScoreAwarded
		}
	}
	return nil
}

func (m *mockRepository) GetAnswersForQuestion(ctx context.Context, matchID, questionID uuid.UUID) ([]match.MatchAnswer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, errors.New("db error")
	}

	var result []match.MatchAnswer
	for _, a := range m.answers {
		if a.MatchID == matchID && a.QuestionID == questionID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockRepository) GetLeaderboard(ctx context.Context, matchID uuid.UUID) ([]match.PlayerScoreItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.forceError {
		return nil, errors.New("db error")
	}

	var activePlayers []*match.MatchPlayer
	for _, p := range m.players {
		if p.MatchID == matchID && !p.IsKicked {
			activePlayers = append(activePlayers, p)
		}
	}

	// Sort by score desc, joined_at asc
	sort.Slice(activePlayers, func(i, j int) bool {
		if activePlayers[i].Score == activePlayers[j].Score {
			return activePlayers[i].JoinedAt.Before(activePlayers[j].JoinedAt)
		}
		return activePlayers[i].Score > activePlayers[j].Score
	})

	var items []match.PlayerScoreItem
	currentRank := 1
	for i, p := range activePlayers {
		if i > 0 && p.Score < activePlayers[i-1].Score {
			currentRank = i + 1
		}

		correctCount := 0
		totalAnswered := 0
		for _, a := range m.answers {
			if a.PlayerID == p.ID && a.MatchID == matchID {
				totalAnswered++
				if a.IsCorrect {
					correctCount++
				}
			}
		}

		items = append(items, match.PlayerScoreItem{
			PlayerID:      p.ID,
			UserID:        p.UserID,
			Nickname:      p.Nickname,
			Score:         p.Score,
			Rank:          currentRank,
			CorrectCount:  correctCount,
			TotalAnswered: totalAnswered,
		})
	}

	if items == nil {
		items = []match.PlayerScoreItem{}
	}
	return items, nil
}

func (m *mockRepository) GetMatchSummary(ctx context.Context, matchID uuid.UUID) (*match.MatchSummaryResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	matchObj, ok := m.matches[matchID]
	if !ok {
		return nil, nil
	}
	qd, ok := m.quizzes[matchObj.QuizID]
	if !ok {
		return nil, nil
	}

	m.mu.RUnlock()
	leaderboard, err := m.GetLeaderboard(ctx, matchID)
	m.mu.RLock()
	if err != nil {
		return nil, err
	}

	podium := make([]match.PlayerScoreItem, 0, 3)
	for i := 0; i < len(leaderboard) && i < 3; i++ {
		podium = append(podium, leaderboard[i])
	}

	return &match.MatchSummaryResponse{
		MatchID:        matchObj.ID,
		QuizTitle:      qd.Title,
		TotalQuestions: len(qd.Questions),
		TotalPlayers:   len(leaderboard),
		Podium:         podium,
		Leaderboard:    leaderboard,
	}, nil
}

func (m *mockRepository) GetPublicInfoByPIN(ctx context.Context, pin string) (*match.MatchPublicInfo, error) {
	matchObj, err := m.GetMatchByPIN(ctx, pin)
	if err != nil || matchObj == nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	qd := m.quizzes[matchObj.QuizID]
	title := ""
	if qd != nil {
		title = qd.Title
	}

	playerCount := 0
	for _, p := range m.players {
		if p.MatchID == matchObj.ID && !p.IsKicked {
			playerCount++
		}
	}

	return &match.MatchPublicInfo{
		ID:          matchObj.ID,
		PIN:         matchObj.PIN,
		QuizTitle:   title,
		HostName:    "Host User",
		Status:      matchObj.Status,
		PlayerCount: playerCount,
	}, nil
}
