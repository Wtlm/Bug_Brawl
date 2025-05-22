import React, { useRef, useState } from 'react';
import { useNavigate } from "react-router-dom";
import { delay, motion } from "framer-motion";
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
  const socketRef = useRef(null);
  const navigate = useNavigate();

  const handleButtonClick = (label) => {
    if (!playerName.trim()) {
      alert("Please enter your name.");
      return;
    }

    setSelectedLabel(label);
    setShowPopup(true);

    if (label === "Join Match") {
      setShowJoinCodeInput(true);
      return;
    }

    // Connect to WebSocket for Create Match and Find Match
    connectWebSocket(label);
  };

  const connectWebSocket = (action) => {
    console.log(`Connecting WebSocket for action: ${action}`);
    
    // Check if WebSocket is supported
    if (!window.WebSocket) {
      console.error("WebSocket is not supported by this browser");
      alert("WebSocket is not supported by this browser");
      return;
    }

    try {
      socketRef.current = new WebSocket(`${location.origin.replace(/^http/, "ws")}/ws`);
      console.log("WebSocket object created, readyState:", socketRef.current.readyState);
    } catch (error) {
      console.error("Error creating WebSocket:", error);
      alert("Failed to create WebSocket connection");
      return;
    }

    socketRef.current.onopen = () => {
      console.log("WebSocket connected successfully, readyState:", socketRef.current.readyState);
      const payload = {
        action: action === "Create Match" ? "create" : "join",
        name: playerName.trim(),
        room: action === "Join Match" ? joinRoomCode.trim() : undefined,
      };
      console.log("Sending payload:", payload);
      try {
        socketRef.current.send(JSON.stringify(payload));
        console.log("Payload sent successfully");
      } catch (error) {
        console.error("Error sending payload:", error);
      }
    };

    socketRef.current.onmessage = (event) => {
      console.log("Received message:", event.data);
      try {
        const data = JSON.parse(event.data);
        console.log("Parsed data:", data);

        if (data.error) {
          console.error("Server error:", data.error);
          alert(data.error);
          socketRef.current.close();
          setShowPopup(false);
          return;
        }

        if (data.type === "room_created") {
          console.log("Room created:", data.roomCode);
          setIsHost(true);
          setRoomCode(data.roomCode);
          setPlayerCount(1);
        }

        if (data.type === "joined") {
          console.log("Joined room successfully");
          setIsHost(false);
          setShowJoinCodeInput(false);
          setShowWaitingForHost(true);
        }

        if (data.type === "waiting") {
          console.log("Player count updated:", data.playerCount);
          setPlayerCount(data.playerCount);
        }

        if (data.type === "start") {
          console.log("Game starting");
          navigate("/game");
        }
      } catch (error) {
        console.error("Error parsing message:", error);
      }
    };

    socketRef.current.onerror = (error) => {
      console.error("WebSocket error:", error);
      console.log("WebSocket readyState on error:", socketRef.current.readyState);
      alert("Connection error. Please try again.");
      setShowPopup(false);
    };

    socketRef.current.onclose = (event) => {
      console.log("WebSocket connection closed:", event.code, event.reason);
      console.log("Was clean:", event.wasClean);
      if (!event.wasClean) {
        console.error("WebSocket closed unexpectedly");
      }
    };

    // Add a timeout to check connection status
    setTimeout(() => {
      if (socketRef.current.readyState === WebSocket.CONNECTING) {
        console.error("WebSocket still connecting after 5 seconds");
        alert("Connection timeout. Please check if the server is running.");
      }
    }, 5000);
  };

  const handleJoinRoom = () => {
    if (!joinRoomCode.trim()) {
      alert("Please enter a room code.");
      return;
    }

    setShowJoinCodeInput(false);
    connectWebSocket("Join Match");
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
    
    if (socketRef.current) {
      socketRef.current.close();
    }
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
                className="bg-[#9f9f9f] rounded-lg m-3 px-4 py-1 w-2/3 text-black text-center !text-base"                placeholder="Room Code"
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
            <p className="text-lg mb-4">Finding Match</p>
            <p className="inline-block">
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
          </div>
        )}
      </Popup>
    </div>
  );
}