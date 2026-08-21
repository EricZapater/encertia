package match

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/encertia/backend/internal/quiz"
	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

// Service defines the business logic operations for the match module.
type MatchFinishedListener interface {
	OnMatchFinished(matchID string) error
}

type Service interface {
	CreateMatch(ctx context.Context, hostID, quizID uuid.UUID) (*MatchCreatedResponse, error)
	GetMatchPublicInfo(ctx context.Context, pin string) (*MatchPublicInfo, error)
	JoinMatch(ctx context.Context, userID uuid.UUID, pin string, nickname string) (*JoinMatchResponse, error)
	GetMatchSummary(ctx context.Context, userID, matchID uuid.UUID) (*MatchSummaryResponse, error)
	GetMatchByID(ctx context.Context, id uuid.UUID) (*Match, error)
	GetMatchByPIN(ctx context.Context, pin string) (*Match, error)
	GetPlayerByMatchAndUser(ctx context.Context, matchID, userID uuid.UUID) (*MatchPlayer, error)
	RegisterFinishedListener(listener MatchFinishedListener)

	// WebSocket handling
	HandleClientConnect(ctx context.Context, client *Client)
	HandleClientDisconnect(ctx context.Context, client *Client)
	HandleWSMessage(ctx context.Context, client *Client, raw []byte)
}

type matchService struct {
	repo      Repository
	hub       *Hub
	baseURL   string
	listeners []MatchFinishedListener
}

func (s *matchService) RegisterFinishedListener(listener MatchFinishedListener) {
	s.listeners = append(s.listeners, listener)
}

// NewService creates a new instance of Service.
func NewService(repo Repository, hub *Hub, baseURL string) Service {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	return &matchService{
		repo:    repo,
		hub:     hub,
		baseURL: baseURL,
	}
}

func (s *matchService) GetMatchByID(ctx context.Context, id uuid.UUID) (*Match, error) {
	return s.repo.GetMatchByID(ctx, id)
}

func (s *matchService) GetMatchByPIN(ctx context.Context, pin string) (*Match, error) {
	return s.repo.GetMatchByPIN(ctx, pin)
}

func (s *matchService) GetPlayerByMatchAndUser(ctx context.Context, matchID, userID uuid.UUID) (*MatchPlayer, error) {
	return s.repo.GetPlayerByMatchAndUser(ctx, matchID, userID)
}

func (s *matchService) CreateMatch(ctx context.Context, hostID, quizID uuid.UUID) (*MatchCreatedResponse, error) {
	pin, err := s.repo.GenerateUniquePIN(ctx)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	m := &Match{
		ID:                   uuid.New(),
		QuizID:               quizID,
		HostID:               hostID,
		PIN:                  pin,
		Status:               StatusLobby,
		CurrentQuestionIndex: 0,
	}

	if err := s.repo.CreateMatch(ctx, m); err != nil {
		return nil, shared.ErrInternal(err)
	}

	// Fetch match with quiz details to enrich title
	matchWithQuiz, qDetail, err := s.repo.GetMatchWithQuizByID(ctx, m.ID)
	if err != nil || qDetail == nil {
		return nil, shared.ErrNotFound("QUIZ_NOT_FOUND", "El qüestionari associat no s'ha trobat.")
	}

	// Create room in hub
	s.hub.GetOrCreateRoom(pin, m.ID, hostID)

	playURL := fmt.Sprintf("%s/play?pin=%s", s.baseURL, pin)
	qrCodeURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=%s", playURL)

	return &MatchCreatedResponse{
		ID:        m.ID,
		QuizID:    m.QuizID,
		QuizTitle: qDetail.Title,
		HostID:    m.HostID,
		PIN:       m.PIN,
		Status:    m.Status,
		QRCodeURL: qrCodeURL,
		PlayURL:   playURL,
		CreatedAt: matchWithQuiz.CreatedAt,
	}, nil
}

func (s *matchService) GetMatchPublicInfo(ctx context.Context, pin string) (*MatchPublicInfo, error) {
	info, err := s.repo.GetPublicInfoByPIN(ctx, pin)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}
	if info == nil {
		return nil, shared.ErrNotFound("MATCH_NOT_FOUND", "La partida amb el PIN especificat no existeix o ja ha finalitzat.")
	}
	return info, nil
}

