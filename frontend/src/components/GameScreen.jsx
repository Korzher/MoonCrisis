import { useEffect, useState } from 'react'
import { api } from '../client'
import GameMap from './GameMap'
import RoverBar from './RoverBar'
import OrderList from './OrderList'

export default function GameScreen({ onMenu }) {
  const [state, setState] = useState(null)
  const [rovers, setRovers] = useState([])
  const [orders, setOrders] = useState([])
  const [selectedRoverId, setSelectedRoverId] = useState(null)
  const [selectedOrderId, setSelectedOrderId] = useState(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function load() {
    const [st, rv, ord] = await Promise.all([api.state(), api.rovers(), api.orders()])
    setState(st); setRovers(rv); setOrders(ord)
  }
  useEffect(() => { load() }, [])

  async function nextDay() {
    setError('')
    await api.nextDay()
    await load()
  }

  async function send() {
    setError('')
    if (!selectedRoverId || !selectedOrderId) { setError('Выберите ровер и заказ'); return }
    setBusy(true)
    try {
      await api.assign(selectedRoverId, selectedOrderId)
      await load()
    } catch (e) {
      setError(e.message)
    } finally { setBusy(false) }
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

      <div className="layout">
        <GameMap
          rovers={rovers} orders={orders}
          selectedOrderId={selectedOrderId}
          onSelectOrder={setSelectedOrderId}
        />
        <aside className="sidebar">
          <RoverBar rovers={rovers} selectedRoverId={selectedRoverId} onSelect={setSelectedRoverId} />
          <OrderList orders={orders} selectedOrderId={selectedOrderId} onSelect={setSelectedOrderId} />
          <button onClick={send} disabled={busy} className="send">🚀 Отправить</button>
          {error && <div className="error">{error}</div>}
        </aside>
      </div>

      {state.game_over && <div className="gameover">Игра окончена</div>}
    </div>
  )
}
