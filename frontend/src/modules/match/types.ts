export type MatchStatus =
  | 'lobby'
  | 'question_preview'
  | 'question_active'
  | 'question_results'
  | 'leaderboard'
  | 'finished'

export interface CreateMatchRequest {
  quizId: string
}

export interface MatchCreatedResponse {
  id: string
  quizId: string
  quizTitle?: string
  hostId: string
  pin: string
  status: MatchStatus
  qrCodeUrl?: string
  playUrl: string
  createdAt?: string
}

export interface MatchPublicInfo {
  id: string
  pin: string
  quizTitle: string
  hostName: string
  status: MatchStatus
  playerCount: number
}

export interface JoinMatchRequest {
  nickname: string
}

export interface JoinMatchResponse {
  matchId: string
  playerId: string
  userId?: string
  nickname: string
  pin: string
  status: MatchStatus
}

export interface PlayerScoreItem {
  playerId: string
  userId?: string
  nickname: string
  score: number
  rank: number
  correctCount?: number
  totalAnswered?: number
}

export interface MatchSummaryResponse {
  matchId: string
  quizTitle: string
  totalQuestions: number
  totalPlayers: number
  podium: PlayerScoreItem[]
  leaderboard: PlayerScoreItem[]
}

export interface MatchError {
  code: string
  message: string
  details?: Record<string, any>
}

export interface MatchErrorResponse {
  error: MatchError
}

export interface MatchAnswerOption {
  id: string
  text: string
  imageUrl?: string
  isCorrect?: boolean
  count?: number
  percentage?: number
}

export interface MatchQuestion {
  id: string
  orderIndex: number
  title: string
  type: 'single_choice' | 'multiple_choice'
  timeLimitSeconds: number
  points: number
  imageUrl?: string
  options: MatchAnswerOption[]
}

export interface MatchPlayer {
  id: string
  userId?: string
  nickname: string
  score: number
  rank?: number
  isConnected: boolean
  isKicked: boolean
  joinedAt?: string
}

// Format i formes d'opcions estil Kahoot
export interface KahootOptionTheme {
  color: string
  bgHover: string
  border: string
  textColor: string
  shape: string
  shapeIcon: string
  label: string
}

export const KAHOOT_THEMES: KahootOptionTheme[] = [
  {
    color: '#e21b3c',
    bgHover: '#c41432',
    border: '#b3102c',
    textColor: '#ffffff',
    shape: 'triangle',
    shapeIcon: '▲',
    label: 'Vermell'
  },
  {
    color: '#1368ce',
    bgHover: '#0f52a5',
    border: '#0d4387',
    textColor: '#ffffff',
    shape: 'diamond',
    shapeIcon: '◆',
    label: 'Blau'
  },
  {
    color: '#d89e00',
    bgHover: '#b88600',
    border: '#9a7000',
    textColor: '#ffffff',
    shape: 'circle',
    shapeIcon: '●',
    label: 'Groc'
  },
  {
    color: '#26890c',
    bgHover: '#1f6e0a',
    border: '#185507',
    textColor: '#ffffff',
    shape: 'square',
    shapeIcon: '■',
    label: 'Verd'
  },
  {
    color: '#864cbf',
    bgHover: '#6c3a9d',
    border: '#572e7f',
    textColor: '#ffffff',
    shape: 'star',
    shapeIcon: '★',
    label: 'Lila'
  },
  {
    color: '#e65c00',
    bgHover: '#c44e00',
    border: '#a34100',
    textColor: '#ffffff',
    shape: 'hexagon',
    shapeIcon: '⬡',
    label: 'Taronja'
  }
]

// Tipus per als esdeveniments WebSocket
export type WSEventName =
  // Client -> Server
  | 'host:start_match'
  | 'host:start_question_timer'
  | 'host:show_results'
  | 'host:show_leaderboard'
  | 'host:next_question'
  | 'host:kick_player'
  | 'player:submit_answer'
  | 'ping'
  // Server -> Client
  | 'match:state'
  | 'match:player_joined'
  | 'match:player_left'
  | 'match:player_kicked'
  | 'match:question_preview'
  | 'match:question_started'
  | 'match:answer_stats'
  | 'match:question_ended'
  | 'match:leaderboard'
  | 'match:finished'
  | 'pong'
  | 'error'

export interface WSEvent<T = any> {
  event: WSEventName | string
  data?: T
}

export interface WSMatchStateData {
  matchId: string
  pin: string
  status: MatchStatus
  quizTitle?: string
  currentQuestionIndex: number
  totalQuestions: number
  players: MatchPlayer[]
  currentQuestion?: MatchQuestion
  role?: 'host' | 'player'
  playerId?: string
  score?: number
  totalPlayers?: number
}

export interface WSPlayerJoinedData {
  playerId: string
  nickname: string
  totalPlayers: number
}

export interface WSPlayerLeftData {
  playerId: string
  nickname: string
  totalPlayers: number
}

export interface WSPlayerKickedData {
  playerId: string
}

export interface WSQuestionPreviewData {
  questionIndex: number
  totalQuestions: number
  question: MatchQuestion
}

export interface WSQuestionStartedData {
  questionIndex: number
  totalQuestions: number
  timeLimitSeconds: number
  question: MatchQuestion
}

export interface WSAnswerStatsData {
  answeredCount: number
  totalPlayers: number
}

export interface WSQuestionEndedData {
  questionIndex: number
  correctOptionIds?: string[]
  optionCounts?: Record<string, number>
  isCorrect?: boolean
  scoreAwarded?: number
  totalScore?: number
}

export interface WSLeaderboardData {
  leaderboard: PlayerScoreItem[]
}

export interface WSFinishedData {
  summary?: MatchSummaryResponse
  podium?: PlayerScoreItem[]
  leaderboard?: PlayerScoreItem[]
  quizTitle?: string
  totalQuestions?: number
  totalPlayers?: number
}
