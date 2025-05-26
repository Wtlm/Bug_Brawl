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

		// Ensure room code is unique
		// clientsMutex.Lock()
		// for {
		// 	if _, exists := clientsPerRoom[roomCode]; !exists {
		// 		break
		// 	}
		// 	roomCode = generateRoomCode()
		// }
		// clientsPerRoom[roomCode] = matched
		// clientsMutex.Unlock()
		clientsPerRoom[roomCode] = matched
		log.Printf("Match found for %d players in room %s\n", len(matched), roomCode)

		for i, c := range matched {
			c.roomCode = roomCode
			c.isHost = (i == 0) // First player is host

			err := c.conn.WriteJSON(map[string]interface{}{
				"type":        "match_found",
				"roomCode":    roomCode,
				"isHost":      c.isHost,
				"playerCount": len(matched),
			})
			if err != nil {
				log.Printf("Error sending match_found to %s: %v\n", c.name, err)
			}
			broadcastPlayerCount(roomCode, len(matched))
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
			if roomClients, exists := clientsPerRoom[roomCode]; exists {
				startGame(roomCode)
				log.Printf("Game started in room %s with %d players\n", roomCode, len(roomClients))
			} else {
				log.Printf("Room %s no longer exists, cannot start game\n", roomCode)
			}
		}()
	}
}

// func removeClientFromRoom(client *Client) {
// 	roomClients, exists := clientsPerRoom[client.roomCode]
// 	if !exists {
// 		log.Printf("removeClientFromRoom: room %s does not exist for %s", client.roomCode, client.name)
// 		return
// 	}

// 	// If the host is leaving, notify all other players that the room is being destroyed
// 	if client.isHost && len(roomClients) > 1 {
// 		// Notify all other players (excluding the host)
// 		for _, c := range roomClients {
// 			if c != client {
// 				c.conn.WriteJSON(map[string]interface{}{
// 					"type":    "room_destroyed",
// 					"message": "The host has left the room. Room has been closed.",
// 				})
// 				// Also send left_room so client resets state
// 				c.conn.WriteJSON(map[string]string{
// 					"type": "left_room",
// 				})
// 			}
// 		}
// 		log.Printf("Host %s left room %s - notifying %d other players\n", client.name, client.roomCode, len(roomClients)-1)
// 	}

// 	// Remove client from room
// 	for i, c := range roomClients {
// 		if c == client {
// 			clientsPerRoom[client.roomCode] = append(roomClients[:i], roomClients[i+1:]...)
// 			break
// 		}
// 	}

// 	// If room is empty or host left, delete it
// 	if len(clientsPerRoom[client.roomCode]) == 0 || client.isHost {
// 		delete(clientsPerRoom, client.roomCode)
// 		log.Printf("Room %s deleted\n", client.roomCode)
// 	} else {
// 		// If a non-host left, make someone else host and broadcast updated player count
// 		if len(clientsPerRoom[client.roomCode]) > 0 {
// 			clientsPerRoom[client.roomCode][0].isHost = true
// 			log.Printf("New host assigned in room %s\n", client.roomCode)
// 		}
// 		// Broadcast updated player count
// 		broadcastPlayerCount(client.roomCode, len(clientsPerRoom[client.roomCode]))
// 	}
// }

func removeClientFromRoom(client *Client) {
	roomClients, exists := clientsPerRoom[client.roomCode]
	if !exists {
		log.Printf("removeClientFromRoom: room %s does not exist for %s", client.roomCode, client.name)
		return
	}

	// Remove client from room first
	var remainingClients []*Client
	for _, c := range roomClients {
		if c != client {
			remainingClients = append(remainingClients, c)
		}
	}
	clientsPerRoom[client.roomCode] = remainingClients

	// If room is empty, delete it
	if len(remainingClients) == 0 {
		delete(clientsPerRoom, client.roomCode)
		log.Printf("Room %s deleted (empty)\n", client.roomCode)
		return
	}

	// If the leaving client was the host, assign new host
	if client.isHost {
		// Assign the next player as host
		remainingClients[0].isHost = true
		remainingClients[0].roomCode = client.roomCode
		log.Printf("New host assigned in room %s: %s\n", client.roomCode, remainingClients[0].name)

		// Notify all remaining players about the host change
		for _, c := range remainingClients {
			c.conn.WriteJSON(map[string]interface{}{
				"type":     "host_changed",
				"isHost":   c == remainingClients[0],
				"roomCode": client.roomCode,
				"message":  "A new host has been assigned.",
			})
		}
	}

	// Broadcast updated player count to remaining players
	broadcastPlayerCount(client.roomCode, len(remainingClients))
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

func broadcastPlayerCount(room string, count int) {
	playerNames := []string{}
	for _, c := range clientsPerRoom[room] {
		playerNames = append(playerNames, c.name)
	}

	for _, c := range clientsPerRoom[room] {
		c.conn.WriteJSON(map[string]interface{}{
			"type":        "waiting",
			"playerCount": count,
			"players":     playerNames,
		})
	}
}

func startGame(room string) {
	players := []map[string]interface{}{}
	for _, c := range clientsPerRoom[room] {
		players = append(players, map[string]interface{}{
			"name":   c.name,
			"health": 5,
		})
	}
	for _, c := range clientsPerRoom[room] {
		c.conn.WriteJSON(map[string]interface{}{
			"type":    "start",
			"players": players,
		})
	}
	log.Printf("Game started in room %s\n", room)
}
