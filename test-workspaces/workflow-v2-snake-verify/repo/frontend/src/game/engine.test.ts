import { describe, expect, it } from 'vitest'

import { DEFAULT_GAME_CONFIG, createInitialState, queueDirection, stepGame } from './engine'

describe('snake engine', () => {
  it('grows the snake and increases the score when food is eaten', () => {
    const initial = createInitialState(DEFAULT_GAME_CONFIG, () => 0)
    const running = {
      ...initial,
      status: 'running' as const,
      food: { x: initial.snake[0].x + 1, y: initial.snake[0].y },
    }

    const next = stepGame(running, DEFAULT_GAME_CONFIG, () => 0)

    expect(next.score).toBe(DEFAULT_GAME_CONFIG.scoreStep)
    expect(next.applesEaten).toBe(1)
    expect(next.snake).toHaveLength(running.snake.length + 1)
  })

  it('marks the game as over after hitting the wall', () => {
    const initial = createInitialState(DEFAULT_GAME_CONFIG, () => 0)
    const running = {
      ...initial,
      status: 'running' as const,
      snake: [
        { x: DEFAULT_GAME_CONFIG.gridSize - 1, y: 0 },
        { x: DEFAULT_GAME_CONFIG.gridSize - 2, y: 0 },
        { x: DEFAULT_GAME_CONFIG.gridSize - 3, y: 0 },
      ],
    }

    const next = stepGame(running, DEFAULT_GAME_CONFIG, () => 0)

    expect(next.status).toBe('over')
  })

  it('prevents reversing into the opposite direction', () => {
    const initial = createInitialState(DEFAULT_GAME_CONFIG, () => 0)
    const next = queueDirection(initial, 'left')

    expect(next.queuedDirection).toBe('right')
  })
})