func (s *matchService) JoinMatch(ctx context.Context, userID uuid.UUID, pin string, nickname string) (*JoinMatchResponse, error) {
	m, err := s.repo.GetMatchByPIN(ctx, pin)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}
	if m == nil {
		return nil, shared.ErrNotFound("MATCH_NOT_FOUND", "La partida amb el PIN especificat no existeix o ja ha finalitzat.")
	}

	if m.Status != StatusLobby {
		return nil, shared.NewAppError(http.StatusConflict, "MATCH_ALREADY_STARTED", "La partida ja ha començat i no admet nous jugadors.", nil, nil)
	}

	// Check if player is already kicked
	existingPlayer, err := s.repo.GetPlayerByMatchAndUser(ctx, m.ID, userID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}
	if existingPlayer != nil && existingPlayer.IsKicked {
		return nil, shared.NewAppError(http.StatusConflict, "PLAYER_KICKED", "Has estat expulsat d'aquesta partida.", nil, nil)
	}

	player := &MatchPlayer{
		ID:          uuid.New(),
		MatchID:     m.ID,
		UserID:      userID,
		Nickname:    nickname,
		Score:       0,
		IsConnected: true,
		IsKicked:    false,
	}
	if existingPlayer != nil {
		player.ID = existingPlayer.ID
		player.Score = existingPlayer.Score
	}

	if err := s.repo.AddOrUpdatePlayer(ctx, player); err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &JoinMatchResponse{
		MatchID:  m.ID,
		PlayerID: player.ID,
		UserID:   userID,
		Nickname: nickname,
		PIN:      m.PIN,
		Status:   m.Status,
	}, nil
}

func (s *matchService) GetMatchSummary(ctx context.Context, userID, matchID uuid.UUID) (*MatchSummaryResponse, error) {
	summary, err := s.repo.GetMatchSummary(ctx, matchID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}
	if summary == nil {
		return nil, shared.ErrNotFound("MATCH_NOT_FOUND", "La partida especificada no s'ha trobat.")
	}
	return summary, nil
}

// WebSocket Event Handling

func (s *matchService) HandleClientConnect(ctx context.Context, client *Client) {
	if client.Room == nil {
		return
	}
	pin := client.Room.PIN
	m, qDetail, err := s.repo.GetMatchWithQuiz(ctx, pin)
	if err != nil || m == nil || qDetail == nil {
		_ = client.SendMessage(OutgoingWSMessage{
			Event: ServerEventError,
			Data:  ErrorPayload{Code: "MATCH_NOT_FOUND", Message: "Partida no trobada."},
		})
		return
	}

	players, _ := s.repo.GetPlayersByMatch(ctx, m.ID)

	role := "player"
	if client.IsHost {
		role = "host"
	} else if client.PlayerID != nil {
		_ = s.repo.UpdatePlayerConnection(ctx, *client.PlayerID, true)
	}

	// Prepare current question details (including options) if during active phases
	var currQ interface{}
	if m.CurrentQuestionIndex >= 0 && m.CurrentQuestionIndex < len(qDetail.Questions) {
		q := qDetail.Questions[m.CurrentQuestionIndex]
		options := make([]QuestionOptionPayload, len(q.Answers))
		for i, a := range q.Answers {
			options[i] = QuestionOptionPayload{
				ID:         a.ID,
				Text:       a.Text,
				OrderIndex: a.OrderIndex,
			}
		}

		currQ = QuestionStartedPayload{
			QuestionIndex:    m.CurrentQuestionIndex,
			TotalQuestions:   len(qDetail.Questions),
			QuestionID:       q.ID,
			Text:             q.Text,
			ImageURL:         q.ImageURL,
			QuestionType:     string(q.QuestionType),
			TimeLimitSeconds: q.TimeLimitSeconds,
			Options:          options,
		}
	}

	var matchPlayer *MatchPlayer
	if client.PlayerID != nil {
		matchPlayer, _ = s.repo.GetPlayerByID(ctx, *client.PlayerID)
	}

	// Send initial state to the connecting client
	stateMsg := OutgoingWSMessage{
		Event: ServerEventMatchState,
		Data: MatchStatePayload{
			MatchID:              m.ID,
			PIN:                  m.PIN,
			QuizTitle:            qDetail.Title,
			Status:               m.Status,
			CurrentQuestionIndex: m.CurrentQuestionIndex,
			TotalQuestions:       len(qDetail.Questions),
			Role:                 role,
			Player:               matchPlayer,
			Players:              players,
			CurrentQuestion:      currQ,
		},
	}
	_ = client.SendMessage(stateMsg)

	// If player joined/reconnected, broadcast to the room
	if !client.IsHost && client.PlayerID != nil {
		client.Room.Broadcast(OutgoingWSMessage{
			Event: ServerEventPlayerJoined,
			Data: PlayerJoinedPayload{
				PlayerID:     *client.PlayerID,
				Nickname:     client.Nickname,
				TotalPlayers: len(players),
			},
		})
	}
}

