package match

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MatchStatus defines the lifecycle status of a live match.
type MatchStatus string

const (
	StatusLobby           MatchStatus = "lobby"
	StatusQuestionPreview MatchStatus = "question_preview"
	StatusQuestionActive  MatchStatus = "question_active"
	StatusQuestionResults MatchStatus = "question_results"
	StatusLeaderboard     MatchStatus = "leaderboard"
	StatusFinished        MatchStatus = "finished"
)

// Match represents a live multiplayer game session.
type Match struct {
	ID                   uuid.UUID   `json:"id"`
	QuizID               uuid.UUID   `json:"quizId"`
	HostID               uuid.UUID   `json:"hostId"`
	PIN                  string      `json:"pin"`
	Status               MatchStatus `json:"status"`
	CurrentQuestionIndex int         `json:"currentQuestionIndex"`
	QuestionStartedAt    *time.Time  `json:"questionStartedAt,omitempty"`
	CreatedAt            time.Time   `json:"createdAt"`
	UpdatedAt            time.Time   `json:"updatedAt"`
	DeletedAt            *time.Time  `json:"deletedAt,omitempty"`

	// Enriched fields from joins
	QuizTitle   string `json:"quizTitle,omitempty"`
	HostName    string `json:"hostName,omitempty"`
	PlayerCount int    `json:"playerCount,omitempty"`
}

