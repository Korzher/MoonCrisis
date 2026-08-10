export default function MenuScreen({ onStart }) {
  return (
    <div className="menu">
      <h1>🌕 MoonCrisis</h1>
      <p>Доставка грузов по Луне</p>
      <button onClick={onStart}>Новая игра</button>
    </div>
  )
}
