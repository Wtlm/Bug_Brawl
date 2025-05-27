package main

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func generateRoomCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 4

	// Create a new random source for each call
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	var code strings.Builder
	for i := 0; i < length; i++ {
		code.WriteByte(charset[rng.Intn(len(charset))])
	}
	return code.String()
}

func addToMatchQueue(client *Client) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// Remove any dead clients
	activeQueue := []*Client{}
	for _, c := range matchQueue {
		if c.conn != nil && c.conn.WriteMessage(websocket.PingMessage, nil) == nil {
			activeQueue = append(activeQueue, c)
		} else {
			log.Printf("Dropped inactive client: %s\n", c.name)
		}
	}
	matchQueue = activeQueue

	matchQueue = append(matchQueue, client)
	log.Printf("Added %s to match queue. Queue length: %d\n", client.name, len(matchQueue))

	limit := 2

	log.Printf("Current match queue: %v\n", func() []string {
		names := []string{}
		for _, c := range matchQueue {
			names = append(names, c.name)
		}
		return names
	}())

	err := client.conn.WriteJSON(map[string]interface{}{
		"type": "searching",
	})
	if err != nil {
		log.Printf("Error sending searching message to %s: %v\n", client.name, err)
	}

	if len(matchQueue) >= limit {
		// Take the first 2 clients
		matched := matchQueue[:limit]
		matchQueue = matchQueue[limit:]

		roomCode := generateRoomCode()

		room := &Room{
			Code:    roomCode,
			Clients: matched,
		}
		rooms[roomCode] = room

		for i, c := range matched {
			c.room = room
			c.isHost = (i == 0)

			err := c.conn.WriteJSON(map[string]interface{}{
				"type":        "match_found",
				"roomCode":    room.Code,
				"isHost":      c.isHost,
				"playerCount": len(room.Clients),
			})
			if err != nil {
				log.Printf("Error sending match_found to %s: %v\n", c.name, err)
			}
			broadcastPlayerCount(room)
		}

		// Start the game after a 5-second delay to allow players to see the match found message
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Recovered in goroutine: %v", r)
				}
			}()
			time.Sleep(5 * time.Second)
			clientsMutex.Lock()
			defer clientsMutex.Unlock()

			// Check if room still exists before starting game
			if _, exists := rooms[room.Code]; exists {
				startGame(room)
				log.Printf("Game started in room %s with\n", room.Code)
			} else {
				log.Printf("Room %s no longer exists, cannot start game\n", room.Code)
			}
		}()
	}
}

func removeClientFromRoom(client *Client) {
	room := client.room
	if room == nil {
		log.Printf("removeClientFromRoom: room does not exist for %s", client.name)
		return
	}

	// Remove client from room first
	var remainingClients []*Client
	for _, c := range room.Clients {
		if c != client {
			remainingClients = append(remainingClients, c)
		}
	}
	room.Clients = remainingClients

	// If room is empty, delete it
	if len(remainingClients) == 0 {
		delete(rooms, room.Code)
		log.Printf("Room %s deleted (empty)\n", room.Code)
		return
	}

	// If the leaving client was the host, assign new host
	if client.isHost {
		// Assign the next player as host
		room.Clients[0].isHost = true
		log.Printf("New host assigned in room %s: %s\n", room.Code, room.Clients[0].name)

		// Notify all remaining players about the host change
		for _, c := range remainingClients {
			c.conn.WriteJSON(map[string]interface{}{
				"type":     "host_changed",
				"isHost":   c == room.Clients[0],
				"roomCode": room.Code,
				"message":  "A new host has been assigned.",
			})
		}
	}

	// Broadcast updated player count to remaining players
	broadcastPlayerCount(room)
}

func removePlayerFromQueue(client *Client) {
	queueMutex.Lock()
	for i, queuedClient := range matchQueue {
		if queuedClient == client {
			matchQueue = append(matchQueue[:i], matchQueue[i+1:]...)
			log.Printf("Removed %s from match queue (switching to create/join)\n", client.name)
			break
		}
	}
	queueMutex.Unlock()
}

func broadcastPlayerCount(room *Room) {
	playerNames := []string{}
	for _, c := range room.Clients {
		playerNames = append(playerNames, c.name)
	}

	for _, c := range room.Clients {
		c.conn.WriteJSON(map[string]interface{}{
			"type":        "waiting",
			"playerCount": len(room.Clients),
			"players":     playerNames,
		})
	}
}

func startGame(room *Room) {
	players := []map[string]interface{}{}
	for _, c := range room.Clients {
		players = append(players, map[string]interface{}{
			"name":   c.name,
			"health": 5,
		})
	}
	for _, c := range room.Clients {
		c.conn.WriteJSON(map[string]interface{}{
			"type":    "start",
			"players": players,
		})
	}
	log.Printf("Game started in room %v\n", room)
}
