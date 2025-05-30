package main

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func generateClientID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 5

	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	var id strings.Builder
	for i := 0; i < length; i++ {
		id.WriteByte(charset[rng.Intn(len(charset))])
	}
	return id.String()
}

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
		if c.Conn != nil && c.Conn.WriteMessage(websocket.PingMessage, nil) == nil {
			activeQueue = append(activeQueue, c)
		} else {
			log.Printf("Dropped inactive client: %s\n", c.Name)
		}
	}
	matchQueue = activeQueue

	matchQueue = append(matchQueue, client)
	log.Printf("Added %s to match queue. Queue length: %d\n", client.Name, len(matchQueue))

	limit := 2

	log.Printf("Current match queue: %v\n", func() []string {
		names := []string{}
		for _, c := range matchQueue {
			names = append(names, c.Name)
		}
		return names
	}())

	err := client.Conn.WriteJSON(map[string]interface{}{
		"type": "searching",
	})
	if err != nil {
		log.Printf("Error sending searching message to %s: %v\n", client.Name, err)
	}

	if len(matchQueue) >= limit {
		// Take the first 2 clients
		matched := matchQueue[:limit]
		matchQueue = matchQueue[limit:]

		roomCode := generateRoomCode()

		newRoom := &Room{
			Players:           matched,
			RoomCode:          roomCode,
			Question:          nil,
			SabotageSelection: nil,
			AnswerLog:         []*PlayerAnswer{},
			AvailableSabotages: map[string][]*Sabotage{
				client.ID: GenerateInitialSabotageList(),
			},
			PlayerEffects: map[string][]*Sabotage{
				client.ID: {},
			},
		}

		roomsMutex.Lock()
		rooms[roomCode] = newRoom
		roomsMutex.Unlock()
		// Ensure room code is unique
		// clientsMutex.Lock()
		// for {
		// 	if _, exists := clientsPerRoom[roomCode]; !exists {
		// 		break
		// 	}
		// 	roomCode = generateRoomCode()
		// }
		// clientsPerRoom[roomCode] = matched
		// // clientsMutex.Unlock()
		// clientsPerRoom[roomCode] = matched
		log.Printf("Match found for %d players in room %s\n", len(matched), roomCode)

		for i, c := range matched {
			c.Room = newRoom
			c.IsHost = (i == 0) // First player is host

			err := c.Conn.WriteJSON(map[string]interface{}{
				"type":        "match_found",
				"roomCode":    roomCode,
				"isHost":      c.IsHost,
				"playerCount": len(matched),
				"id":          c.ID,
			})
			if err != nil {
				log.Printf("Error sending match_found to %s: %v\n", c.Name, err)
			}
			broadcastPlayerCount(newRoom)
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
			roomClients := newRoom.Players
			if len(roomClients) > 0 {
				var host *Client
				for _, c := range roomClients {
					if c.IsHost {
						host = c
						break
					}
				}
				if host != nil {
					startGame(newRoom) // Pass the host client to startGame
					log.Printf("Game started in room %s with %d players\n", roomCode, len(roomClients))
				}
			} else {
				log.Printf("Room %s no longer exists, cannot start game\n", roomCode)
			}
		}()
	}
}

// func removeClientFromRoom(client *Client) {
// 	roomClients, exists := clientsPerRoom[client.RoomCode]
// 	if !exists {
// 		log.Printf("removeClientFromRoom: room %s does not exist for %s", client.RoomCode, client.Name)
// 		return
// 	}

// 	// If the host is leaving, notify all other players that the room is being destroyed
// 	if client.IsHost && len(roomClients) > 1 {
// 		// Notify all other players (excluding the host)
// 		for _, c := range roomClients {
// 			if c != client {
// 				c.Conn.WriteJSON(map[string]interface{}{
// 					"type":    "room_destroyed",
// 					"message": "The host has left the room. Room has been closed.",
// 				})
// 				// Also send left_room so client resets state
// 				c.Conn.WriteJSON(map[string]string{
// 					"type": "left_room",
// 				})
// 			}
// 		}
// 		log.Printf("Host %s left room %s - notifying %d other players\n", client.Name, client.RoomCode, len(roomClients)-1)
// 	}