func (s *matchService) HandleClientDisconnect(ctx context.Context, client *Client) {
	if client.Room == nil {
		return
	}
	if !client.IsHost && client.PlayerID != nil {
		_ = s.repo.UpdatePlayerConnection(ctx, *client.PlayerID, false)
		client.Room.Broadcast(OutgoingWSMessage{
			Event: ServerEventPlayerLeft,
			Data: PlayerLeftPayload{
				PlayerID:     *client.PlayerID,
				Nickname:     client.Nickname,
				TotalPlayers: client.Room.ConnectedPlayerCount(),
			},
		})
	}
}

func (s *matchService) HandleWSMessage(ctx context.Context, client *Client, raw []byte) {
	var msg WSMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("[WS Service] Error parsejant missatge JSON: %v", err)
		return
	}

	if client.Room == nil {
		return
	}

	switch msg.Event {
	// Host Events
	case HostEventStartMatch:
		s.handleHostStartMatch(ctx, client)
	case HostEventStartQuestionTimer:
		s.handleHostStartQuestionTimer(ctx, client)
	case HostEventShowResults:
		s.handleHostShowResults(ctx, client)
	case HostEventShowLeaderboard:
		s.handleHostShowLeaderboard(ctx, client)
	case HostEventNextQuestion:
		s.handleHostNextQuestion(ctx, client)
	case HostEventKickPlayer:
		s.handleHostKickPlayer(ctx, client, msg.Data)

	// Player Events
	case PlayerEventSubmitAnswer:
		s.handlePlayerSubmitAnswer(ctx, client, msg.Data)

	default:
		log.Printf("[WS Service] Esdeveniment desconegut: %s", msg.Event)
	}
}

func (s *matchService) handleHostStartMatch(ctx context.Context, client *Client) {
	if !client.IsHost {
		return
	}
	pin := client.Room.PIN
	m, qDetail, err := s.repo.GetMatchWithQuiz(ctx, pin)
	if err != nil || m == nil || qDetail == nil || len(qDetail.Questions) == 0 {
		_ = client.SendMessage(OutgoingWSMessage{
			Event: ServerEventError,
			Data:  ErrorPayload{Code: "INVALID_QUIZ", Message: "El qüestionari no conté cap pregunta."},
		})
		return
	}

	// Transition to question_preview for question 0
	m.Status = StatusQuestionPreview
	m.CurrentQuestionIndex = 0
	if err := s.repo.UpdateMatchStatus(ctx, m.ID, m.Status, m.CurrentQuestionIndex, nil); err != nil {
		log.Printf("[WS Service] Error actualitzant estat de partida: %v", err)
		return
	}

	firstQ := qDetail.Questions[0]
	previewData := QuestionPreviewPayload{
		QuestionIndex:    0,
		TotalQuestions:   len(qDetail.Questions),
		QuestionID:       firstQ.ID,
		Text:             firstQ.Text,
		ImageURL:         firstQ.ImageURL,
		QuestionType:     string(firstQ.QuestionType),
		TimeLimitSeconds: firstQ.TimeLimitSeconds,
	}

	client.Room.Broadcast(OutgoingWSMessage{
		Event: ServerEventQuestionPreview,
		Data:  previewData,
	})
}

