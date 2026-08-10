import { useState } from 'react'
import { api } from './api/client'
import MenuScreen from './components/MenuScreen'
import GameScreen from './components/GameScreen'
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
        <GameScreen onMenu={() => setScreen('menu')} />
      )}
    </div>
  )
}
