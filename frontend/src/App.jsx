import { useState } from 'react'
import { api } from './client'
import MenuScreen from './components/MenuScreen'
import './index.css'

export default function App() {
  const [screen, setScreen] = useState('menu')

  async function handleStart() {
    await api.start()
    setScreen('game')
  }

  return (
    <div className="app">
      {screen === 'menu' ? (
        <MenuScreen onStart={handleStart} />
      ) : (
        <div className="placeholder">Здесь будет карта Луны 🌑</div>
      )}
    </div>
  )
}