func (s *matchService) handleHostStartQuestionTimer(ctx context.Context, client *Client) {
	if !client.IsHost {
		return
	}
	pin := client.Room.PIN
	m, qDetail, err := s.repo.GetMatchWithQuiz(ctx, pin)
	if err != nil || m == nil || qDetail == nil || m.CurrentQuestionIndex >= len(qDetail.Questions) {
		return
	}

	now := time.Now().UTC()
	m.Status = StatusQuestionActive
	m.QuestionStartedAt = &now

	if err := s.repo.UpdateMatchStatus(ctx, m.ID, m.Status, m.CurrentQuestionIndex, m.QuestionStartedAt); err != nil {
		log.Printf("[WS Service] Error iniciant temporitzador de pregunta: %v", err)
		return
	}

	currQ := qDetail.Questions[m.CurrentQuestionIndex]
	options := make([]QuestionOptionPayload, len(currQ.Answers))
	for i, a := range currQ.Answers {
		options[i] = QuestionOptionPayload{
			ID:         a.ID,
			Text:       a.Text,
			OrderIndex: a.OrderIndex,
		}
	}

	startedData := QuestionStartedPayload{
		QuestionIndex:    m.CurrentQuestionIndex,
		TotalQuestions:   len(qDetail.Questions),
		QuestionID:       currQ.ID,
		Text:             currQ.Text,
		ImageURL:         currQ.ImageURL,
		QuestionType:     string(currQ.QuestionType),
		TimeLimitSeconds: currQ.TimeLimitSeconds,
		StartedAt:        now,
		Options:          options,
	}

	client.Room.Broadcast(OutgoingWSMessage{
		Event: ServerEventQuestionStarted,
		Data:  startedData,
	})
}

func (s *matchService) handleHostShowResults(ctx context.Context, client *Client) {
	if !client.IsHost {
		return
	}
	pin := client.Room.PIN
	m, qDetail, err := s.repo.GetMatchWithQuiz(ctx, pin)
	if err != nil || m == nil || qDetail == nil || m.CurrentQuestionIndex >= len(qDetail.Questions) {
		return
	}

	m.Status = StatusQuestionResults
	if err := s.repo.UpdateMatchStatus(ctx, m.ID, m.Status, m.CurrentQuestionIndex, m.QuestionStartedAt); err != nil {
		log.Printf("[WS Service] Error actualitzant a question_results: %v", err)
		return
	}

	currQ := qDetail.Questions[m.CurrentQuestionIndex]
	answers, err := s.repo.GetAnswersForQuestion(ctx, m.ID, currQ.ID)
	if err != nil {
		log.Printf("[WS Service] Error obtenint respostes per pregunta: %v", err)
		return
	}

	players, _ := s.repo.GetPlayersByMatch(ctx, m.ID)

	// Tally votes per option and identify correct answer IDs
	optionCountsMap := make(map[uuid.UUID]int)
	for _, a := range answers {
		for _, optID := range a.SelectedAnswerIDs {
			optionCountsMap[optID]++
		}
	}

	var correctIDs []uuid.UUID
	optionResults := make([]OptionResultCount, len(currQ.Answers))
	for i, opt := range currQ.Answers {
		if opt.IsCorrect {
			correctIDs = append(correctIDs, opt.ID)
		}
		optionResults[i] = OptionResultCount{
			OptionID:  opt.ID,
			Text:      opt.Text,
			Count:     optionCountsMap[opt.ID],
			IsCorrect: opt.IsCorrect,
		}
	}

	// 1. Send Host summary
	client.Room.SendToHost(OutgoingWSMessage{
		Event: ServerEventQuestionEnded,
		Data: QuestionEndedHostPayload{
			QuestionID:       currQ.ID,
			TotalAnswered:    len(answers),
			TotalPlayers:     len(players),
			CorrectAnswerIDs: correctIDs,
			OptionCounts:     optionResults,
		},
	})

	// 2. Send personalized result to each player
	playerAnswersMap := make(map[uuid.UUID]MatchAnswer)
	for _, a := range answers {
		playerAnswersMap[a.PlayerID] = a
	}

	for _, p := range players {
		pAns, answered := playerAnswersMap[p.ID]
		isCorrect := false
		scoreAwarded := 0
		if answered {
			isCorrect = pAns.IsCorrect
			scoreAwarded = pAns.ScoreAwarded
		}

		client.Room.SendToPlayer(p.ID, OutgoingWSMessage{
			Event: ServerEventQuestionEnded,
			Data: QuestionEndedPlayerPayload{
				QuestionID:       currQ.ID,
				IsCorrect:        isCorrect,
				ScoreAwarded:     scoreAwarded,
				TotalScore:       p.Score,
				CorrectAnswerIDs: correctIDs,
			},
		})
	}
}

