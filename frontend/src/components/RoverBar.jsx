const STATUS = { idle: 'свободен', on_mission: 'в полёте', broken: 'сломан', charging: 'зарядка' }

export default function RoverBar({ rovers = [], selectedRoverId, onSelect }) {
  return (
    <div className="panel">
      <h3>Роверы</h3>
      <div className="list">
        {rovers.map(r => (
          <div
            key={r.id}
            className={`item ${r.id === selectedRoverId ? 'selected' : ''} ${r.status !== 'idle' ? 'busy' : ''}`}
            onClick={() => r.status === 'idle' && onSelect(r.id)}
          >
            <b>{r.name}</b>
            <span>🔋 {r.battery}% · ⚖️ {r.capacity}кг · {STATUS[r.status] || r.status}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
