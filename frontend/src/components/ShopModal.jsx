export default function ShopModal({ rovers = [], money, onBuy, onRepair, onCharge, onClose }) {
  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2>🛒 Магазин</h2>

        <section>
          <h3>Новый ровер — 500₽</h3>
          <button onClick={onBuy} disabled={money < 500}>Купить ровер (г/п 200кг)</button>
          {money < 500 && <div className="muted">Недостаточно средств</div>}
        </section>

        <section>
          <h3>Обслуживание</h3>
          <div className="list">
            {rovers.map(r => (
              <div key={r.id} className="item">
                <b>{r.name}</b>
                <span>🔋 {r.battery}% · {r.status}</span>
                <span className="row">
                  <button onClick={() => onCharge(r.id)}
                    disabled={r.battery >= 100 || r.status !== 'idle'}>Зарядить 50₽</button>
                  <button onClick={() => onRepair(r.id)}
                    disabled={r.status !== 'broken'}>Ремонт 100₽</button>
                </span>
              </div>
            ))}
          </div>
        </section>

        <button className="ghost" onClick={onClose}>Закрыть</button>
      </div>
    </div>
  )
}
