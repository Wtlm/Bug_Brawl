import React, { useRef, useState, useEffect, memo } from 'react';
import { useNavigate } from "react-router-dom";
import { delay, motion } from "framer-motion";
import { getSocket } from "../socket/socket.js";
import Popup from "../widget/popup";
import { LobbyHandlers } from '../socket/lobbyHandlers.js';
import { useSocket } from '../socket/socketContext.jsx';

function Lobby() {
      console.log("Lobby component mounted");

  const buttonLabels = ["Create Match", "Join Match", "Find Match"];
  const [selectedLabel, setSelectedLabel] = useState('');
  const [showPopup, setShowPopup] = useState(false);
  const [playerName, setPlayerName] = useState("");
  const [roomCode, setRoomCode] = useState("");
  const [joinRoomCode, setJoinRoomCode] = useState("");
  const [showJoinCodeInput, setShowJoinCodeInput] = useState(false);
  const [showWaitingForHost, setShowWaitingForHost] = useState(false);
  const [playerCount, setPlayerCount] = useState(1);
  const [isHost, setIsHost] = useState(false);
  const [isNewHost, setIsNewHost] = useState(false);
  const [status, setStatus] = useState("");
  const [isFindingMatch, setIsFindingMatch] = useState(false);
  const [countdown, setCountdown] = useState(null);
  const socketRef = useSocket();
  const navigate = useNavigate();
  const handlers = new LobbyHandlers(
    // States
    {
      playerName,
      roomCode,
      joinRoomCode,
      showJoinCodeInput,
      showWaitingForHost,
      playerCount,
      isNewHost,
      isHost,
      status,
      isFindingMatch,
      countdown
    },
    // Setters
    {
      setSelectedLabel,
      setShowPopup,
      setPlayerName,
      setRoomCode,
      setJoinRoomCode,
      setShowJoinCodeInput,
      setShowWaitingForHost,
      setPlayerCount,
      setIsNewHost,
      setIsHost,
      setStatus,
      setIsFindingMatch,
      setCountdown
    },
    socketRef,
    navigate
  );

  useEffect(() => {
    // Create WebSocket once on component mount
    if (!socketRef.current) {
      handlers.getOrCreateSocket();
    }
    // Clean up WebSocket connection on component unmount
    return () => {
    };
  }, []);

  useEffect(() => {
    if (countdown === 0 && isHost) {
      navigate("/game");
    }
  }, [countdown, isHost]);

  return (
    <div className="h-full w-full flex flex-col gap-3 justify-center items-center">
      <motion.h1
        layoutId="name"
        className="lg:text-9xl text-8xl"
        initial={{ scale: 1 }}
        animate={{ scale: 0.8 }}
        transition={{ type: "spring", duration: 1 }}
      >
        BUG BRAWL
      </motion.h1>

      <motion.input
        type="text"
        value={playerName}
        onChange={(e) => setPlayerName(e.target.value)}
        placeholder="Enter your name"
        className=" mb-6 px-4 py-2 rounded-full bg-[#9f9f9f] decoration-[#3a3a3a] font-bold text-2xl text-center"
        initial={{ opacity: 0, scale: 0 }}
        animate={{ opacity: 1, scale: 1, }}
        transition={{
          type: "spring",
          stiffness: 200,
          damping: 20,
        }}
      />

      <div className="flex gap-6 justify-evenly">
        {buttonLabels.map((label) => (
          <motion.button
            key={label}
            className="w-max lg:text-lg text-sm"
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{
              type: "spring",
              stiffness: 200,
              damping: 20,
            }}
            whileHover={{ scale: 0.9 }}
            onClick={() => handlers.handleButtonClick(label)}
          >
            {label}
          </motion.button>
        ))}
      </div>

      <Popup show={showPopup} onClose={handlers.handleClosePopup} className="lg:w-1/4 lg:h-2/5">
        {selectedLabel === "Create Match" && !showWaitingForHost && (
          <div className="flex flex-col justify-between text-center items-center h-full">
            <div>
              <p className="text-lg mb-2">Room Code: <strong>{roomCode}</strong></p>
              <p className="text-lg">Players: {playerCount}/4</p>
            </div>
            <div>
              {isHost && playerCount >= 2 ? (
                <motion.button
                  className="!text-black w-full !px-4 !py-2 !text-base bg-green-500 rounded"
                  whileHover={{ scale: 0.9 }}
                  onClick={handlers.handleStartGame}
                >
                  Start Game
                </motion.button>
              ) : (
                <p>Waiting for more players...</p>
              )}
            </div>
          </div>
        )}

        {isNewHost && (
          <div className="flex flex-col justify-center text-center items-center h-full">
            <p className="text-lg mb-2">Room Code: <strong>{roomCode}</strong></p>
            <p className="text-lg">Players: {playerCount}/4</p>
            <div>
              {playerCount >= 2 ? (
                <motion.button
                  className="!text-black w-full !px-4 !py-2 !text-base bg-green-500 rounded"
                  whileHover={{ scale: 0.9 }}
                  onClick={handlers.handleStartGame}
                >
                  Start Game
                </motion.button>
              ) : (
                <p>Waiting for more players...</p>
              )}
            </div>
          </div>
        )}

        {selectedLabel === "Join Match" && showJoinCodeInput && (
          <div className="flex flex-col justify-between text-center items-center h-full">
            <div>
              <p className="text-lg mb-4">Enter Room Code</p>
              <motion.input
                type="text"
                value={joinRoomCode}
                onChange={(e) => setJoinRoomCode(e.target.value.toUpperCase())}
                className="bg-[#9f9f9f] rounded-lg m-3 px-4 py-1 w-2/3 text-black text-center !text-base" placeholder="Room Code"
                maxLength={4}
              />
            </div>
            <motion.button
              className="!text-black w-1/3 !px-4 !py-2 !text-base !bottom-1"
              whileHover={{ scale: 0.9 }}
              onClick={handlers.handleJoinRoom}
            >
              Join Room
            </motion.button>
          </div>
        )}

        {showWaitingForHost && (
          <div className="flex flex-col justify-center text-center items-center h-full">
            <p className="text-lg">Waiting for host to start the game...</p>
            <p className="text-lg">Players: {playerCount}/4</p>
            <motion.button
              className="!text-black w-1/3 !px-4 !py-2 !text-base !bottom-1"
              whileHover={{ scale: 0.9 }}
              onClick={handlers.handleClosePopup}
            >
              Leave Room
            </motion.button>
          </div>
        )}

        {selectedLabel === "Find Match" && (
          <div className="flex flex-col justify-center text-center items-center h-full">
            <p className="text-lg mb-4">
              {countdown !== null ? `Starting in ${countdown}...` : "Finding Match"}
            </p>

            {!isFindingMatch && countdown !== null ? (
              // Show match found state
              <div className="mb-4">
                <p className="text-lg">Players: {playerCount}/4</p>
              </div>
            ) : isFindingMatch ? (
              // Show loading animation while searching
              <p className="inline-block mb-4">
                {[0, 1, 2].map((i) => (
                  <motion.span
                    key={i}
                    animate={{ y: [0, -5, 0] }}
                    transition={{
                      duration: 1.2,
                      repeat: Infinity,
                      ease: "easeInOut",
                      delay: i * 0.25,
                    }}
                    className="inline-block"
                  >
                    .&nbsp;
                  </motion.span>
                ))}
              </p>
            ) : null}

            {countdown === null && (
              <button
                onClick={handlers.handleCancelFindMatch}
                className="px-4 py-2 bg-red-500 text-black rounded hover:bg-red-600 transition"
              >
                Cancel
              </button>
            )}
          </div>
        )}

        {countdown !== null && (
          <div className="text-center mt-4">
            <p className="text-xl font-bold text-green-500">
              Starting in {countdown}...
            </p>
          </div>
        )}
      </Popup>
    </div>
  );
}

export default memo(Lobby)