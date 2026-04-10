export interface Point {
  x: number
  y: number
}

export type Direction = 'up' | 'down' | 'left' | 'right'
export type GameStatus = 'idle' | 'running' | 'paused' | 'over'

export interface GameConfig {
  gridSize: number
  initialSpeedMs: number
  minSpeedMs: number
  speedStepMs: number
  scoreStep: number
  leaderboardLimit: number
  defaultPlayer: string
}

export interface SnakeState {
  snake: Point[]
  direction: Direction
  queuedDirection: Direction
  food: Point
  status: GameStatus
  score: number
  applesEaten: number
  speedMs: number
  startedAtMs: number | null
  finishedAtMs: number | null
}

export interface ScoreEntry {
  playerName: string
  score: number
  durationSeconds: number
  applesEaten: number
  recordedAt: string
}

export interface SubmitScorePayload {
  playerName: string
  score: number
  durationSeconds: number
  applesEaten: number
}

export interface SubmitScoreResult {
  rank: number
  entry: ScoreEntry
  message?: string
}

export interface HealthStatus {
  status: string
  service: string
  version: string
  storageReady: boolean
}

export interface GameOverPayload {
  score: number
  applesEaten: number
  durationSeconds: number
}