// MatchPlayer represents a participant in a live match.
type MatchPlayer struct {
	ID          uuid.UUID `json:"id"`
	MatchID     uuid.UUID `json:"matchId"`
	UserID      uuid.UUID `json:"userId"`
	Nickname    string    `json:"nickname"`
	Score       int       `json:"score"`
	IsConnected bool      `json:"isConnected"`
	IsKicked    bool      `json:"isKicked"`
	JoinedAt    time.Time `json:"joinedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// MatchAnswer represents a recorded answer by a player for a specific question.
type MatchAnswer struct {
	ID                uuid.UUID   `json:"id"`
	MatchID           uuid.UUID   `json:"matchId"`
	QuestionID        uuid.UUID   `json:"questionId"`
	PlayerID          uuid.UUID   `json:"playerId"`
	SelectedAnswerIDs []uuid.UUID `json:"selectedAnswerIds"`
	IsCorrect         bool        `json:"isCorrect"`
	ScoreAwarded      int         `json:"scoreAwarded"`
	ResponseTimeMS    int         `json:"responseTimeMs"`
	AnsweredAt        time.Time   `json:"answeredAt"`
}

// REST DTOs

// CreateMatchRequest defines the payload required to create a new match.
type CreateMatchRequest struct {
	QuizID uuid.UUID `json:"quizId" binding:"required"`
}

// MatchCreatedResponse defines the response after successfully creating a match.
type MatchCreatedResponse struct {
	ID        uuid.UUID   `json:"id"`
	QuizID    uuid.UUID   `json:"quizId"`
	QuizTitle string      `json:"quizTitle"`
	HostID    uuid.UUID   `json:"hostId"`
	PIN       string      `json:"pin"`
	Status    MatchStatus `json:"status"`
	QRCodeURL string      `json:"qrCodeUrl"`
	PlayURL   string      `json:"playUrl"`
	CreatedAt time.Time   `json:"createdAt"`
}

// MatchPublicInfo defines the public info of a match before or during joining.
type MatchPublicInfo struct {
	ID          uuid.UUID   `json:"id"`
	PIN         string      `json:"pin"`
	QuizTitle   string      `json:"quizTitle"`
	HostName    string      `json:"hostName"`
	Status      MatchStatus `json:"status"`
	PlayerCount int         `json:"playerCount"`
}

// JoinMatchRequest defines the payload to join a match as a player.
type JoinMatchRequest struct {
	Nickname string `json:"nickname" binding:"required,min=2,max=30"`
}

// JoinMatchResponse defines the response after a player joins.
type JoinMatchResponse struct {
	MatchID  uuid.UUID   `json:"matchId"`
	PlayerID uuid.UUID   `json:"playerId"`
	UserID   uuid.UUID   `json:"userId"`
	Nickname string      `json:"nickname"`
	PIN      string      `json:"pin"`
	Status   MatchStatus `json:"status"`
}

// PlayerScoreItem defines a player's rank and score in a match.
type PlayerScoreItem struct {
	PlayerID      uuid.UUID `json:"playerId"`
	UserID        uuid.UUID `json:"userId"`
	Nickname      string    `json:"nickname"`
	Score         int       `json:"score"`
	Rank          int       `json:"rank"`
	CorrectCount  int       `json:"correctCount"`
	TotalAnswered int       `json:"totalAnswered"`
}

// MatchSummaryResponse defines the final podium and statistics.
type MatchSummaryResponse struct {
	MatchID        uuid.UUID         `json:"matchId"`
	QuizTitle      string            `json:"quizTitle"`
	TotalQuestions int               `json:"totalQuestions"`
	TotalPlayers   int               `json:"totalPlayers"`
	Podium         []PlayerScoreItem `json:"podium"`
	Leaderboard    []PlayerScoreItem `json:"leaderboard"`
}

// WebSocket Protocol constants & payloads

const (
	// Host -> Server events
	HostEventStartMatch         = "host:start_match"
	HostEventStartQuestionTimer = "host:start_question_timer"
	HostEventShowResults        = "host:show_results"
	HostEventShowLeaderboard    = "host:show_leaderboard"
	HostEventNextQuestion       = "host:next_question"
	HostEventKickPlayer         = "host:kick_player"

	// Player -> Server events
	PlayerEventSubmitAnswer = "player:submit_answer"

	// Server -> Client events
	ServerEventMatchState      = "match:state"
	ServerEventPlayerJoined    = "match:player_joined"
	ServerEventPlayerLeft      = "match:player_left"
	ServerEventPlayerKicked    = "match:player_kicked"
	ServerEventQuestionPreview = "match:question_preview"
	ServerEventQuestionStarted = "match:question_started"
	ServerEventAnswerStats     = "match:answer_stats"
	ServerEventQuestionEnded   = "match:question_ended"
	ServerEventLeaderboard     = "match:leaderboard"
	ServerEventFinished        = "match:finished"
	ServerEventError           = "match:error"
)

// WSMessage represents a standard JSON WebSocket envelope.
type WSMessage struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// OutgoingWSMessage is a helper to construct WS messages for serialization.
type OutgoingWSMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// KickPlayerPayload represents the payload for host:kick_player.
type KickPlayerPayload struct {
	PlayerID uuid.UUID `json:"playerId"`
}

// SubmitAnswerPayload represents the payload for player:submit_answer.
type SubmitAnswerPayload struct {
	QuestionID uuid.UUID   `json:"questionId"`
	AnswerIDs  []uuid.UUID `json:"answerIds"`
}

// MatchStatePayload represents the state synchronization payload.
type MatchStatePayload struct {
	MatchID              uuid.UUID               `json:"matchId"`
	PIN                  string                  `json:"pin"`
	QuizTitle            string                  `json:"quizTitle"`
	Status               MatchStatus             `json:"status"`
	CurrentQuestionIndex int                     `json:"currentQuestionIndex"`
	TotalQuestions       int                     `json:"totalQuestions"`
	Role                 string                  `json:"role"` // "host" or "player"
	Player               *MatchPlayer            `json:"player,omitempty"`
	Players              []MatchPlayer           `json:"players,omitempty"`
	CurrentQuestion      *QuestionPreviewPayload `json:"currentQuestion,omitempty"`
}

// PlayerJoinedPayload represents match:player_joined broadcast.
type PlayerJoinedPayload struct {
	PlayerID     uuid.UUID `json:"playerId"`
	Nickname     string    `json:"nickname"`
	TotalPlayers int       `json:"totalPlayers"`
}

// PlayerLeftPayload represents match:player_left broadcast.
type PlayerLeftPayload struct {
	PlayerID     uuid.UUID `json:"playerId"`
	Nickname     string    `json:"nickname"`
	TotalPlayers int       `json:"totalPlayers"`
}

// PlayerKickedPayload represents match:player_kicked broadcast.
type PlayerKickedPayload struct {
	PlayerID uuid.UUID `json:"playerId"`
}

// QuestionOptionPayload represents an option shown on question_started.
type QuestionOptionPayload struct {
	ID         uuid.UUID `json:"id"`
	Text       string    `json:"text"`
	OrderIndex int       `json:"orderIndex"`
}

// QuestionPreviewPayload represents match:question_preview broadcast.
type QuestionPreviewPayload struct {
	QuestionIndex    int       `json:"questionIndex"`
	TotalQuestions   int       `json:"totalQuestions"`
	QuestionID       uuid.UUID `json:"questionId"`
	Text             string    `json:"text"`
	ImageURL         *string   `json:"imageUrl,omitempty"`
	QuestionType     string    `json:"questionType"`
	TimeLimitSeconds int       `json:"timeLimitSeconds"`
}

// QuestionStartedPayload represents match:question_started broadcast.
type QuestionStartedPayload struct {
	QuestionIndex    int                     `json:"questionIndex"`
	TotalQuestions   int                     `json:"totalQuestions"`
	QuestionID       uuid.UUID               `json:"questionId"`
	Text             string                  `json:"text"`
	ImageURL         *string                 `json:"imageUrl,omitempty"`
	QuestionType     string                  `json:"questionType"`
	TimeLimitSeconds int                     `json:"timeLimitSeconds"`
	StartedAt        time.Time               `json:"startedAt"`
	Options          []QuestionOptionPayload `json:"options"`
}

// AnswerStatsPayload represents match:answer_stats sent to host.
type AnswerStatsPayload struct {
	AnsweredCount int `json:"answeredCount"`
	TotalPlayers  int `json:"totalPlayers"`
}

// OptionResultCount counts answers per choice.
type OptionResultCount struct {
	OptionID  uuid.UUID `json:"optionId"`
	Text      string    `json:"text"`
	Count     int       `json:"count"`
	IsCorrect bool      `json:"isCorrect"`
}

// QuestionEndedHostPayload represents match:question_ended for the host.
type QuestionEndedHostPayload struct {
	QuestionID       uuid.UUID           `json:"questionId"`
	TotalAnswered    int                 `json:"totalAnswered"`
	TotalPlayers     int                 `json:"totalPlayers"`
	CorrectAnswerIDs []uuid.UUID         `json:"correctAnswerIds"`
	OptionCounts     []OptionResultCount `json:"optionCounts"`
}

// QuestionEndedPlayerPayload represents match:question_ended for an individual player.
type QuestionEndedPlayerPayload struct {
	QuestionID       uuid.UUID   `json:"questionId"`
	IsCorrect        bool        `json:"isCorrect"`
	ScoreAwarded     int         `json:"scoreAwarded"`
	TotalScore       int         `json:"totalScore"`
	CorrectAnswerIDs []uuid.UUID `json:"correctAnswerIds"`
}

// LeaderboardPayload represents match:leaderboard broadcast.
type LeaderboardPayload struct {
	QuestionIndex  int               `json:"questionIndex"`
	TotalQuestions int               `json:"totalQuestions"`
	Items          []PlayerScoreItem `json:"items"`
}

// ErrorPayload represents match:error sent to a client.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
