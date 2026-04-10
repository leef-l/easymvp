import { startTransition, useEffect, useState } from 'react'

import './App.css'
import {
  fetchGameConfig,
  fetchHealth,
  fetchLeaderboard,
  submitScore,
} from './api/client'
import {
  DEFAULT_GAME_CONFIG,
  directionLabels,
} from './game/engine'
import { useSnakeGame } from './hooks/useSnakeGame'
import type {
  Direction,
  GameConfig,
  HealthStatus,
  ScoreEntry,
} from './types'

const PLAYER_NAME_KEY = 'snake-player-name'
const BEST_SCORE_KEY = 'snake-best-score'

function App() {
  const [config, setConfig] = useState<GameConfig>(DEFAULT_GAME_CONFIG)
  const [health, setHealth] = useState<HealthStatus | null>(null)
  const [leaderboard, setLeaderboard] = useState<ScoreEntry[]>([])
  const [backendState, setBackendState] = useState<'checking' | 'online' | 'offline'>('checking')
  const [playerName, setPlayerName] = useState(
    () => window.localStorage.getItem(PLAYER_NAME_KEY) || DEFAULT_GAME_CONFIG.defaultPlayer,
  )
  const [bestScore, setBestScore] = useState(() =>
    Number(window.localStorage.getItem(BEST_SCORE_KEY) || 0),
  )
  const [banner, setBanner] = useState('后端配置加载中...')

  const refreshLeaderboard = async () => {
    const data = await fetchLeaderboard()
    startTransition(() => {
      setLeaderboard(data.items)
    })
  }

  useEffect(() => {
    let cancelled = false

    const load = async () => {
      try {
        const [healthStatus, configResponse, scoreResponse] = await Promise.all([
          fetchHealth(),
          fetchGameConfig(),
          fetchLeaderboard(),
        ])
        if (cancelled) {
          return
        }
        startTransition(() => {
          setHealth(healthStatus)
          setConfig(configResponse.config)
          setLeaderboard(scoreResponse.items)
          setBackendState('online')
          setBanner('后端已连接，排行榜与速度参数来自 GoFrame 服务。')
        })
      } catch {
        if (cancelled) {
          return
        }
        startTransition(() => {
          setBackendState('offline')
          setBanner('后端暂不可用，当前使用前端默认配置运行游戏。')
        })
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    window.localStorage.setItem(PLAYER_NAME_KEY, playerName)
  }, [playerName])

  const game = useSnakeGame(config, async ({ score, applesEaten, durationSeconds }) => {
    if (score > bestScore) {
      setBestScore(score)
      window.localStorage.setItem(BEST_SCORE_KEY, String(score))
    }
    if (score <= 0 || backendState !== 'online') {
      setBanner(
        score > 0
          ? '本局已结束，但后端离线，排行榜没有提交。'
          : '本局得分为 0，排行榜未提交。',
      )
      return
    }

    try {
      const payload = await submitScore({
        playerName: playerName.trim() || config.defaultPlayer,
        score,
        durationSeconds,
        applesEaten,
      })
      setBanner(`已提交本局成绩，当前排行第 ${payload.rank}。`)
      await refreshLeaderboard()
    } catch (error) {
      setBanner(error instanceof Error ? error.message : '成绩提交失败')
    }
  })

  const snakeLookup = new Map(
    game.state.snake.map((point, index) => [`${point.x},${point.y}`, index]),
  )

  const controls: Direction[] = ['up', 'left', 'down', 'right']

  return (
    <main className="page-shell">
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Workflow V2 Snake Sample</p>
          <h1>Neon Snake Control</h1>
          <p className="hero-copy">
            React 前端负责网格渲染与输入，GoFrame v2 后端负责配置与排行榜。
            这是一个为 Workflow V2 真实验证准备的全栈贪吃蛇样例。
          </p>
        </div>
        <div className={`backend-pill is-${backendState}`}>
          <span className="backend-dot" />
          {backendState === 'online' ? 'Backend Online' : backendState === 'offline' ? 'Backend Offline' : 'Checking'}
        </div>
      </section>

      <section className="layout-grid">
        <article className="play-panel">
          <div className="board-toolbar">
            <div>
              <p className="panel-kicker">实时棋盘</p>
              <h2>20x20 光栅竞技场</h2>
            </div>
            <div className={`status-badge is-${game.state.status}`}>{game.state.status}</div>
          </div>

          <div
            className="snake-board"
            style={{ gridTemplateColumns: `repeat(${config.gridSize}, minmax(0, 1fr))` }}
          >
            {Array.from({ length: config.gridSize * config.gridSize }, (_, index) => {
              const x = index % config.gridSize
              const y = Math.floor(index / config.gridSize)
              const key = `${x},${y}`
              const snakePart = snakeLookup.get(key)
              const isFood = game.state.food.x === x && game.state.food.y === y

              let cellClass = 'cell'
              if (snakePart === 0) {
                cellClass += ' is-head'
              } else if (typeof snakePart === 'number') {
                cellClass += ' is-body'
              } else if (isFood) {
                cellClass += ' is-food'
              }

              return <div key={key} className={cellClass} />
            })}
          </div>

          <div className="controls-row">
            <button type="button" className="action-button" onClick={game.start}>
              {game.state.status === 'idle' ? '开始游戏' : '继续前进'}
            </button>
            <button type="button" className="action-button ghost" onClick={game.togglePause}>
              {game.state.status === 'running' ? '暂停' : '恢复'}
            </button>
            <button type="button" className="action-button ghost" onClick={game.restart}>
              重新开始
            </button>
          </div>

          <div className="direction-pad">
            {controls.map((direction) => (
              <button
                key={direction}
                type="button"
                className={`pad-button dir-${direction}`}
                onClick={() => game.turn(direction)}
              >
                {directionLabels[direction]}
              </button>
            ))}
          </div>
        </article>

        <aside className="side-panel">
          <section className="card">
            <p className="panel-kicker">控制台</p>
            <div className="stats-grid">
              <div>
                <span>当前分数</span>
                <strong>{game.state.score}</strong>
              </div>
              <div>
                <span>吞食苹果</span>
                <strong>{game.state.applesEaten}</strong>
              </div>
              <div>
                <span>当前方向</span>
                <strong>{directionLabels[game.state.direction]}</strong>
              </div>
              <div>
                <span>速度</span>
                <strong>{game.state.speedMs} ms</strong>
              </div>
              <div>
                <span>个人最佳</span>
                <strong>{bestScore}</strong>
              </div>
              <div>
                <span>后端存储</span>
                <strong>{health?.storageReady ? 'ready' : 'pending'}</strong>
              </div>
            </div>
          </section>

          <section className="card">
            <p className="panel-kicker">玩家配置</p>
            <label className="input-label">
              玩家名
              <input
                value={playerName}
                maxLength={24}
                onChange={(event) => setPlayerName(event.target.value)}
                placeholder={config.defaultPlayer}
              />
            </label>
            <p className="banner-text">{banner}</p>
          </section>

          <section className="card">
            <div className="leaderboard-head">
              <div>
                <p className="panel-kicker">GoFrame 排行榜</p>
                <h3>Top {config.leaderboardLimit}</h3>
              </div>
              <button type="button" className="mini-button" onClick={() => void refreshLeaderboard()}>
                刷新
              </button>
            </div>

            {leaderboard.length === 0 ? (
              <p className="empty-state">还没有成绩，先跑一局再把分数打上榜。</p>
            ) : (
              <ol className="leaderboard-list">
                {leaderboard.map((entry) => (
                  <li key={`${entry.playerName}-${entry.recordedAt}`} className="leaderboard-item">
                    <div>
                      <strong>{entry.playerName}</strong>
                      <span>{entry.applesEaten} 苹果</span>
                    </div>
                    <div>
                      <strong>{entry.score}</strong>
                      <span>{entry.durationSeconds}s</span>
                    </div>
                  </li>
                ))}
              </ol>
            )}
          </section>

          <section className="card tips-card">
            <p className="panel-kicker">玩法说明</p>
            <ul>
              <li>键盘方向键或 WASD 控制方向。</li>
              <li>空格键可以暂停或恢复。</li>
              <li>撞墙或撞到自己会立即结束。</li>
              <li>速度会随着分数逐渐提升。</li>
            </ul>
          </section>
        </aside>
      </section>
    </main>
  )
}

export default App