func (s *matchService) handleHostShowLeaderboard(ctx context.Context, client *Client) {
	if !client.IsHost {
		return
	}
	pin := client.Room.PIN
	m, qDetail, err := s.repo.GetMatchWithQuiz(ctx, pin)
	if err != nil || m == nil || qDetail == nil {
		return
	}

	m.Status = StatusLeaderboard
	if err := s.repo.UpdateMatchStatus(ctx, m.ID, m.Status, m.CurrentQuestionIndex, m.QuestionStartedAt); err != nil {
		log.Printf("[WS Service] Error actualitzant a leaderboard: %v", err)
		return
	}

	leaderboard, err := s.repo.GetLeaderboard(ctx, m.ID)
	if err != nil {
		log.Printf("[WS Service] Error calculant rànquing: %v", err)
		return
	}

	client.Room.Broadcast(OutgoingWSMessage{
		Event: ServerEventLeaderboard,
		Data: LeaderboardPayload{
			QuestionIndex:  m.CurrentQuestionIndex,
			TotalQuestions: len(qDetail.Questions),
			Items:          leaderboard,
		},
	})
}

func (s *matchService) handleHostNextQuestion(ctx context.Context, client *Client) {
	if !client.IsHost {
		return
	}
	pin := client.Room.PIN
	m, qDetail, err := s.repo.GetMatchWithQuiz(ctx, pin)
	if err != nil || m == nil || qDetail == nil {
		return
	}

	nextIdx := m.CurrentQuestionIndex + 1
	if nextIdx < len(qDetail.Questions) {
		// Advance to next question preview
		m.Status = StatusQuestionPreview
		m.CurrentQuestionIndex = nextIdx
		if err := s.repo.UpdateMatchStatus(ctx, m.ID, m.Status, m.CurrentQuestionIndex, nil); err != nil {
			log.Printf("[WS Service] Error avançant a la següent pregunta: %v", err)
			return
		}

		nextQ := qDetail.Questions[nextIdx]
		client.Room.Broadcast(OutgoingWSMessage{
			Event: ServerEventQuestionPreview,
			Data: QuestionPreviewPayload{
				QuestionIndex:    nextIdx,
				TotalQuestions:   len(qDetail.Questions),
				QuestionID:       nextQ.ID,
				Text:             nextQ.Text,
				ImageURL:         nextQ.ImageURL,
				QuestionType:     string(nextQ.QuestionType),
				TimeLimitSeconds: nextQ.TimeLimitSeconds,
			},
		})
	} else {
		// All questions completed -> Finish match
		m.Status = StatusFinished
		if err := s.repo.UpdateMatchStatus(ctx, m.ID, m.Status, m.CurrentQuestionIndex, nil); err != nil {
			log.Printf("[WS Service] Error finalitzant partida: %v", err)
			return
		}

		summary, err := s.repo.GetMatchSummary(ctx, m.ID)
		if err != nil {
			log.Printf("[WS Service] Error generant resum final: %v", err)
			return
		}

		client.Room.Broadcast(OutgoingWSMessage{
			Event: ServerEventFinished,
			Data:  summary,
		})
	}
}

func (s *matchService) handleHostKickPlayer(ctx context.Context, client *Client, rawData json.RawMessage) {
	if !client.IsHost {
		return
	}
	var payload KickPlayerPayload
	if err := json.Unmarshal(rawData, &payload); err != nil {
		return
	}

	pin := client.Room.PIN
	m, err := s.repo.GetMatchByPIN(ctx, pin)
	if err != nil || m == nil {
		return
	}

	if err := s.repo.KickPlayer(ctx, m.ID, payload.PlayerID); err != nil {
		log.Printf("[WS Service] Error expulsant jugador: %v", err)
		return
	}

	// Notify room & kick the specific connection
	client.Room.Broadcast(OutgoingWSMessage{
		Event: ServerEventPlayerKicked,
		Data:  PlayerKickedPayload{PlayerID: payload.PlayerID},
	})

	playerClient := client.Room.GetPlayerClient(payload.PlayerID)
	if playerClient != nil {
		playerClient.Close()
		client.Room.UnregisterClient(playerClient)
	}
}

