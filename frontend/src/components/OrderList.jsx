export default function OrderList({ orders = [], selectedOrderId, onSelect }) {
  const shown = orders.filter(o => o.status === 'available' || o.status === 'active')
  return (
    <div className="panel">
      <h3>Заказы</h3>
      <div className="list">
        {shown.map(o => (
          <div
            key={o.id}
            className={`item ${o.id === selectedOrderId ? 'selected' : ''}`}
            onClick={() => o.status === 'available' && onSelect(o.id)}
          >
            <b>#{o.id} {o.title}</b>
            <span>⚖️ {o.weight}кг · 💰 {o.reward}₽ · ⏳ до {o.deadline}д · ⚠️ риск {o.risk}</span>
            <span className="coords">({o.x},{o.y})</span>
          </div>
        ))}
      </div>
    </div>
  )
}
