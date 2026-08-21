import { ACCESS_TOKEN_KEY } from '@/api/client'
import type { WSEvent, WSEventName } from './types'

export type WSEventHandler<T = any> = (data: T) => void

export interface WSClientOptions {
  pin: string
  token?: string
  autoReconnect?: boolean
  maxReconnectAttempts?: number
  heartbeatIntervalMs?: number
}

export class MatchWSClient {
  private pin: string
  private token: string | null = null
  private ws: WebSocket | null = null
  private listeners: Map<string, Set<WSEventHandler>> = new Map()
  private autoReconnect: boolean
  private maxReconnectAttempts: number
  private reconnectAttempts = 0
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null
  private heartbeatInterval: ReturnType<typeof setInterval> | null = null
  private heartbeatIntervalMs: number
  private isExplicitlyClosed = false
  private connectionPromise: Promise<void> | null = null

  constructor(options: WSClientOptions) {
    this.pin = options.pin
    this.token = options.token || (typeof window !== 'undefined' ? localStorage.getItem(ACCESS_TOKEN_KEY) : null)
    this.autoReconnect = options.autoReconnect ?? true
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 5
    this.heartbeatIntervalMs = options.heartbeatIntervalMs ?? 20000
  }

  /**
   * Construeix la URL WebSocket amb token JWT i PIN.
   */
  private buildWSUrl(): string {
    const rawWsUrl = import.meta.env.VITE_WS_URL
    let baseUrl: string

    if (rawWsUrl) {
      baseUrl = rawWsUrl.replace(/\/$/, '')
    } else if (typeof window !== 'undefined') {
      const isHttps = window.location.protocol === 'https:'
      const wsProto = isHttps ? 'wss:' : 'ws:'
      const host = window.location.host
      baseUrl = `${wsProto}//${host}`
    } else {
      baseUrl = 'ws://localhost:8080'
    }

    const token = this.token || (typeof window !== 'undefined' ? localStorage.getItem(ACCESS_TOKEN_KEY) || '' : '')
    const cleanPin = encodeURIComponent(this.pin)
    const cleanToken = encodeURIComponent(token)

    return `${baseUrl}/api/ws/match/${cleanPin}?token=${cleanToken}`
  }

  /**
   * Obre la connexió WebSocket.
   */
  public connect(): Promise<void> {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return this.connectionPromise || Promise.resolve()
    }

    this.isExplicitlyClosed = false

    this.connectionPromise = new Promise<void>((resolve, reject) => {
      try {
        const url = this.buildWSUrl()
        this.ws = new WebSocket(url)

        this.ws.onopen = () => {
          this.reconnectAttempts = 0
          this.startHeartbeat()
          this.dispatchInternal('open', { pin: this.pin })
          resolve()
        }

        this.ws.onmessage = (event: MessageEvent) => {
          this.handleMessage(event.data)
        }

        this.ws.onerror = (err) => {
          this.dispatchInternal('error', { error: err })
        }

        this.ws.onclose = (event: CloseEvent) => {
          this.stopHeartbeat()
          this.dispatchInternal('close', { code: event.code, reason: event.reason })

          if (!this.isExplicitlyClosed && this.autoReconnect) {
            this.scheduleReconnect()
          }
        }
      } catch (err) {
        reject(err)
      }
    })

    return this.connectionPromise
  }

  /**
   * Tanca la connexió WebSocket de forma explícita.
   */
  public disconnect(): void {
    this.isExplicitlyClosed = true
    this.stopHeartbeat()

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }

    if (this.ws) {
      // Treure listeners per evitar reconnect
      this.ws.onclose = null
      this.ws.onerror = null
      this.ws.onmessage = null
      this.ws.onopen = null
      this.ws.close()
      this.ws = null
    }

    this.connectionPromise = null
    this.dispatchInternal('disconnected', {})
  }

  /**
   * Envia un esdeveniment al servidor en format JSON.
   */
  public send<T = any>(event: WSEventName | string, data?: T): boolean {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn(`[MatchWSClient] No es pot enviar "${event}": WebSocket no connectat.`)
      return false
    }

    const payload: WSEvent<T> = {
      event,
      data: data || ({} as T)
    }

    try {
      this.ws.send(JSON.stringify(payload))
      return true
    } catch (err) {
      console.error(`[MatchWSClient] Error enviant esdeveniment "${event}":`, err)
      return false
    }
  }

  /**
   * Registra un oïdor per a un esdeveniment concret.
   */
  public on<T = any>(event: WSEventName | string, handler: WSEventHandler<T>): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(handler)

    // Retorna funció per desubscriure's fàcilment
    return () => this.off(event, handler)
  }

  /**
   * Desregistra un oïdor d'esdeveniment.
   */
  public off<T = any>(event: WSEventName | string, handler: WSEventHandler<T>): void {
    const handlers = this.listeners.get(event)
    if (handlers) {
      handlers.delete(handler)
      if (handlers.size === 0) {
        this.listeners.delete(event)
      }
    }
  }

  /**
   * Elimina tots els oïdors d'un o de tots els esdeveniments.
   */
  public clearListeners(event?: string): void {
    if (event) {
      this.listeners.delete(event)
    } else {
      this.listeners.clear()
    }
  }

  /**
   * Retorna true si el socket està obert.
   */
  public get isConnected(): boolean {
    return Boolean(this.ws && this.ws.readyState === WebSocket.OPEN)
  }

  // --- Mètodes interns ---

  private handleMessage(rawData: any): void {
    try {
      let parsed: WSEvent
      if (typeof rawData === 'string') {
        parsed = JSON.parse(rawData)
      } else {
        parsed = rawData
      }

      if (!parsed || !parsed.event) {
        return
      }

      // Si és resposta de keep-alive
      if (parsed.event === 'pong') {
        return
      }

      this.dispatchInternal(parsed.event, parsed.data)
    } catch (err) {
      console.error('[MatchWSClient] Error parsejant missatge WebSocket:', err, rawData)
    }
  }

  private dispatchInternal(event: string, data: any): void {
    const handlers = this.listeners.get(event)
    if (handlers) {
      handlers.forEach((fn) => {
        try {
          fn(data)
        } catch (err) {
          console.error(`[MatchWSClient] Error en handler per a "${event}":`, err)
        }
      })
    }
  }

  private startHeartbeat(): void {
    this.stopHeartbeat()
    this.heartbeatInterval = setInterval(() => {
      if (this.isConnected) {
        this.send('ping', {})
      }
    }, this.heartbeatIntervalMs)
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval)
      this.heartbeatInterval = null
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn(`[MatchWSClient] S'ha assolit el límit d'intents de reconnexió (${this.maxReconnectAttempts}).`)
      this.dispatchInternal('reconnect_failed', { attempts: this.reconnectAttempts })
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 10000)

    this.reconnectTimeout = setTimeout(() => {
      if (!this.isExplicitlyClosed) {
        this.connect().catch((err) => {
          console.warn('[MatchWSClient] Reconnect fallit:', err)
        })
      }
    }, delay)
  }
}
