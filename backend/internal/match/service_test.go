package match_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/encertia/backend/internal/match"
	"github.com/encertia/backend/internal/quiz"
	"github.com/google/uuid"
)

func setupTestService() (*mockRepository, *match.Hub, match.Service, uuid.UUID, *quiz.QuizDetail) {
	repo := newMockRepository()
	hub := match.NewHub()
	svc := match.NewService(repo, hub, "https://encertia.test")

	hostID := uuid.New()
	q1Ans1 := quiz.QuizAnswer{ID: uuid.New(), Text: "Opció 1 (Correcta)", IsCorrect: true, OrderIndex: 0}
	q1Ans2 := quiz.QuizAnswer{ID: uuid.New(), Text: "Opció 2 (Incorrecta)", IsCorrect: false, OrderIndex: 1}

	q2Ans1 := quiz.QuizAnswer{ID: uuid.New(), Text: "Opció Multi 1 (Correcta)", IsCorrect: true, OrderIndex: 0}
	q2Ans2 := quiz.QuizAnswer{ID: uuid.New(), Text: "Opció Multi 2 (Correcta)", IsCorrect: true, OrderIndex: 1}
	q2Ans3 := quiz.QuizAnswer{ID: uuid.New(), Text: "Opció Multi 3 (Incorrecta)", IsCorrect: false, OrderIndex: 2}

	qd := &quiz.QuizDetail{
		Quiz: quiz.Quiz{
			ID:          uuid.New(),
			CreatorID:   hostID,
			CreatorName: "Professor Prova",
			Title:       "Qüestionari de Prova",
			Status:      quiz.StatusPublished,
		},
		Questions: []quiz.QuizQuestion{
			{
				ID:               uuid.New(),
				Text:             "Pregunta 1 Single",
				QuestionType:     quiz.QuestionTypeSingle,
				TimeLimitSeconds: 20,
				OrderIndex:       0,
				Answers:          []quiz.QuizAnswer{q1Ans1, q1Ans2},
			},
			{
				ID:               uuid.New(),
				Text:             "Pregunta 2 Multiple",
				QuestionType:     quiz.QuestionTypeMultiple,
				TimeLimitSeconds: 30,
				OrderIndex:       1,
				Answers:          []quiz.QuizAnswer{q2Ans1, q2Ans2, q2Ans3},
			},
		},
	}
	repo.quizzes[qd.ID] = qd

	return repo, hub, svc, hostID, qd
}

func TestMatchService_CreateAndPublicInfo(t *testing.T) {
	ctx := context.Background()
	_, _, svc, hostID, qd := setupTestService()

	created, err := svc.CreateMatch(ctx, hostID, qd.ID)
	if err != nil {
		t.Fatalf("error creant partida: %v", err)
	}

	if created.PIN == "" || len(created.PIN) != 6 {
		t.Errorf("PIN invàlid: %s", created.PIN)
	}
	if created.Status != match.StatusLobby {
		t.Errorf("estat inicial no és lobby: %s", created.Status)
	}
	if created.QuizTitle != qd.Title {
		t.Errorf("títol incorrecte: got %s, want %s", created.QuizTitle, qd.Title)
	}

	// Public Info
	info, err := svc.GetMatchPublicInfo(ctx, created.PIN)
	if err != nil {
		t.Fatalf("error obtenint info pública: %v", err)
	}
	if info.PIN != created.PIN || info.Status != match.StatusLobby {
		t.Errorf("dades públiques incorrectes: %+v", info)
	}
}

func TestMatchService_JoinMatch(t *testing.T) {
	ctx := context.Background()
	repo, _, svc, hostID, qd := setupTestService()

	created, _ := svc.CreateMatch(ctx, hostID, qd.ID)
	userID := uuid.New()

	// 1. Successful Join
	joinRes, err := svc.JoinMatch(ctx, userID, created.PIN, "Joan")
	if err != nil {
		t.Fatalf("error unint-se a la partida: %v", err)
	}
	if joinRes.Nickname != "Joan" || joinRes.PIN != created.PIN {
		t.Errorf("dades d'unió incorrectes: %+v", joinRes)
	}

	// 2. Conflict when match already started
	_ = repo.UpdateMatchStatus(ctx, created.ID, match.StatusQuestionActive, 0, nil)
	_, err = svc.JoinMatch(ctx, uuid.New(), created.PIN, "Maria")
	if err == nil {
		t.Errorf("esperava error 409 ja que la partida ha començat")
	}

	// 3. Conflict when player kicked
	_ = repo.UpdateMatchStatus(ctx, created.ID, match.StatusLobby, 0, nil)
	_ = repo.KickPlayer(ctx, created.ID, joinRes.PlayerID)
	_, err = svc.JoinMatch(ctx, userID, created.PIN, "Joan")
	if err == nil {
		t.Errorf("esperava error 409 ja que el jugador ha estat expulsat")
	}
}

