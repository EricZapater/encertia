import apiClient from '@/api/client'
import type {
  CreateMatchRequest,
  MatchCreatedResponse,
  MatchPublicInfo,
  JoinMatchRequest,
  JoinMatchResponse,
  MatchSummaryResponse
} from './types'

/**
 * Crea una nova partida en directe per a un qüestionari publicat.
 */
export async function createMatch(data: CreateMatchRequest): Promise<MatchCreatedResponse> {
  const response = await apiClient.post<MatchCreatedResponse>('/matches', data)
  return response.data
}

/**
 * Obté informació pública bàsica d'una partida pel seu PIN de 6 dígits.
 */
export async function getMatchByPin(pin: string): Promise<MatchPublicInfo> {
  const response = await apiClient.get<MatchPublicInfo>(`/matches/${pin}`)
  return response.data
}

/**
 * Registra l'usuari autenticat com a jugador a una partida en estat lobby.
 */
export async function joinMatch(
  pin: string,
  data: JoinMatchRequest
): Promise<JoinMatchResponse> {
  const response = await apiClient.post<JoinMatchResponse>(`/matches/${pin}/join`, data)
  return response.data
}

/**
 * Obté el resum complet i podi d'una partida finalitzada.
 */
export async function getMatchSummary(matchId: string): Promise<MatchSummaryResponse> {
  const response = await apiClient.get<MatchSummaryResponse>(`/matches/${matchId}/summary`)
  return response.data
}
