import React, { useRef, useState, useEffect } from 'react';
import { useNavigate } from "react-router-dom";
import { delay, motion } from "framer-motion";
import { getSocket } from "../socket/socket.js";
import Popup from "../widget/popup";

export default function Lobby() {
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
  const [status, setStatus] = useState("");
  const [isFindingMatch, setIsFindingMatch] = useState(false);
  const [countdown, setCountdown] = useState(null);
  const socketRef = useRef(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Create WebSocket once on component mount
    if (!socketRef.current) {
      getOrCreateSocket();
    }
    // Clean up WebSocket connection on component unmount
    return () => {
    };
  }, []);

  const sendPayload = (payload) => {
    if (!socketRef.current) {
      getOrCreateSocket();
    }

    if (!socketRef.current) {
      alert("WebSocket is not connected.");
      return;
    }

    if (socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify(payload));
    } else if (socketRef.current.readyState === WebSocket.CONNECTING) {
      // Send when socket opens
      socketRef.current.onopen = () => {
        socketRef.current.send(JSON.stringify(payload));
      };
    } else {
      alert("WebSocket connection is not ready. Please refresh the page.");
    }
  };

  const handleButtonClick = (label) => {
    if (!playerName.trim()) {
      alert("Please enter your name.");
      return;
    }

    setSelectedLabel(label);
    setShowPopup(true);

    // Reset UI states
    setStatus("");
    setRoomCode("");
    setPlayerCount(1);
    setIsHost(false);
    setCountdown(null);
    setShowWaitingForHost(false);
    setShowJoinCodeInput(false);
    setIsFindingMatch(false);

    if (label === "Join Match") {
      setShowJoinCodeInput(true);
      return;
    }

    if (label === "Find Match") {
      setIsFindingMatch(true);
    }

    // Prepare payload depending on label
    let payload;
    if (label === "Create Match") {
      payload = { action: "create", name: playerName.trim() };
    } else if (label === "Find Match") {
      payload = { action: "find_match", name: playerName.trim() };
    }

    if (payload) {
      sendPayload(payload);
    }
  };

  const handleJoinRoom = () => {
    if (!joinRoomCode.trim()) {
      alert("Please enter a room code.");
      return;
    }

    setShowJoinCodeInput(false);

    // Send join payload via existing socket connection
    sendPayload({ action: "join", name: playerName.trim(), room: joinRoomCode.trim() });
  };

  const handleStartGame = () => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ action: "start_game" }));
    }
  };

  const handleClosePopup = () => {
    setShowPopup(false);
    setSelectedLabel('');
    setShowJoinCodeInput(false);
    setShowWaitingForHost(false);
    setJoinRoomCode('');
    setRoomCode('');
    setPlayerCount(1);
    setIsHost(false);
    setStatus("");
    setIsFindingMatch(false);
    setCountdown(null);

    if (roomCode && socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ action: "leave_room" }));
    }
  };

  const handleCancelFindMatch = () => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      // Instead of closing, send cancel action
      socketRef.current.send(JSON.stringify({ action: "cancel_find_match" }));
    }

    setIsFindingMatch(false);
    setSelectedLabel('');
    setShowPopup(false);
    setStatus("");
    setCountdown(null);
    setRoomCode("");
    setIsHost(false);
    setPlayerCount(1);
  };

  const startCountdown = (seconds) => {
    setCountdown(seconds);
    const interval = setInterval(() => {
      setCountdown((prev) => {
        if (prev === 1) {
          clearInterval(interval);
          navigate("/game");
          return null;
        }
        return prev - 1;
      });
    }, 1000);
  };

  const getOrCreateSocket = () => {
    socketRef.current = getSocket();

    socketRef.current.onopen = () => {
      console.log("WebSocket connected");
    };

    socketRef.current.onmessage = (event) => {
      console.log("Received message:", event.data);
      try {
        const data = JSON.parse(event.data);
        if (data.error) {
          alert(data.error);
          if (socketRef.current) socketRef.current.close();
          setShowPopup(false);
          return;
        }

        switch (data.type) {
          case "room_created":
            setIsHost(true);
            setRoomCode(data.roomCode);
            setPlayerCount(1);
            break;
          case "joined":
            setIsHost(false);
            setShowJoinCodeInput(false);
            setShowWaitingForHost(true);
            break;
          case "searching":
            setStatus("Finding Match");
            break;
          case "match_found":
            setRoomCode(data.roomCode);
            setIsHost(data.isHost);
            setPlayerCount(data.playerCount || 2);
            setStatus("Match found! Game will start soon...");
            setIsFindingMatch(false);
            startCountdown(5);
            break;
          case "waiting":
            setPlayerCount(data.playerCount);
            break;
          case "room_destroyed":
            alert(data.message || "The room has been closed by the host.");
            handleClosePopup();
            break;
          case "left_room":
            setRoomCode('');
            setPlayerCount(1);
            setIsHost(false);
            setStatus('');
            break;
          case "start":
            navigate("/game", { state: { players: data.players } });
            break;
          default:
            console.warn("Unknown message type:", data.type);
            break;
        }
      } catch (error) {
        console.error("Error parsing message:", error);
      }
    };

    socketRef.current.onerror = (error) => {
      console.error("WebSocket error:", error);
      alert("Connection error. Please try again.");
      setShowPopup(false);
    };

    socketRef.current.onclose = (event) => {
      console.log("WebSocket closed:", event.code, event.reason);
      socketRef.current = null;
    };

  };

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
            onClick={() => handleButtonClick(label)}
          >
            {label}
          </motion.button>
        ))}
      </div>

      <Popup show={showPopup} onClose={handleClosePopup} className="lg:w-1/4 lg:h-2/5">
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
                  onClick={handleStartGame}
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
              onClick={handleJoinRoom}
            >
              Join Room
            </motion.button>
          </div>
        )}

        {showWaitingForHost && (
          <div className="flex flex-col justify-center text-center items-center h-full">
            <p className="text-lg">Waiting for host to start the game...</p>
            <p className="text-lg">Players: {playerCount}/4</p>
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
                <p className="text-lg">Room Code: <strong>{roomCode}</strong></p>
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
                onClick={handleCancelFindMatch}
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