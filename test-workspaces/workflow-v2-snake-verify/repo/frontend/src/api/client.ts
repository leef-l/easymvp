import type {
  GameConfig,
  HealthStatus,
  ScoreEntry,
  SubmitScorePayload,
  SubmitScoreResult,
} from '../types'

interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

const API_BASE = import.meta.env.VITE_API_BASE ?? ''

async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
    ...init,
  })

  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`)
  }

  const payload = (await response.json()) as ApiEnvelope<T>
  if (payload.code !== 0) {
    throw new Error(payload.message || 'unexpected api error')
  }
  return payload.data
}

export const fetchHealth = () =>
  apiRequest<{ status: HealthStatus }>('/api/health').then((payload) => payload.status)

export const fetchGameConfig = () =>
  apiRequest<{ config: GameConfig }>('/api/game/config')

export const fetchLeaderboard = () =>
  apiRequest<{ items: ScoreEntry[]; total: number }>('/api/game/scores')

export const submitScore = (payload: SubmitScorePayload) =>
  apiRequest<SubmitScoreResult>('/api/game/scores', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
