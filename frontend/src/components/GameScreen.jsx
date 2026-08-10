import { useEffect, useState } from 'react'
import { api } from '../api/client'
import GameMap from './GameMap'

export default function GameScreen({ onMenu }) {
  const [state, setState] = useState(null)
  const [rovers, setRovers] = useState([])
  const [orders, setOrders] = useState([])
  const [selectedOrderId, setSelectedOrderId] = useState(null)

  async function load() {
    const [st, rv, ord] = await Promise.all([api.state(), api.rovers(), api.orders()])
    setState(st)
    setRovers(rv)
    setOrders(ord)
  }

  useEffect(() => { load() }, [])

  async function nextDay() {
    await api.nextDay()
    await load()
  }

  if (!state) return <div className="placeholder">Загрузка…</div>

  return (
    <div className="game">
      <header className="hud">
        <span>День {state.day}</span>
        <span>💰 {state.money}</span>
        <span>⭐ Рейтинг {state.rating}</span>
        <button onClick={nextDay}>Следующий день ➡</button>
        <button onClick={onMenu} className="ghost">В меню</button>
      </header>
      <GameMap
        rovers={rovers}
        orders={orders}
        selectedOrderId={selectedOrderId}
        onSelectOrder={setSelectedOrderId}
      />
      {state.game_over && <div className="gameover">Игра окончена</div>}
    </div>
  )
}
