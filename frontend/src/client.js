const BASE = ''

async function req(path, options = {}) {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => null)
  if (!res.ok) {
    throw new Error(data?.error || `Ошибка ${res.status}`)
  }
  return data
}

export const api = {
  start: () => req('/api/game/start', { method: 'POST' }),
  state: () => req('/api/game/state'),
  nextDay: () => req('/api/game/next-day', { method: 'POST' }),
  rovers: () => req('/api/rovers'),
  orders: () => req('/api/orders'),
  events: () => req('/api/events'),
  assign: (roverId, orderId) =>
    req('/api/deliveries', {
      method: 'POST',
      body: JSON.stringify({ rover_id: roverId, order_id: orderId }),
    }),
  buyRover: () => req('/api/shop/buy', { method: 'POST' }),
  repair: (id) =>
    req('/api/shop/repair', { method: 'POST', body: JSON.stringify({ id }) }),
  charge: (id) =>
    req('/api/shop/charge', { method: 'POST', body: JSON.stringify({ id }) }),
}
