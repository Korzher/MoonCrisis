const BASE = { x: 4, y: 9 }

// Зоны повторяют логику backend (service/game.go : zoneAt)
function zoneClass(x, y) {
  if (y >= 7 && x > 0 && x < 9) return 'zone-plateau'
  if (y >= 4) return 'zone-mountains'
  return 'zone-craters'
}

export default function GameMap({ rovers = [], orders = [], selectedOrderId, onSelectOrder }) {
  const cells = []
  for (let y = 0; y < 10; y++) {
    for (let x = 0; x < 10; x++) {
      const isBase = x === BASE.x && y === BASE.y
      const rover = rovers.find(r => r.x === x && r.y === y)
      const order = orders.find(o => o.x === x && o.y === y &&
        (o.status === 'available' || o.status === 'active'))
      cells.push(
        <div
          key={`${x}-${y}`}
          className={`cell ${zoneClass(x, y)} ${isBase ? 'base' : ''}`}
          onClick={() => order && onSelectOrder && onSelectOrder(order.id)}
        >
          {isBase && <span className="base-icon">🏠</span>}
          {rover && <span className="rover-icon">🚙</span>}
          {order && !isBase && (
            <span className={`order-mark ${order.id === selectedOrderId ? 'selected' : ''}`}>
              {order.id}
            </span>
          )}
        </div>
      )
    }
  }
  return <div className="map">{cells}</div>
}
