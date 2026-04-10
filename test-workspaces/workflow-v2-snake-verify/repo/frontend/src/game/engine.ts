import type { Direction, GameConfig, Point, SnakeState } from '../types'

export const DEFAULT_GAME_CONFIG: GameConfig = {
  gridSize: 20,
  initialSpeedMs: 160,
  minSpeedMs: 80,
  speedStepMs: 6,
  scoreStep: 10,
  leaderboardLimit: 8,
  defaultPlayer: 'Player 1',
}

export const directionLabels: Record<Direction, string> = {
  up: '↑',
  down: '↓',
  left: '←',
  right: '→',
}

const vectors: Record<Direction, Point> = {
  up: { x: 0, y: -1 },
  down: { x: 0, y: 1 },
  left: { x: -1, y: 0 },
  right: { x: 1, y: 0 },
}

const oppositeDirections: Record<Direction, Direction> = {
  up: 'down',
  down: 'up',
  left: 'right',
  right: 'left',
}

export function createInitialState(config: GameConfig, random: () => number = Math.random): SnakeState {
  const center = Math.floor(config.gridSize / 2)
  const snake = [
    { x: center, y: center },
    { x: center - 1, y: center },
    { x: center - 2, y: center },
  ]

  return {
    snake,
    direction: 'right',
    queuedDirection: 'right',
    food: spawnFood(config.gridSize, snake, random),
    status: 'idle',
    score: 0,
    applesEaten: 0,
    speedMs: config.initialSpeedMs,
    startedAtMs: null,
    finishedAtMs: null,
  }
}

export function beginGame(config: GameConfig, random: () => number = Math.random): SnakeState {
  return {
    ...createInitialState(config, random),
    status: 'running',
    startedAtMs: Date.now(),
  }
}

export function queueDirection(state: SnakeState, direction: Direction): SnakeState {
  if (
    direction === state.queuedDirection ||
    oppositeDirections[state.queuedDirection] === direction
  ) {
    return state
  }
  return {
    ...state,
    queuedDirection: direction,
  }
}

export function stepGame(
  state: SnakeState,
  config: GameConfig,
  random: () => number = Math.random,
): SnakeState {
  if (state.status !== 'running') {
    return state
  }

  const direction = state.queuedDirection
  const head = state.snake[0]
  const vector = vectors[direction]
  const nextHead = {
    x: head.x + vector.x,
    y: head.y + vector.y,
  }

  if (
    nextHead.x < 0 ||
    nextHead.y < 0 ||
    nextHead.x >= config.gridSize ||
    nextHead.y >= config.gridSize
  ) {
    return finishGame(state, direction)
  }

  const ateFood = nextHead.x === state.food.x && nextHead.y === state.food.y
  const nextBody = ateFood ? state.snake : state.snake.slice(0, -1)

  if (nextBody.some((segment) => segment.x === nextHead.x && segment.y === nextHead.y)) {
    return finishGame(state, direction)
  }

  const nextSnake = [nextHead, ...nextBody]
  const nextFood = ateFood
    ? spawnFood(config.gridSize, nextSnake, random)
    : state.food

  return {
    ...state,
    snake: nextSnake,
    direction,
    queuedDirection: direction,
    food: nextFood,
    score: state.score + (ateFood ? config.scoreStep : 0),
    applesEaten: state.applesEaten + (ateFood ? 1 : 0),
    speedMs: ateFood
      ? Math.max(config.minSpeedMs, state.speedMs - config.speedStepMs)
      : state.speedMs,
  }
}

function finishGame(state: SnakeState, direction: Direction): SnakeState {
  return {
    ...state,
    direction,
    queuedDirection: direction,
    status: 'over',
    finishedAtMs: state.finishedAtMs ?? Date.now(),
  }
}

export function spawnFood(gridSize: number, snake: Point[], random: () => number): Point {
  const occupied = new Set(snake.map((point) => `${point.x},${point.y}`))
  const available: Point[] = []

  for (let y = 0; y < gridSize; y += 1) {
    for (let x = 0; x < gridSize; x += 1) {
      const key = `${x},${y}`
      if (!occupied.has(key)) {
        available.push({ x, y })
      }
    }
  }

  if (available.length === 0) {
    return snake[snake.length - 1]
  }

  const index = Math.min(available.length - 1, Math.floor(random() * available.length))
  return available[index]
}
