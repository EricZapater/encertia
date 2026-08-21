import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  MatchStatus,
  MatchPlayer,
  MatchQuestion,
  MatchAnswerOption,
  PlayerScoreItem,
  MatchSummaryResponse,
  MatchCreatedResponse,
  WSPlayerJoinedData,
  WSPlayerLeftData,
  WSPlayerKickedData,
  WSAnswerStatsData,
  WSLeaderboardData,
  WSFinishedData
} from './types'
import * as matchApi from './api'
import { MatchWSClient } from './wsClient'

export const useMatchStore = defineStore('match', () => {
  // State
  const matchId = ref<string | null>(null)
  const pin = ref<string | null>(null)
  const status = ref<MatchStatus | null>(null)
  const quizTitle = ref<string>('')
  const role = ref<'host' | 'player' | null>(null)
  const myPlayerId = ref<string | null>(null)
  const myNickname = ref<string>('')
  const qrCodeUrl = ref<string | null>(null)
  const playUrl = ref<string | null>(null)

  const players = ref<MatchPlayer[]>([])
  const currentQuestionIndex = ref<number>(0)
  const totalQuestions = ref<number>(0)
  const currentQuestion = ref<MatchQuestion | null>(null)

  const timerSeconds = ref<number>(0)
  const initialTimeLimit = ref<number>(0)
  let timerInterval: ReturnType<typeof setInterval> | null = null

  const answeredCount = ref<number>(0)
  const totalPlayers = ref<number>(0)

  // Estat de resposta del jugador
  const mySelectedAnswerIds = ref<string[]>([])
  const hasSubmittedAnswer = ref<boolean>(false)
  const lastAnswerResult = ref<{
    isCorrect: boolean
    scoreAwarded: number
    totalScore: number
    correctOptionIds?: string[]
    optionCounts?: Record<string, number>
  } | null>(null)
  const myScore = ref<number>(0)

  // Rànquings i resultats
  const leaderboard = ref<PlayerScoreItem[]>([])
  const podium = ref<PlayerScoreItem[]>([])
  const summary = ref<MatchSummaryResponse | null>(null)

  // Connexió i control d'estat
  const isConnected = ref<boolean>(false)
  const isKicked = ref<boolean>(false)
  const error = ref<string | null>(null)
  const isLoading = ref<boolean>(false)

  // Client WebSocket actiu
  let wsClient: MatchWSClient | null = null

  // Getters
  const isHost = computed(() => role.value === 'host')
  const isPlayer = computed(() => role.value === 'player')
  const activePlayersCount = computed(() => players.value.filter((p) => p.isConnected && !p.isKicked).length)
  const questionNumber = computed(() => currentQuestionIndex.value + 1)
  const isLastQuestion = computed(() => totalQuestions.value > 0 && questionNumber.value >= totalQuestions.value)
  const timerPercentage = computed(() => {
    if (initialTimeLimit.value <= 0) return 0
    return Math.max(0, Math.min(100, (timerSeconds.value / initialTimeLimit.value) * 100))
  })

  // --- Gestió del Temporitzador ---
  function startLocalTimer(duration: number) {
    stopLocalTimer()
    initialTimeLimit.value = duration
    timerSeconds.value = duration

    timerInterval = setInterval(() => {
      if (timerSeconds.value > 0) {
        timerSeconds.value--
      } else {
        stopLocalTimer()
      }
    }, 1000)
  }

  function stopLocalTimer() {
    if (timerInterval) {
      clearInterval(timerInterval)
      timerInterval = null
    }
  }

  // --- Handlers d'Esdeveniments WebSocket ---
  function setupWebSocketListeners(client: MatchWSClient) {
    client.on('open', () => {
      isConnected.value = true
      error.value = null
    })

    client.on('close', () => {
      isConnected.value = false
    })

    client.on('error', (err) => {
      console.warn('[MatchStore] Error WebSocket:', err)
    })

    client.on('reconnect_failed', () => {
      error.value = 'No es pot connectar amb el servidor de la partida.'
    })

    function parseQuestionPayload(raw: any, existing?: MatchQuestion | null): MatchQuestion {
      if (!raw) return existing || (null as any)
      const qObj = raw.question || raw
      const rawOptions = raw.options && raw.options.length > 0 ? raw.options : qObj.options && qObj.options.length > 0 ? qObj.options : null
      const optionsSrc = rawOptions || existing?.options || []

      const options: MatchAnswerOption[] = optionsSrc.map((opt: any) => ({
        id: opt.id || opt.optionId || '',
        text: opt.text || '',
        orderIndex: opt.orderIndex ?? 0,
        imageUrl: opt.imageUrl,
        isCorrect: opt.isCorrect,
        count: opt.count,
        percentage: opt.percentage
      }))

      return {
        id: raw.questionId || raw.id || qObj.questionId || qObj.id || existing?.id || '',
        orderIndex: raw.questionIndex ?? raw.orderIndex ?? qObj.questionIndex ?? qObj.orderIndex ?? existing?.orderIndex ?? 0,
        title: raw.text || raw.title || qObj.text || qObj.title || existing?.title || '',
        type: (raw.questionType || raw.type || qObj.questionType || qObj.type || existing?.type || 'single_choice') as any,
        timeLimitSeconds: raw.timeLimitSeconds || qObj.timeLimitSeconds || existing?.timeLimitSeconds || 20,
        points: raw.points || qObj.points || existing?.points || 1,
        imageUrl: raw.imageUrl || qObj.imageUrl || existing?.imageUrl,
        options: options
      }
    }

    client.on<any>('match:state', (data) => {
      if (!data) return
      if (data.matchId) matchId.value = data.matchId
      if (data.pin) pin.value = data.pin
      if (data.status) status.value = data.status
      if (data.quizTitle) quizTitle.value = data.quizTitle
      if (data.role) role.value = data.role
      if (data.playerId) myPlayerId.value = data.playerId
      if (typeof data.score === 'number') myScore.value = data.score
      if (data.players) players.value = data.players
      if (typeof data.currentQuestionIndex === 'number') currentQuestionIndex.value = data.currentQuestionIndex
      if (typeof data.totalQuestions === 'number') totalQuestions.value = data.totalQuestions
      if (data.currentQuestion) {
        currentQuestion.value = parseQuestionPayload(data.currentQuestion, currentQuestion.value)
      }
      if (typeof data.totalPlayers === 'number') totalPlayers.value = data.totalPlayers
    })

    client.on<WSPlayerJoinedData>('match:player_joined', (data) => {
      if (!data) return
      totalPlayers.value = data.totalPlayers
      const existing = players.value.find((p) => p.id === data.playerId)
      if (existing) {
        existing.isConnected = true
        existing.isKicked = false
        existing.nickname = data.nickname
      } else {
        players.value.push({
          id: data.playerId,
          nickname: data.nickname,
          score: 0,
          isConnected: true,
          isKicked: false
        })
      }
    })

    client.on<WSPlayerLeftData>('match:player_left', (data) => {
      if (!data) return
      totalPlayers.value = data.totalPlayers
      const existing = players.value.find((p) => p.id === data.playerId)
      if (existing) {
        existing.isConnected = false
      }
    })

    client.on<WSPlayerKickedData>('match:player_kicked', (data) => {
      if (!data) return
      if (myPlayerId.value && data.playerId === myPlayerId.value) {
        isKicked.value = true
        leaveMatch()
        return
      }
      const existing = players.value.find((p) => p.id === data.playerId)
      if (existing) {
        existing.isKicked = true
        existing.isConnected = false
      }
    })

    client.on<any>('match:question_preview', (data) => {
      if (!data) return
      status.value = 'question_preview'
      if (typeof data.questionIndex === 'number') currentQuestionIndex.value = data.questionIndex
      if (typeof data.totalQuestions === 'number') totalQuestions.value = data.totalQuestions
      currentQuestion.value = parseQuestionPayload(data, currentQuestion.value)

      // Reinicia estat de resposta per a la nova pregunta
      mySelectedAnswerIds.value = []
      hasSubmittedAnswer.value = false
      lastAnswerResult.value = null
      answeredCount.value = 0
      stopLocalTimer()
      timerSeconds.value = currentQuestion.value?.timeLimitSeconds || data.timeLimitSeconds || 0
      initialTimeLimit.value = timerSeconds.value
    })

    client.on<any>('match:question_started', (data) => {
      if (!data) return
      status.value = 'question_active'
      if (typeof data.questionIndex === 'number') currentQuestionIndex.value = data.questionIndex
      if (typeof data.totalQuestions === 'number') totalQuestions.value = data.totalQuestions
      currentQuestion.value = parseQuestionPayload(data, currentQuestion.value)
      answeredCount.value = 0

      const seconds = data.timeLimitSeconds || currentQuestion.value?.timeLimitSeconds || 20
      startLocalTimer(seconds)
    })

    client.on<WSAnswerStatsData>('match:answer_stats', (data) => {
      if (!data) return
      answeredCount.value = data.answeredCount
      if (data.totalPlayers) totalPlayers.value = data.totalPlayers
    })

    client.on<any>('match:question_ended', (data) => {
      if (!data) return
      status.value = 'question_results'
      stopLocalTimer()

      const correctOptionIds: string[] = data.correctAnswerIDs || data.correctOptionIds || []
      const optionCounts = data.optionCounts

      if (currentQuestion.value) {
        currentQuestion.value.options.forEach((opt) => {
          if (correctOptionIds.length > 0) {
            opt.isCorrect = correctOptionIds.includes(opt.id)
          }
          if (Array.isArray(optionCounts)) {
            const matchCount = optionCounts.find((item: any) => item.optionId === opt.id || item.id === opt.id)
            if (matchCount) {
              opt.count = matchCount.count
              if (matchCount.isCorrect !== undefined) opt.isCorrect = matchCount.isCorrect
            }
          } else if (optionCounts && typeof optionCounts[opt.id] === 'number') {
            opt.count = optionCounts[opt.id]
          }
        })
      }

      if (data.isCorrect !== undefined || data.scoreAwarded !== undefined || data.totalScore !== undefined) {
        lastAnswerResult.value = {
          isCorrect: Boolean(data.isCorrect),
          scoreAwarded: data.scoreAwarded ?? (data.isCorrect ? 1 : 0),
          totalScore: data.totalScore ?? myScore.value,
          correctOptionIds: correctOptionIds,
          optionCounts: optionCounts
        }
        if (typeof data.totalScore === 'number') {
          myScore.value = data.totalScore
        }
      }
    })

    client.on<WSLeaderboardData>('match:leaderboard', (data) => {
      if (!data) return
      status.value = 'leaderboard'
      leaderboard.value = data.leaderboard || []
    })

    client.on<WSFinishedData>('match:finished', (data) => {
      if (!data) return
      status.value = 'finished'
      stopLocalTimer()

      if (data.summary) {
        summary.value = data.summary
        podium.value = data.summary.podium || []
        leaderboard.value = data.summary.leaderboard || []
      } else {
        podium.value = data.podium || []
        leaderboard.value = data.leaderboard || []
      }
    })

    client.on('error', (errData: any) => {
      if (errData?.message) {
        error.value = errData.message
      }
    })
  }

  // --- Actions del Moderador (Host) ---

  async function initHostMatch(quizId: string): Promise<MatchCreatedResponse> {
    isLoading.value = true
    error.value = null
    try {
      const res = await matchApi.createMatch({ quizId })
      matchId.value = res.id
      pin.value = res.pin
      status.value = res.status
      quizTitle.value = res.quizTitle || 'Partida en directe'
      qrCodeUrl.value = res.qrCodeUrl || null
      playUrl.value = res.playUrl
      role.value = 'host'

      // Connecta WebSocket immediatament com a Host
      await connectAsHost(res.pin)
      return res
    } catch (err: any) {
      const msg = err.response?.data?.error?.message || err.message || 'Error en crear la partida.'
      error.value = msg
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function connectAsHost(targetPin: string): Promise<void> {
    pin.value = targetPin
    role.value = 'host'

    if (wsClient) {
      wsClient.disconnect()
    }

    wsClient = new MatchWSClient({ pin: targetPin })
    setupWebSocketListeners(wsClient)
    await wsClient.connect()
  }

  function startMatch(): boolean {
    if (!wsClient || !isHost.value) return false
    return wsClient.send('host:start_match', {})
  }

  function startQuestionTimer(): boolean {
    if (!wsClient || !isHost.value) return false
    return wsClient.send('host:start_question_timer', {})
  }

  function showResults(): boolean {
    if (!wsClient || !isHost.value) return false
    return wsClient.send('host:show_results', {})
  }

  function showLeaderboard(): boolean {
    if (!wsClient || !isHost.value) return false
    return wsClient.send('host:show_leaderboard', {})
  }

  function nextQuestion(): boolean {
    if (!wsClient || !isHost.value) return false
    return wsClient.send('host:next_question', {})
  }

  function kickPlayer(playerId: string): boolean {
    if (!wsClient || !isHost.value) return false
    return wsClient.send('host:kick_player', { playerId })
  }

  // --- Actions del Jugador (Player) ---

  async function joinAndConnectAsPlayer(targetPin: string, nickname: string): Promise<void> {
    isLoading.value = true
    error.value = null
    isKicked.value = false
    try {
      const res = await matchApi.joinMatch(targetPin, { nickname })
      matchId.value = res.matchId
      pin.value = res.pin
      status.value = res.status
      myPlayerId.value = res.playerId
      myNickname.value = res.nickname
      role.value = 'player'

      // Connectar al WebSocket com a jugador
      await connectAsPlayer(targetPin)
    } catch (err: any) {
      const msg = err.response?.data?.error?.message || err.message || 'Error en unir-se a la partida.'
      error.value = msg
      throw err
    } finally {
      isLoading.value = false
    }
  }

  async function connectAsPlayer(targetPin: string): Promise<void> {
    pin.value = targetPin
    role.value = 'player'

    if (wsClient) {
      wsClient.disconnect()
    }

    wsClient = new MatchWSClient({ pin: targetPin })
    setupWebSocketListeners(wsClient)
    await wsClient.connect()
  }

  function selectAnswer(answerId: string, isMultiple = false) {
    if (hasSubmittedAnswer.value || status.value !== 'question_active') return

    if (isMultiple) {
      const idx = mySelectedAnswerIds.value.indexOf(answerId)
      if (idx > -1) {
        mySelectedAnswerIds.value.splice(idx, 1)
      } else {
        mySelectedAnswerIds.value.push(answerId)
      }
    } else {
      mySelectedAnswerIds.value = [answerId]
    }
  }

  function submitAnswer(): boolean {
    if (
      !wsClient ||
      !isPlayer.value ||
      hasSubmittedAnswer.value ||
      status.value !== 'question_active' ||
      !currentQuestion.value ||
      mySelectedAnswerIds.value.length === 0
    ) {
      return false
    }

    const ok = wsClient.send('player:submit_answer', {
      questionId: currentQuestion.value.id,
      answerIds: mySelectedAnswerIds.value
    })

    if (ok) {
      hasSubmittedAnswer.value = true
    }
    return ok
  }

  async function fetchSummary(targetMatchId?: string): Promise<MatchSummaryResponse | null> {
    const id = targetMatchId || matchId.value
    if (!id) return null

    isLoading.value = true
    try {
      const res = await matchApi.getMatchSummary(id)
      summary.value = res
      podium.value = res.podium
      leaderboard.value = res.leaderboard
      return res
    } catch (err: any) {
      const msg = err.response?.data?.error?.message || err.message || 'Error en carregar el resum.'
      error.value = msg
      return null
    } finally {
      isLoading.value = false
    }
  }

  function leaveMatch() {
    stopLocalTimer()
    if (wsClient) {
      wsClient.disconnect()
      wsClient = null
    }
    matchId.value = null
    pin.value = null
    status.value = null
    quizTitle.value = ''
    role.value = null
    myPlayerId.value = null
    players.value = []
    currentQuestionIndex.value = 0
    totalQuestions.value = 0
    currentQuestion.value = null
    timerSeconds.value = 0
    initialTimeLimit.value = 0
    answeredCount.value = 0
    totalPlayers.value = 0
    mySelectedAnswerIds.value = []
    hasSubmittedAnswer.value = false
    lastAnswerResult.value = null
    myScore.value = 0
    leaderboard.value = []
    podium.value = []
    summary.value = null
    isConnected.value = false
    error.value = null
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    matchId,
    pin,
    status,
    quizTitle,
    role,
    myPlayerId,
    myNickname,
    qrCodeUrl,
    playUrl,
    players,
    currentQuestionIndex,
    totalQuestions,
    currentQuestion,
    timerSeconds,
    initialTimeLimit,
    answeredCount,
    totalPlayers,
    mySelectedAnswerIds,
    hasSubmittedAnswer,
    lastAnswerResult,
    myScore,
    leaderboard,
    podium,
    summary,
    isConnected,
    isKicked,
    error,
    isLoading,
    // Getters
    isHost,
    isPlayer,
    activePlayersCount,
    questionNumber,
    isLastQuestion,
    timerPercentage,
    // Host Actions
    initHostMatch,
    connectAsHost,
    startMatch,
    startQuestionTimer,
    showResults,
    showLeaderboard,
    nextQuestion,
    kickPlayer,
    // Player Actions
    joinAndConnectAsPlayer,
    connectAsPlayer,
    selectAnswer,
    submitAnswer,
    // General Actions
    fetchSummary,
    leaveMatch,
    clearError,
    // Helper exposat per a tests
    _setupWebSocketListeners: setupWebSocketListeners
  }
})