// 	// Remove client from room
// 	for i, c := range roomClients {
// 		if c == client {
// 			clientsPerRoom[client.RoomCode] = append(roomClients[:i], roomClients[i+1:]...)
// 			break
// 		}
// 	}

// 	// If room is empty or host left, delete it
// 	if len(clientsPerRoom[client.RoomCode]) == 0 || client.IsHost {
// 		delete(clientsPerRoom, client.RoomCode)
// 		log.Printf("Room %s deleted\n", client.RoomCode)
// 	} else {
// 		// If a non-host left, make someone else host and broadcast updated player count
// 		if len(clientsPerRoom[client.RoomCode]) > 0 {
// 			clientsPerRoom[client.RoomCode][0].IsHost = true
// 			log.Printf("New host assigned in room %s\n", client.RoomCode)
// 		}
// 		// Broadcast updated player count
// 		broadcastPlayerCount(client.RoomCode, len(clientsPerRoom[client.RoomCode]))
// 	}
// }

func removeClientFromRoom(client *Client) {
	room := client.Room
	if room == nil {
		log.Printf("removeClientFromRoom: client %s is not in any room\n", client.Name)
		return
	}

	// Remove client from room first
	var remainingClients []*Client
	for _, c := range room.Players {
		if c != client {
			remainingClients = append(remainingClients, c)
		}
	}
	room.Players = remainingClients

	// If room is empty, delete it
	if len(remainingClients) == 0 {
		delete(rooms, room.RoomCode)
		log.Printf("Room %s deleted (empty)\n", room.RoomCode)
		return
	}

	// If the leaving client was the host, assign new host
	if client.IsHost {
		// Assign the next player as host
		remainingClients[0].IsHost = true
		log.Printf("New host assigned in room %s: %s\n", room.RoomCode, remainingClients[0].Name)

		// Notify all remaining players about the host change
		for _, c := range remainingClients {
			c.Conn.WriteJSON(map[string]interface{}{
				"type":     "host_changed",
				"isHost":   c == remainingClients[0],
				"roomCode": room.RoomCode,
				"message":  "A new host has been assigned.",
				"id":       c.ID,
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
			log.Printf("Removed %s from match queue (switching to create/join)\n", client.Name)
			break
		}
	}
	queueMutex.Unlock()
}

func broadcastPlayerCount(room *Room) {
	playerNames := []string{}
	for _, c := range room.Players {
		playerNames = append(playerNames, c.Name)
	}

	for _, c := range room.Players {
		c.Conn.WriteJSON(map[string]interface{}{
			"type":        "waiting",
			"playerCount": len(room.Players),
			"players":     playerNames,
			"id":          c.ID,
		})
	}
}

func startGame(room *Room) {
	// clientsMutex.Lock()
	// defer clientsMutex.Unlock()

	// if _, exists := clientsPerRoom[room]; !exists {
	// 	log.Printf("startGame: room %s does not exist\n", room)
	// 	return
	// }
	if room == nil {
		log.Println("startGame: room is nil")
		return
	}

	players := []map[string]interface{}{}
	for _, c := range room.Players {
		players = append(players, map[string]interface{}{
			"type":   "player_info",
			"id":     c.ID,
			"name":   c.Name,
			"health": c.Health,
		})
	}
	for _, c := range room.Players {
		c.Conn.WriteJSON(map[string]interface{}{
			"type":     "start",
			"players":  players,
			"roomCode": room.RoomCode,
		})
	}
	log.Printf("Game started in room %v\n", room)

	time.Sleep(2 * time.Second)
	room.StartQuestion()
}
