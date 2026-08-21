import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useMatchStore } from '../store'
import * as matchApi from '../api'

vi.mock('../api', () => ({
  createMatch: vi.fn(),
  getMatchByPin: vi.fn(),
  joinMatch: vi.fn(),
  getMatchSummary: vi.fn()
}))

// Mock de MatchWSClient com a classe (constructor)
vi.mock('../wsClient', () => {
  class MockMatchWSClient {
    listeners: Record<string, Function[]> = {}
    isConnected = true
    connect = vi.fn().mockResolvedValue(undefined)
    disconnect = vi.fn()
    send = vi.fn().mockReturnValue(true)
    on = vi.fn().mockImplementation((evt: string, fn: Function) => {
      if (!this.listeners[evt]) this.listeners[evt] = []
      this.listeners[evt].push(fn)
      return () => {}
    })
    off = vi.fn()
  }
  return {
    MatchWSClient: MockMatchWSClient
  }
})

describe('useMatchStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('estat inicial correcte', () => {
    const store = useMatchStore()
    expect(store.matchId).toBeNull()
    expect(store.pin).toBeNull()
    expect(store.status).toBeNull()
    expect(store.role).toBeNull()
    expect(store.players).toEqual([])
    expect(store.myScore).toBe(0)
    expect(store.isConnected).toBe(false)
  })

  it('initHostMatch crea la partida i configura el rol de host', async () => {
    const store = useMatchStore()
    vi.mocked(matchApi.createMatch).mockResolvedValueOnce({
      id: 'm123',
      quizId: 'q456',
      quizTitle: 'Història de Catalunya',
      hostId: 'u1',
      pin: '987654',
      status: 'lobby',
      playUrl: 'https://encertia.ericzapater.cat/play?pin=987654'
    })

    const res = await store.initHostMatch('q456')

    expect(matchApi.createMatch).toHaveBeenCalledWith({ quizId: 'q456' })
    expect(store.matchId).toBe('m123')
    expect(store.pin).toBe('987654')
    expect(store.quizTitle).toBe('Història de Catalunya')
    expect(store.isHost).toBe(true)
    expect(res.pin).toBe('987654')
  })

  it('joinAndConnectAsPlayer registra el jugador i assigna rol i nickname', async () => {
    const store = useMatchStore()
    vi.mocked(matchApi.joinMatch).mockResolvedValueOnce({
      matchId: 'm123',
      playerId: 'p1',
      nickname: 'Berta',
      pin: '987654',
      status: 'lobby'
    })

    await store.joinAndConnectAsPlayer('987654', 'Berta')

    expect(matchApi.joinMatch).toHaveBeenCalledWith('987654', { nickname: 'Berta' })
    expect(store.pin).toBe('987654')
    expect(store.myNickname).toBe('Berta')
    expect(store.myPlayerId).toBe('p1')
    expect(store.isPlayer).toBe(true)
  })

  it('selectAnswer gestiona selecció única i selecció múltiple', () => {
    const store = useMatchStore()
    store.status = 'question_active'

    // Selecció única
    store.selectAnswer('opt-1', false)
    expect(store.mySelectedAnswerIds).toEqual(['opt-1'])

    store.selectAnswer('opt-2', false)
    expect(store.mySelectedAnswerIds).toEqual(['opt-2'])

    // Selecció múltiple (toggle)
    store.selectAnswer('opt-1', true)
    expect(store.mySelectedAnswerIds).toContain('opt-1')
    expect(store.mySelectedAnswerIds).toContain('opt-2')

    store.selectAnswer('opt-1', true)
    expect(store.mySelectedAnswerIds).not.toContain('opt-1')
    expect(store.mySelectedAnswerIds).toContain('opt-2')
  })

  it('submitAnswer canvia hasSubmittedAnswer a true', async () => {
    const store = useMatchStore()
    store.role = 'player'
    store.status = 'question_active'
    store.currentQuestion = {
      id: 'q1',
      orderIndex: 0,
      title: 'Pregunta 1',
      type: 'single_choice',
      timeLimitSeconds: 20,
      points: 1,
      options: [{ id: 'opt-1', text: 'Opció A' }]
    }

    await store.connectAsPlayer('123456')

    store.selectAnswer('opt-1', false)
    const success = store.submitAnswer()

    expect(success).toBe(true)
    expect(store.hasSubmittedAnswer).toBe(true)
  })

  it('leaveMatch restableix l’estat complet', () => {
    const store = useMatchStore()
    store.matchId = 'm123'
    store.pin = '123456'
    store.role = 'host'
    store.players = [{ id: 'p1', nickname: 'Test', score: 10, isConnected: true, isKicked: false }]

    store.leaveMatch()

    expect(store.matchId).toBeNull()
    expect(store.pin).toBeNull()
    expect(store.role).toBeNull()
    expect(store.players).toEqual([])
  })
})
