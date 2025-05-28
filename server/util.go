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

		newRoom := &Room{
			Players: matched,
			// Host:               client,
			RoomCode:           roomCode,
			Question:           nil,
			SabotageSelection:  nil,
			AnswerLog:          []*PlayerAnswer{},
			AvailableSabotages: make(map[string][]*Sabotage),
			PlayerEffects:      make(map[string][]*Sabotage),
		}
		rooms[roomCode] = newRoom
		// for _, c := range clientsPerRoom[roomCode] {
		// 	newRoom.Players[c] = true
		// 	newRoom.AvailableSabotages[c.id] = GenerateInitialSabotageList()
		// 	newRoom.PlayerEffects[c.id] = []*Sabotage{}
		// }

		// roomsMutex.Lock()
		// rooms[roomCode] = newRoom
		// roomsMutex.Unlock()
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
		// clientsPerRoom[roomCode] = matched
		log.Printf("Match found for %d players in room %s\n", len(matched), roomCode)

		for i, c := range matched {
			c.room = newRoom
			c.isHost = (i == 0) // First player is host

			err := c.conn.WriteJSON(map[string]interface{}{
				"type":        "match_found",
				"roomCode":    roomCode,
				"isHost":      c.isHost,
				"playerCount": len(matched),
				"id":          c.id,
			})
			if err != nil {
				log.Printf("Error sending match_found to %s: %v\n", c.name, err)
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
			// if roomClients, exists := rooms[newRoom.RoomCode]; exists {
			if _, exists := rooms[newRoom.RoomCode]; exists {
				// var host *Client
				// for _, c := range roomClients {
				// 	if c.isHost {
				// 		host = c
				// 		break
				// 	}
				// }
				// if host != nil {
				startGame(newRoom) // Pass the host client to startGame
				log.Printf("Game started in room %s with %d players\n", roomCode, len(newRoom.Players))
				// }
			} else {
				log.Printf("Room %s no longer exists, cannot start game\n", roomCode)
			}
		}()
	}
}

func removeClientFromRoom(client *Client) {
	// roomClients, exists := clientsPerRoom[client.roomCode]
	room := client.room
	if room == nil {
		log.Printf("removeClientFromRoom: room %s does not exist for %s", client.room.RoomCode, client.name)
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
		delete(rooms, client.room.RoomCode)
		log.Printf("Room %s deleted (empty)\n", room.RoomCode)
		return
	}

	// If the leaving client was the host, assign new host
	if client.isHost {
		// Assign the next player as host
		remainingClients[0].isHost = true
		// remainingClients[0].roomCode = client.roomCode
		log.Printf("New host assigned in room %s: %s\n", room.RoomCode, remainingClients[0].name)

		// Notify all remaining players about the host change
		for _, c := range remainingClients {
			c.conn.WriteJSON(map[string]interface{}{
				"type":     "host_changed",
				"isHost":   c == remainingClients[0],
				"roomCode": room.RoomCode,
				"message":  "A new host has been assigned.",
				"id":       c.id,
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
	for _, c := range room.Players {
		playerNames = append(playerNames, c.name)
	}

	for _, c := range room.Players {
		c.conn.WriteJSON(map[string]interface{}{
			"type":        "waiting",
			"playerCount": len(room.Players),
			"players":     playerNames,
			"id":          c.id,
		})
	}
}

func startGame(room *Room) {
	// clientsMutex.Lock()
	// defer clientsMutex.Unlock()

	// roomClients, exists := clientsPerRoom[room.RoomCode]
	// if !exists {
	// 	log.Printf("startGame: room %s does not exist in clientsPerRoom\n", room.RoomCode)
	// 	return
	// }

	players := []map[string]interface{}{}
	for _, c := range room.Players {
		players = append(players, map[string]interface{}{
			"id":     c.id,
			"name":   c.name,
			"health": 5,
		})
	}

	for _, c := range room.Players {
		c.conn.WriteJSON(map[string]interface{}{
			"type":    "start",
			"players": players,
		})
	}

	log.Printf("Game started in room %s\n", room.RoomCode)
	room.StartQuestion()
}