func (s *matchService) handlePlayerSubmitAnswer(ctx context.Context, client *Client, rawData json.RawMessage) {
	if client.IsHost || client.PlayerID == nil {
		return
	}

	var payload SubmitAnswerPayload
	if err := json.Unmarshal(rawData, &payload); err != nil {
		return
	}

	pin := client.Room.PIN
	m, qDetail, err := s.repo.GetMatchWithQuiz(ctx, pin)
	if err != nil || m == nil || qDetail == nil {
		return
	}

	// Verify match status
	if m.Status != StatusQuestionActive {
		_ = client.SendMessage(OutgoingWSMessage{
			Event: ServerEventError,
			Data:  ErrorPayload{Code: "NOT_ACTIVE", Message: "El període de resposta no està actiu."},
		})
		return
	}

	if m.CurrentQuestionIndex >= len(qDetail.Questions) {
		return
	}
	currQ := qDetail.Questions[m.CurrentQuestionIndex]
	if currQ.ID != payload.QuestionID {
		_ = client.SendMessage(OutgoingWSMessage{
			Event: ServerEventError,
			Data:  ErrorPayload{Code: "QUESTION_MISMATCH", Message: "La ID de la pregunta no coincideix."},
		})
		return
	}

	// Calculate response time
	responseTimeMS := 0
	if m.QuestionStartedAt != nil {
		diff := time.Since(*m.QuestionStartedAt)
		responseTimeMS = int(diff.Milliseconds())
	}

	// Check answer correctness
	correctMap := make(map[uuid.UUID]bool)
	correctCount := 0
	for _, a := range currQ.Answers {
		if a.IsCorrect {
			correctMap[a.ID] = true
			correctCount++
		}
	}

	isCorrect := false
	if currQ.QuestionType == quiz.QuestionTypeSingle {
		if len(payload.AnswerIDs) == 1 && correctMap[payload.AnswerIDs[0]] {
			isCorrect = true
		}
	} else if currQ.QuestionType == quiz.QuestionTypeMultiple {
		// Must select all correct and no incorrect
		selectedCorrect := 0
		hasIncorrect := false
		for _, selectedID := range payload.AnswerIDs {
			if correctMap[selectedID] {
				selectedCorrect++
			} else {
				hasIncorrect = true
			}
		}
		if !hasIncorrect && selectedCorrect == correctCount {
			isCorrect = true
		}
	}

	scoreAwarded := 0
	if isCorrect {
		scoreAwarded = 1
	}

	matchAns := &MatchAnswer{
		ID:                uuid.New(),
		MatchID:           m.ID,
		QuestionID:        currQ.ID,
		PlayerID:          *client.PlayerID,
		SelectedAnswerIDs: payload.AnswerIDs,
		IsCorrect:         isCorrect,
		ScoreAwarded:      scoreAwarded,
		ResponseTimeMS:    responseTimeMS,
	}

	if err := s.repo.RecordAnswer(ctx, matchAns); err != nil {
		log.Printf("[WS Service] Error registrant resposta a la bd per al jugador %s: %v", client.PlayerID, err)
		_ = client.SendMessage(OutgoingWSMessage{
			Event: ServerEventError,
			Data:  ErrorPayload{Code: "RECORD_ANSWER_FAILED", Message: "No s'ha pogut registrar la resposta o ja havia estat enviada."},
		})
		return
	}

	// Notify host with updated answer stats
	answers, _ := s.repo.GetAnswersForQuestion(ctx, m.ID, currQ.ID)
	players, _ := s.repo.GetPlayersByMatch(ctx, m.ID)

	client.Room.SendToHost(OutgoingWSMessage{
		Event: ServerEventAnswerStats,
		Data: AnswerStatsPayload{
			AnsweredCount: len(answers),
			TotalPlayers:  len(players),
		},
	})
}
