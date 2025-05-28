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
		if _, exists := rooms[roomCode]; !exists {
			break
		}
		roomCode = generateRoomCode()
	}

	newRoom := &Room{
		Players: []*Client{client},
		// Host:               client,
		RoomCode:           roomCode,
		Question:           nil,
		SabotageSelection:  nil,
		AnswerLog:          []*PlayerAnswer{},
		AvailableSabotages: make(map[string][]*Sabotage),
		PlayerEffects:      make(map[string][]*Sabotage),
	}

	client.room = newRoom
	client.isHost = true
	rooms[roomCode] = newRoom

	// for _, c := range rooms[roomCode] {
	// 	newRoom.Players[c] = true
	// 	newRoom.AvailableSabotages[c.id] = GenerateInitialSabotageList()
	// 	newRoom.PlayerEffects[c.id] = []*Sabotage{}
	// }

	// roomsMutex.Lock()
	// rooms[roomCode] = newRoom
	// roomsMutex.Unlock()

	log.Printf("Room %s created by %s (%s)\n", roomCode, client.name, client.id)

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
	if len(room.Players) >= 4 {
		conn.WriteJSON(map[string]string{"error": "Room is full"})
		clientsMutex.Unlock()
		return
	}

	client.room = room
	client.isHost = false
	room.Players = append(room.Players, client)

	// roomsMutex.RLock()
	// room := rooms[msg.Room]
	// roomsMutex.RUnlock()

	if room != nil {
		room.PlayerEffects[client.id] = GenerateInitialSabotageList()
	}

	conn.WriteJSON(map[string]interface{}{
		"type": "joined",
		"id":   client.id,
		"name": client.name,
	})

	log.Printf("%s (%s) joined room %s\n", client.name, client.id, msg.Room)
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
	// roomsMutex.RLock()
	// room := rooms[msg.Room]
	// roomsMutex.RUnlock()

	// clientsMutex.Lock()
	// defer clientsMutex.Unlock()
	room := client.room

	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Not in any room"})
		// clientsMutex.Unlock()
		return
	}

	// roomClients, exists := [client.roomCode]
	// if !exists {
	// 	conn.WriteJSON(map[string]string{"error": "Room does not exist"})
	// 	// clientsMutex.Unlock()
	// 	return
	// }

	if !client.isHost {
		conn.WriteJSON(map[string]string{"error": "Only the host can start the game"})
		// clientsMutex.Unlock()
		return
	}

	if len(room.Players) < 2 {
		conn.WriteJSON(map[string]string{"error": "Need at least 2 players to start"})
		// clientsMutex.Unlock()
		return
	}

	startGame(room)
	log.Printf("Host %s started the game in room %s\n", client.name, room.RoomCode)
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
	client.isHost = false
	client.room = nil
}

// func handleAnswer(client *Client, msg Message, conn *websocket.Conn) {
// 	roomsMutex.RLock()
// 	room := rooms[client.roomCode]
// 	roomsMutex.RUnlock()

// 	if client.roomCode == "" {
// 		conn.WriteJSON(map[string]string{"error": "Not in any room"})
// 		clientsMutex.Unlock()
// 		return
// 	}

// 	if room == nil {
// 		conn.WriteJSON(map[string]string{"error": "Room not found"})
// 		return
// 	}

// 	room.RoundMutex.Lock()
// 	defer room.RoundMutex.Unlock()

// 	now := time.Now().UnixNano() / int64(time.Millisecond)
// 	if now-room.QuestionStart > 30000 { // 30 seconds timeout
// 		conn.WriteJSON(map[string]string{"error": "Answer time expired"})
// 		return
// 	}

// 	// Check if client already answered
// 	for _, ans := range room.AnswerLog {
// 		if ans.Client == client {
// 			conn.WriteJSON(map[string]string{"error": "Already answered"})
// 			return
// 		}
// 	}

// 	// Check correctness
// 	correct := false
// 	currentQuestion := room.Question
// 	if msg.Answer == currentQuestion.Answer {
// 		correct = true
// 	}

// 	room.AnswerLog = append(room.AnswerLog, &PlayerAnswer{
// 		Client:     client,
// 		Answer:     msg.Answer,
// 		AnswerTime: msg.AnswerTime,
// 		Correct:    correct,
// 	})
// 	if len(room.AnswerLog) == len(room.Players) {
// 		result := room.EvaluateRoundResults()

// 		// Broadcast round result
// 		for p := range room.Players {
// 			p.conn.WriteJSON(map[string]interface{}{
// 				"type":    "round_result",
// 				"winner":  result.Winner.Client.name,
// 				"results": result,
// 			})
// 		}

// 		// Save result for sabotage step
// 		room.AssignSabotagesToLosers(result)

// 		// If only one player left → end game
// 		if len(room.Players) <= 1 {

// 			for p := range room.Players {
// 				p.conn.WriteJSON(map[string]string{
// 					"type":    "game_over",
// 					"message": "You are the last player standing!",
// 				})
// 			}
// 			return
// 		}

// 	}

// }

// func handleUseSabotage(winner *Client, msg Message, conn *websocket.Conn) {
// 	room := rooms[winner.roomCode]
// 	if room == nil {
// 		conn.WriteJSON(map[string]string{"error": "Room not found"})
// 		return
// 	}

// 	room.RoundMutex.Lock()
// 	defer room.RoundMutex.Unlock()

// 	// Verify sabotage selection is in progress
// 	if room.SabotageSelection == nil || room.SabotageSelection.WinnerID != winner.id {
// 		conn.WriteJSON(map[string]string{"error": "Not allowed to use sabotage"})
// 		return
// 	}

// 	sabotageName := msg.Name
// 	targetInfos := []map[string]string{}

// 	// Apply to all losers (those in the choices map)
// 	for playerID := range room.SabotageSelection.Choices {
// 		targetName := "Unknown"
// 		for client := range room.Players {
// 			if client.id == playerID {
// 				targetName = client.name
// 				break
// 			}
// 		}

// 		targetInfos = append(targetInfos, map[string]string{
// 			"id":   playerID,
// 			"name": targetName,
// 		})

// 		// Store the sabotage effect
// 		s := &Sabotage{
// 			Name:     sabotageName,
// 			Used:     true,
// 			TargetID: playerID,
// 			UsedByID: winner.id,
// 		}

// 		room.PlayerEffects[playerID] = append(room.PlayerEffects[playerID], s)
// 		// Remove the used sabotage from the player's available sabotages
// 		for i, sab := range room.AvailableSabotages[playerID] {
// 			if sab.Name == sabotageName {
// 				room.AvailableSabotages[playerID] = slices.Delete(room.AvailableSabotages[playerID], i, i+1)
// 				break
// 			}
// 		}

// 		log.Printf("Applied sabotage %s from %s to %s", sabotageName, winner.id, playerID)
// 	}

// 	// Notify all players in the room
// 	for c := range room.Players {
// 		c.conn.WriteJSON(map[string]interface{}{
// 			"type":     "sabotage_applied",
// 			"sabotage": sabotageName,
// 			"usedBy":   winner.name,
// 			"targets":  targetInfos,
// 		})
// 	}

// 	room.StartQuestion()
// }
