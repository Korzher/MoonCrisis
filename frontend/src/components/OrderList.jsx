export default function OrderList({ orders = [], day = 1, selectedOrderId, onSelect }) {
  const shown = orders.filter(o => o.status === 'available' || o.status === 'active')
  return (
    <div className="panel">
      <h3>Заказы</h3>
      <div className="list">
        {shown.map(o => {
          const left = o.deadline - day
          return (
            <div
              key={o.id}
              className={`item ${o.id === selectedOrderId ? 'selected' : ''} ${o.status === 'active' ? 'busy' : ''}`}
              onClick={() => o.status === 'available' && onSelect(o.id)}>
              <b>#{o.id} {o.title}</b>
              <span>⚖️ {o.weight}кг · 💰 {o.reward}₽ · ⚠️ риск {o.risk}</span>
              <span className="coords">({o.x},{o.y}) · ⏳ {left >= 0 ? `осталось ${left} дн` : 'просрочен'}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
