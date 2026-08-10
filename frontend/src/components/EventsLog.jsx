export default function EventsLog({ events = [] }) {
  return (
    <div className="panel events">
      <h3>Журнал событий</h3>
      <div className="events-list">
        {events.length === 0 && <span className="muted">Пока пусто</span>}
        {events.map(e => (
          <div key={e.id} className="event">
            <span className="day">[{e.day}]</span> {e.message}
          </div>
        ))}
      </div>
    </div>
  )
}
