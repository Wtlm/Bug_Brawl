package main

import (
	"log"

	"github.com/gorilla/websocket"
)

func handleCreateRoom(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	roomCode := generateRoomCode()
	// Ensure room code is unique
	for {
		if _, exists := clientsPerRoom[roomCode]; !exists {
			break
		}
		roomCode = generateRoomCode()
	}

	client.roomCode = roomCode
	client.isHost = true
	clientsPerRoom[roomCode] = []*Client{client}

	log.Printf("Room %s created by %s\n", roomCode, client.name)

	err := conn.WriteJSON(map[string]interface{}{
		"type":     "room_created",
		"roomCode": roomCode,
	})
	if err != nil {
		log.Printf("Error sending room_created message: %v\n", err)
	}
}

func handleJoinRoom(client *Client, conn *websocket.Conn, msg Message) {
	removePlayerFromQueue(client)

	if msg.Room == "" {
		conn.WriteJSON(map[string]string{"error": "Room code required to join"})
		clientsMutex.Unlock()
		return
	}
	roomClients, exists := clientsPerRoom[msg.Room]
	if !exists {
		conn.WriteJSON(map[string]string{"error": "Room does not exist"})
		clientsMutex.Unlock()
		return
	}
	if len(roomClients) >= 4 {
		conn.WriteJSON(map[string]string{"error": "Room is full"})
		clientsMutex.Unlock()
		return
	}

	client.roomCode = msg.Room
	client.isHost = false
	clientsPerRoom[msg.Room] = append(clientsPerRoom[msg.Room], client)

	conn.WriteJSON(map[string]interface{}{
		"type": "joined",
	})

	log.Printf("%s joined room %s\n", client.name, msg.Room)
	broadcastPlayerCount(client.roomCode, len(clientsPerRoom[client.roomCode]))
}

func handleFindMatch(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	err := conn.WriteJSON(map[string]string{
		"type": "searching",
	})
	if err != nil {
		log.Printf("Error sending searching message: %v\n", err)
	}
	client.isHost = false
	addToMatchQueue(client)
}

func handleStartGame(client *Client, conn *websocket.Conn) {
	if client.roomCode == "" {
		conn.WriteJSON(map[string]string{"error": "Not in any room"})
		clientsMutex.Unlock()
		return
	}

	roomClients, exists := clientsPerRoom[client.roomCode]
	if !exists {
		conn.WriteJSON(map[string]string{"error": "Room does not exist"})
		clientsMutex.Unlock()
		return
	}

	if !client.isHost {
		conn.WriteJSON(map[string]string{"error": "Only the host can start the game"})
		clientsMutex.Unlock()
		return
	}

	if len(roomClients) < 2 {
		conn.WriteJSON(map[string]string{"error": "Need at least 2 players to start"})
		clientsMutex.Unlock()
		return
	}

	startGame(client.roomCode)
	log.Printf("Host %s started the game in room %s\n", client.name, client.roomCode)
}

func handleCancelFindMatch(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	client.roomCode = ""
	client.isHost = false
	err := conn.WriteJSON(map[string]string{
		"type": "cancelled",
	})
	if err != nil {
		log.Printf("Error sending cancelled message: %v\n", err)
	}
	log.Printf("Cancelled find match for %s\n", client.name)
	conn.WriteJSON(map[string]string{"type": "find_match_cancelled"})
}

func handleLeaveRoom(client *Client, conn *websocket.Conn) {
	if client.roomCode != "" {
		removeClientFromRoom(client)
		client.isHost = false
		client.roomCode = ""
		conn.WriteJSON(map[string]string{"type": "left_room"})
		log.Printf("%s left the room\n", client.name)
	}
}
