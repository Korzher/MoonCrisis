const STATUS = { idle: 'свободен', on_mission: 'в пути', broken: 'сломан', charging: 'зарядка' }

export default function RoverBar({ rovers = [], deliveries = [], orders = [], selectedRoverId, onSelect }) {
  return (
    <div className="panel">
      <h3>Роверы</h3>
      <div className="list">
        {rovers.map(r => {
          const d = (deliveries || []).find(x => x.rover_id === r.id)
          const target = d && (orders || []).find(o => o.id === d.order_id)
          return (
            <div
              key={r.id}
              className={`item ${r.id === selectedRoverId ? 'selected' : ''} ${r.status !== 'idle' ? 'busy' : ''}`}
              onClick={() => r.status === 'idle' && onSelect(r.id)}
            >
              <b>{r.name}</b>
              <span>🔋 {r.battery}% · ⚖️ {r.capacity}кг · {STATUS[r.status] || r.status}</span>
              {r.status === 'on_mission' && target && (
                <span className="coords">→ заказ #{target.id} {target.title}</span>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
