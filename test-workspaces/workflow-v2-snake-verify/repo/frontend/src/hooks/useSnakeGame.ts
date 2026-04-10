import { useEffect, useEffectEvent, useRef, useState } from 'react'

import { beginGame, createInitialState, queueDirection, stepGame } from '../game/engine'
import type { Direction, GameConfig, GameOverPayload, SnakeState } from '../types'

type GameOverHandler = (payload: GameOverPayload) => void | Promise<void>

export function useSnakeGame(config: GameConfig, onGameOver?: GameOverHandler) {
  const [state, setState] = useState<SnakeState>(() => createInitialState(config))
  const previousStatus = useRef(state.status)

  useEffect(() => {
    setState(createInitialState(config))
  }, [config])

  const emitGameOver = useEffectEvent((nextState: SnakeState) => {
    if (!onGameOver || nextState.status !== 'over' || !nextState.startedAtMs) {
      return
    }

    const endTime = nextState.finishedAtMs ?? Date.now()
    void onGameOver({
      score: nextState.score,
      applesEaten: nextState.applesEaten,
      durationSeconds: Math.max(1, Math.round((endTime - nextState.startedAtMs) / 1000)),
    })
  })

  useEffect(() => {
    if (state.status === 'over' && previousStatus.current !== 'over') {
      emitGameOver(state)
    }
    previousStatus.current = state.status
  }, [state])

  const tick = useEffectEvent(() => {
    setState((current) => stepGame(current, config))
  })

  useEffect(() => {
    if (state.status !== 'running') {
      return undefined
    }

    const timer = window.setInterval(() => {
      tick()
    }, state.speedMs)

    return () => window.clearInterval(timer)
  }, [state.speedMs, state.status])

  const handleKeyDown = useEffectEvent((event: KeyboardEvent) => {
    switch (event.key) {
      case 'ArrowUp':
      case 'w':
      case 'W':
        setState((current) => queueDirection(current, 'up'))
        break
      case 'ArrowDown':
      case 's':
      case 'S':
        setState((current) => queueDirection(current, 'down'))
        break
      case 'ArrowLeft':
      case 'a':
      case 'A':
        setState((current) => queueDirection(current, 'left'))
        break
      case 'ArrowRight':
      case 'd':
      case 'D':
        setState((current) => queueDirection(current, 'right'))
        break
      case ' ':
        event.preventDefault()
        setState((current) => {
          if (current.status === 'running') {
            return { ...current, status: 'paused' }
          }
          if (current.status === 'paused') {
            return { ...current, status: 'running' }
          }
          return current
        })
        break
      default:
        break
    }
  })

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [])

  return {
    state,
    start: () =>
      setState((current) => {
        if (current.status === 'paused') {
          return { ...current, status: 'running' }
        }
        return beginGame(config)
      }),
    restart: () => setState(beginGame(config)),
    togglePause: () =>
      setState((current) => {
        if (current.status === 'running') {
          return { ...current, status: 'paused' }
        }
        if (current.status === 'paused') {
          return { ...current, status: 'running' }
        }
        return current
      }),
    turn: (direction: Direction) =>
      setState((current) => queueDirection(current, direction)),
  }
}