func TestMatchService_WSLifecycle(t *testing.T) {
	ctx := context.Background()
	_, hub, svc, hostID, qd := setupTestService()

	created, _ := svc.CreateMatch(ctx, hostID, qd.ID)
	room := hub.GetRoom(created.PIN)

	// Join player 1
	p1UserID := uuid.New()
	join1, _ := svc.JoinMatch(ctx, p1UserID, created.PIN, "Alumne 1")

	// Join player 2
	p2UserID := uuid.New()
	join2, _ := svc.JoinMatch(ctx, p2UserID, created.PIN, "Alumne 2")

	// Connect Host WS
	hostClient := &match.Client{
		Hub:      hub,
		Room:     room,
		Send:     make(chan []byte, 20),
		UserID:   hostID,
		Nickname: "Host",
		IsHost:   true,
	}
	room.RegisterClient(hostClient)
	svc.HandleClientConnect(ctx, hostClient)

	// Connect Player 1 WS
	p1Client := &match.Client{
		Hub:      hub,
		Room:     room,
		Send:     make(chan []byte, 20),
		UserID:   p1UserID,
		PlayerID: &join1.PlayerID,
		Nickname: "Alumne 1",
		IsHost:   false,
	}
	room.RegisterClient(p1Client)
	svc.HandleClientConnect(ctx, p1Client)

	// Connect Player 2 WS
	p2Client := &match.Client{
		Hub:      hub,
		Room:     room,
		Send:     make(chan []byte, 20),
		UserID:   p2UserID,
		PlayerID: &join2.PlayerID,
		Nickname: "Alumne 2",
		IsHost:   false,
	}
	room.RegisterClient(p2Client)
	svc.HandleClientConnect(ctx, p2Client)

	// 1. Host starts match -> question_preview
	startMsg, _ := json.Marshal(match.WSMessage{Event: match.HostEventStartMatch})
	svc.HandleWSMessage(ctx, hostClient, startMsg)

	m, _ := svc.GetMatchByID(ctx, created.ID)
	if m.Status != match.StatusQuestionPreview || m.CurrentQuestionIndex != 0 {
		t.Errorf("esperava question_preview a índex 0, got %s i %d", m.Status, m.CurrentQuestionIndex)
	}

	// 2. Host starts question timer -> question_active
	timerMsg, _ := json.Marshal(match.WSMessage{Event: match.HostEventStartQuestionTimer})
	svc.HandleWSMessage(ctx, hostClient, timerMsg)

	m, _ = svc.GetMatchByID(ctx, created.ID)
	if m.Status != match.StatusQuestionActive || m.QuestionStartedAt == nil {
		t.Errorf("esperava question_active amb timer iniciat, got %s", m.Status)
	}

	// 3. Player 1 submits CORRECT answer (Single Choice)
	q1 := qd.Questions[0]
	correctAnsID := q1.Answers[0].ID
	incorrectAnsID := q1.Answers[1].ID

	p1AnswerMsg, _ := json.Marshal(match.WSMessage{
		Event: match.PlayerEventSubmitAnswer,
		Data: func() json.RawMessage {
			b, _ := json.Marshal(match.SubmitAnswerPayload{
				QuestionID: q1.ID,
				AnswerIDs:  []uuid.UUID{correctAnsID},
			})
			return b
		}(),
	})
	svc.HandleWSMessage(ctx, p1Client, p1AnswerMsg)

	// Player 2 submits INCORRECT answer
	p2AnswerMsg, _ := json.Marshal(match.WSMessage{
		Event: match.PlayerEventSubmitAnswer,
		Data: func() json.RawMessage {
			b, _ := json.Marshal(match.SubmitAnswerPayload{
				QuestionID: q1.ID,
				AnswerIDs:  []uuid.UUID{incorrectAnsID},
			})
			return b
		}(),
	})
	svc.HandleWSMessage(ctx, p2Client, p2AnswerMsg)

	// Check updated scores in DB
	p1Record, _ := svc.GetPlayerByMatchAndUser(ctx, m.ID, p1UserID)
	if p1Record.Score != 1 {
		t.Errorf("esperava score 1 per a P1, got %d", p1Record.Score)
	}
	p2Record, _ := svc.GetPlayerByMatchAndUser(ctx, m.ID, p2UserID)
	if p2Record.Score != 0 {
		t.Errorf("esperava score 0 per a P2, got %d", p2Record.Score)
	}

	// 4. Host shows results -> question_results
	showResultsMsg, _ := json.Marshal(match.WSMessage{Event: match.HostEventShowResults})
	svc.HandleWSMessage(ctx, hostClient, showResultsMsg)

	m, _ = svc.GetMatchByID(ctx, created.ID)
	if m.Status != match.StatusQuestionResults {
		t.Errorf("esperava question_results, got %s", m.Status)
	}

	// 5. Host shows leaderboard -> leaderboard
	leaderboardMsg, _ := json.Marshal(match.WSMessage{Event: match.HostEventShowLeaderboard})
	svc.HandleWSMessage(ctx, hostClient, leaderboardMsg)

	m, _ = svc.GetMatchByID(ctx, created.ID)
	if m.Status != match.StatusLeaderboard {
		t.Errorf("esperava leaderboard, got %s", m.Status)
	}

	// 6. Host advances to next question (Question 2: Multiple Choice)
	nextQMsg, _ := json.Marshal(match.WSMessage{Event: match.HostEventNextQuestion})
	svc.HandleWSMessage(ctx, hostClient, nextQMsg)

	m, _ = svc.GetMatchByID(ctx, created.ID)
	if m.Status != match.StatusQuestionPreview || m.CurrentQuestionIndex != 1 {
		t.Errorf("esperava question_preview índex 1, got %s i %d", m.Status, m.CurrentQuestionIndex)
	}

	// Start timer for Question 2
	svc.HandleWSMessage(ctx, hostClient, timerMsg)

	// Question 2 Multiple Choice: P1 selects BOTH correct answers
	q2 := qd.Questions[1]
	p1Q2AnswerMsg, _ := json.Marshal(match.WSMessage{
		Event: match.PlayerEventSubmitAnswer,
		Data: func() json.RawMessage {
			b, _ := json.Marshal(match.SubmitAnswerPayload{
				QuestionID: q2.ID,
				AnswerIDs:  []uuid.UUID{q2.Answers[0].ID, q2.Answers[1].ID},
			})
			return b
		}(),
	})
	svc.HandleWSMessage(ctx, p1Client, p1Q2AnswerMsg)

	// P2 selects only 1 correct and 1 incorrect answer (wrong)
	p2Q2AnswerMsg, _ := json.Marshal(match.WSMessage{
		Event: match.PlayerEventSubmitAnswer,
		Data: func() json.RawMessage {
			b, _ := json.Marshal(match.SubmitAnswerPayload{
				QuestionID: q2.ID,
				AnswerIDs:  []uuid.UUID{q2.Answers[0].ID, q2.Answers[2].ID},
			})
			return b
		}(),
	})
	svc.HandleWSMessage(ctx, p2Client, p2Q2AnswerMsg)

	p1Record, _ = svc.GetPlayerByMatchAndUser(ctx, m.ID, p1UserID)
	if p1Record.Score != 2 {
		t.Errorf("esperava score 2 per a P1, got %d", p1Record.Score)
	}

	// 7. Kick Player Test (during active match)
	kickMsg, _ := json.Marshal(match.WSMessage{
		Event: match.HostEventKickPlayer,
		Data: func() json.RawMessage {
			b, _ := json.Marshal(match.KickPlayerPayload{PlayerID: join2.PlayerID})
			return b
		}(),
	})
	svc.HandleWSMessage(ctx, hostClient, kickMsg)

	p2Record, _ = svc.GetPlayerByMatchAndUser(ctx, m.ID, p2UserID)
	if !p2Record.IsKicked {
		t.Errorf("P2 hauria d'estar expulsat")
	}

	// 8. Host finishes match on final NextQuestion
	svc.HandleWSMessage(ctx, hostClient, nextQMsg)

	m, _ = svc.GetMatchByID(ctx, created.ID)
	if m.Status != match.StatusFinished {
		t.Errorf("esperava match finished, got %s", m.Status)
	}

	// Summary check
	summary, err := svc.GetMatchSummary(ctx, hostID, m.ID)
	if err != nil {
		t.Fatalf("error obtenint resum: %v", err)
	}
	if len(summary.Podium) != 1 || summary.Podium[0].PlayerID != join1.PlayerID {
		t.Errorf("podi incorrecte: %+v", summary.Podium)
	}

	// Disconnects
	svc.HandleClientDisconnect(ctx, p1Client)
	svc.HandleClientDisconnect(ctx, hostClient)
}
