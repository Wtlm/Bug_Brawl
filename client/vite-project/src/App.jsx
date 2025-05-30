import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import { AnimatePresence } from "framer-motion";
import { Routes, Route, useLocation } from 'react-router-dom';

import './App.css'

import Home from './frontend/home.jsx'
import Lobby from './frontend/lobby.jsx';
import Game from './frontend/game.jsx';
import GameOver from './frontend/gameover.jsx';
import { SocketProvider } from './socket/socketContext.jsx';



function App() {
  const location = useLocation();
  return (
    <SocketProvider>
      <AnimatePresence mode="wait">
        <Routes location={location} key={location.pathname}>
          <Route path="/" element={<Home />} />
          <Route path="/lobby" element={<Lobby />} />
          <Route path="/game" element={<Game />} />
          <Route path="/gameover" element={<GameOver />} />
        </Routes>
      </AnimatePresence>
    </SocketProvider>
  )
}

export default App
