import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { MatchWSClient } from '../wsClient'

class MockWebSocket {
  static instances: MockWebSocket[] = []
  url: string
  readyState = WebSocket.CONNECTING
  onopen: ((ev: any) => void) | null = null
  onmessage: ((ev: any) => void) | null = null
  onerror: ((ev: any) => void) | null = null
  onclose: ((ev: any) => void) | null = null
  sentMessages: string[] = []

  constructor(url: string) {
    this.url = url
    MockWebSocket.instances.push(this)
    setTimeout(() => {
      this.readyState = WebSocket.OPEN
      if (this.onopen) this.onopen(new Event('open'))
    }, 10)
  }

  send(data: string) {
    this.sentMessages.push(data)
  }

  close() {
    this.readyState = WebSocket.CLOSED
    if (this.onclose) this.onclose(new CloseEvent('close'))
  }

  // Helper per simular missatges rebuts des del servidor
  simulateMessage(payload: any) {
    if (this.onmessage) {
      this.onmessage({ data: JSON.stringify(payload) } as MessageEvent)
    }
  }
}

describe('MatchWSClient', () => {
  beforeEach(() => {
    MockWebSocket.instances = []
    vi.stubGlobal('WebSocket', MockWebSocket)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('connecta correctament i construeix la URL amb el PIN i token', async () => {
    const client = new MatchWSClient({
      pin: '123456',
      token: 'fake-jwt-token'
    })

    await client.connect()
    expect(client.isConnected).toBe(true)

    const wsInstance = MockWebSocket.instances[0]
    expect(wsInstance.url).toContain('/api/ws/match/123456?token=fake-jwt-token')
  })

  it('envia missatges JSON correctament amb send()', async () => {
    const client = new MatchWSClient({ pin: '123456', token: 'token' })
    await client.connect()

    const sent = client.send('host:start_match', { foo: 'bar' })
    expect(sent).toBe(true)

    const wsInstance = MockWebSocket.instances[0]
    expect(wsInstance.sentMessages).toHaveLength(1)
    const parsed = JSON.parse(wsInstance.sentMessages[0])
    expect(parsed.event).toBe('host:start_match')
    expect(parsed.data).toEqual({ foo: 'bar' })
  })

  it('notifica els handlers registrats amb on() quan arriba un esdeveniment', async () => {
    const client = new MatchWSClient({ pin: '123456', token: 'token' })
    await client.connect()

    const handler = vi.fn()
    client.on('match:question_preview', handler)

    const wsInstance = MockWebSocket.instances[0]
    wsInstance.simulateMessage({
      event: 'match:question_preview',
      data: { questionIndex: 0, totalQuestions: 5 }
    })

    expect(handler).toHaveBeenCalledWith({ questionIndex: 0, totalQuestions: 5 })
  })

  it('permet cancel·lar subscripcions amb off() o la funció de retorn', async () => {
    const client = new MatchWSClient({ pin: '123456', token: 'token' })
    await client.connect()

    const handler = vi.fn()
    const unsubscribe = client.on('match:player_joined', handler)

    unsubscribe()

    const wsInstance = MockWebSocket.instances[0]
    wsInstance.simulateMessage({
      event: 'match:player_joined',
      data: { playerId: 'p1', nickname: 'Test', totalPlayers: 1 }
    })

    expect(handler).not.toHaveBeenCalled()
  })

  it('desconnecta netament amb disconnect()', async () => {
    const client = new MatchWSClient({ pin: '123456', token: 'token' })
    await client.connect()

    expect(client.isConnected).toBe(true)
    client.disconnect()

    expect(client.isConnected).toBe(false)
  })
})
