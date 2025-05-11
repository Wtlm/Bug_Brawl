import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import { AnimatePresence } from "framer-motion";
import { Routes, Route, useLocation } from 'react-router-dom';

import './App.css'

import Home from './frontend/home.jsx'
import Lobby from './frontend/lobby.jsx';


function App() {
  const location = useLocation(); 
  return (
    <AnimatePresence mode="wait">
      <Routes location={location} key={location.pathname}>
        <Route path="/" element={<Home />} />
        <Route path="/lobby" element={<Lobby />} />
      </Routes>
    </AnimatePresence>
  )
}

export default App
