// src/socket/SocketContext.js
import React, { createContext, useContext, useRef } from "react";
import { getSocket } from "./socket";

const SocketContext = createContext();

export function SocketProvider({ children }) {
  const socketRef = useRef(getSocket());
  return (
    <SocketContext.Provider value={socketRef.current}>
      {children}
    </SocketContext.Provider>
  );
}

export function useSocket() {
  return useContext(SocketContext);
}