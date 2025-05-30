import { getSocket } from "./socket.js";
import { useSocket } from "../socket/socketContext.jsx";
import { use } from "react";

export class LobbyHandlers {
    constructor(states, setters, socketRef, navigate) {
        this.states = states;
        this.setters = setters;
        this.socketRef = socketRef;
        this.navigate = navigate;
    }

    sendPayload = (payload) => {
        if (!this.socketRef) {
            this.getOrCreateSocket();
        }

        if (!this.socketRef) {
            alert("WebSocket is not connected.");
            return;
        }

        if (this.socketRef.readyState === WebSocket.OPEN) {
            this.socketRef.send(JSON.stringify(payload));
        } else if (this.socketRef.readyState === WebSocket.CONNECTING) {
            this.socketRef.onopen = () => {
                this.socketRef.send(JSON.stringify(payload));
            };
        } else {
            alert("WebSocket connection is not ready. Please refresh the page.");
        }
    };

    handleButtonClick = (label) => {
        if (!this.states.playerName.trim()) {
            alert("Please enter your name.");
            return;
        }

        this.setters.setSelectedLabel(label);
        this.setters.setShowPopup(true);

        // Reset UI states
        this.setters.setStatus("");
        this.setters.setRoomCode("");
        this.setters.setPlayerCount(1);
        this.setters.setIsHost(false);
        this.setters.setCountdown(null);
        this.setters.setShowWaitingForHost(false);
        this.setters.setShowJoinCodeInput(false);
        this.setters.setIsFindingMatch(false);

        if (label === "Join Match") {
            this.setters.setShowJoinCodeInput(true);
            return;
        }

        if (label === "Find Match") {
            this.setters.setIsFindingMatch(true);
        }

        let payload;
        if (label === "Create Match") {
            payload = { action: "create", name: this.states.playerName.trim() };
        } else if (label === "Find Match") {
            payload = { action: "find_match", name: this.states.playerName.trim() };
        }

        if (payload) {
            this.sendPayload(payload);
        }
    };

    handleJoinRoom = () => {
        if (!this.states.joinRoomCode.trim()) {
            alert("Please enter a room code.");
            return;
        }

        this.setters.setShowJoinCodeInput(false);
        this.sendPayload({
            action: "join",
            name: this.states.playerName.trim(),
            room: this.states.joinRoomCode.trim()
        });
    };

    handleStartGame = () => {
        this.sendPayload({
            action: "start_game",
            name: this.states.playerName.trim(),
            room: this.states.joinRoomCode.trim()
        });
    };

    handleClosePopup = () => {
        this.setters.setShowPopup(false);
        this.setters.setSelectedLabel('');
        this.setters.setShowJoinCodeInput(false);
        this.setters.setShowWaitingForHost(false);
        this.setters.setJoinRoomCode('');
        this.setters.setRoomCode('');
        this.setters.setPlayerCount(1);
        this.setters.setIsHost(false);
        this.setters.setIsNewHost(false);
        this.setters.setStatus("");
        this.setters.setIsFindingMatch(false);
        this.setters.setCountdown(null);

        this.sendPayload({
            action: "leave_room",
            name: this.states.playerName.trim(),
            room: this.states.joinRoomCode.trim()
        });
    };

    handleCancelFindMatch = () => {
        this.setters.setIsFindingMatch(false);
        this.setters.setSelectedLabel('');
        this.setters.setShowPopup(false);
        this.setters.setStatus("");
        this.setters.setCountdown(null);
        this.setters.setRoomCode("");
        this.setters.setIsHost(false);
        this.setters.setIsNewHost(false);
        this.setters.setPlayerCount(1);

        this.sendPayload({
            action: "cancel_find_match",
            name: this.states.playerName.trim()
        });
    };

    startCountdown = (seconds) => {
        this.setters.setCountdown(seconds);
        const interval = setInterval(() => {
            this.setters.setCountdown((prev) => {
                if (prev === 1) {
                    clearInterval(interval);
                    return null;
                }
                return prev - 1;
            });
        }, 1000);
    };

    getOrCreateSocket = () => {
        // const context = useSocket();
        // this.socketRef = context;

        this.socketRef.onopen = () => {
            console.log("WebSocket connected");
        };

        this.socketRef.onmessage = (event) => {
            console.log("Received message:", event.data);
            try {
                const data = JSON.parse(event.data);
                if (data.error) {
                    alert(data.error);
                    if (this.socketRef) this.socketRef.close();
                    this.setters.setShowPopup(false);
                    return;
                }

                this.handleSocketMessage(data);
            } catch (error) {
                console.error("Error parsing message:", error);
            }
        };

        this.socketRef.onerror = (error) => {
            console.error("WebSocket error:", error);
            alert("Connection error. Please try again.");
            this.setters.setShowPopup(false);
        };

        this.socketRef.onclose = (event) => {
            console.log("WebSocket closed:", event.code, event.reason);
            this.socketRef = null;
        };
    };

    handleSocketMessage = (data) => {
        switch (data.type) {
            case "room_created":
                this.setters.setIsHost(true);
                this.setters.setIsNewHost(false);
                this.setters.setRoomCode(data.roomCode);
                this.setters.setPlayerCount(1);
                break;
            case "joined":
                this.setters.setIsHost(false);
                this.setters.setShowJoinCodeInput(false);
                this.setters.setShowWaitingForHost(true);
                break;
            case "searching":
                this.setters.setStatus("Finding Match");
                break;
            case "match_found":
                this.setters.setRoomCode(data.roomCode);
                this.setters.setIsHost(data.isHost);
                this.setters.setIsNewHost(false);
                this.setters.setPlayerCount(data.playerCount || 2);
                this.setters.setStatus("Match found! Game will start soon...");
                this.setters.setIsFindingMatch(false);
                this.startCountdown(5);
                break;
            case "waiting":
                this.setters.setPlayerCount(data.playerCount);
                break;
            case "room_destroyed":
                alert(data.message || "The room has been closed by the host.");
                this.handleClosePopup();
                break;
            case "left_room":
                this.setters.setRoomCode('');
                this.setters.setPlayerCount(1);
                this.setters.setIsHost(false);
                this.setters.setStatus('');
                break;
            case "start":
                console.log("Game starting...:", data);
                this.navigate("/game", { state: { players: data.players, roomId: data.roomCode, playerName : this.states.playerName.trim()} });
                break;
            case "host_changed":
                this.setters.setIsHost(data.isHost);
                this.setters.setIsNewHost(true);
                this.setters.setShowWaitingForHost(false);
                this.setters.setRoomCode(data.roomCode);
                alert(data.message);
                break;
            default:
                console.warn("Unknown message type:", data.type);
                break;
        }
    };
}