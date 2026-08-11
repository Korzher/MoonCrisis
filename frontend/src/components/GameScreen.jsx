import { useEffect, useState } from 'react'
import { api } from '../client'
import GameMap from './GameMap'
import RoverBar from './RoverBar'
import OrderList from './OrderList'
import EventsLog from './EventsLog'
import ShopModal from './ShopModal'

export default function GameScreen({ onMenu }) {
  const [state, setState] = useState(null)
  const [rovers, setRovers] = useState([])
  const [orders, setOrders] = useState([])
  const [events, setEvents] = useState([])
  const [selectedRoverId, setSelectedRoverId] = useState(null)
  const [selectedOrderId, setSelectedOrderId] = useState(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const [shopOpen, setShopOpen] = useState(false)

  async function load() {
    const [st, rv, ord, ev] = await Promise.all([
      api.state(), api.rovers(), api.orders(), api.events(),
    ])
    setState(st); setRovers(rv); setOrders(ord); setEvents(ev)
  }
  useEffect(() => { load() }, [])

  async function nextDay() {
    setError('')
    try { await api.nextDay(); await load() }
    catch (e) { setError(e.message) }
  }

  async function send() {
    setError('')
    if (!selectedRoverId || !selectedOrderId) { setError('Выберите ровер и заказ'); return }
    setBusy(true)
    try { await api.assign(selectedRoverId, selectedOrderId); await load() }
    catch (e) { setError(e.message) }
    finally { setBusy(false) }
  }

  async function shopAction(fn) {
    setError('')
    try { await fn(); await load() }
    catch (e) { setError(e.message) }
  }

  if (!state) return <div className="placeholder">Загрузка…</div>

  return (
    <div className="game">
      <header className="hud">
        <span>День {state.day}</span>
        <span>💰 {state.money}</span>
        <span>⭐ Рейтинг {state.rating}</span>
        <button onClick={nextDay}>Следующий день ➡</button>
        <button onClick={() => setShopOpen(true)} className="shop-btn">🛒 Магазин</button>
        <button onClick={onMenu} className="ghost">В меню</button>
      </header>

      <div className="layout">
        <div className="left">
          <GameMap rovers={rovers} orders={orders}
            selectedOrderId={selectedOrderId} onSelectOrder={setSelectedOrderId} />
          <EventsLog events={events} />
        </div>
        <aside className="sidebar">
          <RoverBar rovers={rovers} selectedRoverId={selectedRoverId} onSelect={setSelectedRoverId} />
          <OrderList orders={orders} day={state.day} selectedOrderId={selectedOrderId} onSelect={setSelectedOrderId} />
          <button onClick={send} disabled={busy} className="send">🚀 Отправить</button>
          {error && <div className="error">{error}</div>}
        </aside>
      </div>

      {state.game_over && <div className="gameover">Игра окончена</div>}

      {shopOpen && (
        <ShopModal
          rovers={rovers} money={state.money}
          onBuy={() => shopAction(api.buyRover)}
          onRepair={id => shopAction(() => api.repair(id))}
          onCharge={id => shopAction(() => api.charge(id))}
          onClose={() => setShopOpen(false)}
        />
      )}
    </div>
  )
}
