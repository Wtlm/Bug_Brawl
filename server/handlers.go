package main

import (
	"log"

	"github.com/gorilla/websocket"
)

func handleCreateRoom(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	roomCode := generateRoomCode()
	room := &Room{
		Code:    roomCode,
		Clients: []*Client{client},
	}
	// Ensure room code is unique
	// for {
	// 	if _, exists := clientsPerRoom[roomCode]; !exists {
	// 		break
	// 	}
	// 	roomCode = generateRoomCode()
	// }

	client.room = room
	client.isHost = true
	rooms[roomCode] = room

	log.Printf("Room %s created by %s\n", roomCode, client.name)

	err := conn.WriteJSON(map[string]interface{}{
		"type": "room_created",
		"room": roomCode,
		"id":   client.id,
		"name": client.name,
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
	room, exists := rooms[msg.Room]
	if !exists {
		conn.WriteJSON(map[string]string{"error": "Room does not exist"})
		clientsMutex.Unlock()
		return
	}
	if len(room.Clients) >= 4 {
		conn.WriteJSON(map[string]string{"error": "Room is full"})
		clientsMutex.Unlock()
		return
	}

	client.room = room
	client.isHost = false
	room.Clients = append(room.Clients, client)

	conn.WriteJSON(map[string]interface{}{
		"type": "joined",
		"id":   client.id,
		"name": client.name,
	})

	log.Printf("%s joined room %s\n", client.name, room.Code)
	broadcastPlayerCount(room)
}

func handleFindMatch(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	client.isHost = false
	client.room = nil

	err := conn.WriteJSON(map[string]string{
		"type": "searching",
		"id":   client.id,
		"name": client.name,
	})
	if err != nil {
		log.Printf("Error sending searching message: %v\n", err)
	}
	client.isHost = false
	addToMatchQueue(client)
}

func handleStartGame(client *Client, conn *websocket.Conn) {
	room := client.room
	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Not in any room"})
		clientsMutex.Unlock()
		return
	}

	if !client.isHost {
		conn.WriteJSON(map[string]string{"error": "Only the host can start the game"})
		return
	}

	if len(room.Clients) < 2 {
		conn.WriteJSON(map[string]string{"error": "Need at least 2 players to start"})
		return
	}

	startGame(room)
	log.Printf("Host %s started the game in room %s\n", client.name, room.Code)
}

func handleCancelFindMatch(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	client.room = nil
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
	if client.room != nil {
		removeClientFromRoom(client)
		conn.WriteJSON(map[string]string{"type": "left_room"})
		log.Printf("%s left the room\n", client.name)
	}
	client.room = nil
	client.isHost = false
}
