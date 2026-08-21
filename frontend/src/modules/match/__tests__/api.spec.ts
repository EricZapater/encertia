import { describe, it, expect, vi, beforeEach } from 'vitest'
import apiClient from '@/api/client'
import {
  createMatch,
  getMatchByPin,
  joinMatch,
  getMatchSummary
} from '../api'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn()
  }
}))

describe('Match API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('createMatch crida correctament a POST /matches amb el quizId', async () => {
    const mockResponse = {
      data: {
        id: 'm123',
        quizId: 'q456',
        hostId: 'u789',
        pin: '123456',
        status: 'lobby',
        playUrl: 'https://encertia.ericzapater.cat/play?pin=123456'
      }
    }
    vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

    const result = await createMatch({ quizId: 'q456' })

    expect(apiClient.post).toHaveBeenCalledWith('/matches', { quizId: 'q456' })
    expect(result.pin).toBe('123456')
    expect(result.status).toBe('lobby')
  })

  it('getMatchByPin crida a GET /matches/:pin', async () => {
    const mockResponse = {
      data: {
        id: 'm123',
        pin: '654321',
        quizTitle: 'Matemàtiques Divertides',
        hostName: 'Professora Maria',
        status: 'lobby',
        playerCount: 5
      }
    }
    vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse)

    const result = await getMatchByPin('654321')

    expect(apiClient.get).toHaveBeenCalledWith('/matches/654321')
    expect(result.quizTitle).toBe('Matemàtiques Divertides')
    expect(result.playerCount).toBe(5)
  })

  it('joinMatch crida a POST /matches/:pin/join amb el Nickname', async () => {
    const mockResponse = {
      data: {
        matchId: 'm123',
        playerId: 'p999',
        nickname: 'SuperPol',
        pin: '123456',
        status: 'lobby'
      }
    }
    vi.mocked(apiClient.post).mockResolvedValueOnce(mockResponse)

    const result = await joinMatch('123456', { nickname: 'SuperPol' })

    expect(apiClient.post).toHaveBeenCalledWith('/matches/123456/join', { nickname: 'SuperPol' })
    expect(result.playerId).toBe('p999')
    expect(result.nickname).toBe('SuperPol')
  })

  it('getMatchSummary crida a GET /matches/:id/summary', async () => {
    const mockResponse = {
      data: {
        matchId: 'm123',
        quizTitle: 'Ciències Naturals',
        totalQuestions: 10,
        totalPlayers: 15,
        podium: [
          { playerId: 'p1', nickname: 'Anna', score: 10, rank: 1 }
        ],
        leaderboard: [
          { playerId: 'p1', nickname: 'Anna', score: 10, rank: 1 }
        ]
      }
    }
    vi.mocked(apiClient.get).mockResolvedValueOnce(mockResponse)

    const result = await getMatchSummary('m123')

    expect(apiClient.get).toHaveBeenCalledWith('/matches/m123/summary')
    expect(result.podium).toHaveLength(1)
    expect(result.totalPlayers).toBe(15)
  })
})
