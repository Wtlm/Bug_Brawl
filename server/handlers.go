package main

import (
	"log"
	"slices"
	"time"

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

	// client.RoomCode = roomCode
	client.IsHost = true
	client.Health = 5 // Reset health for new room
	// clientsPerRoom[roomCode] = []*Client{client}
	newRoom := &Room{
		Players:           []*Client{client},
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
	// for _, c := range newRoom.Players{
	// 	newRoom.Players,
	// 	newRoom.AvailableSabotages[c.ID] = GenerateInitialSabotageList(),
	// 	newRoom.PlayerEffects[c.ID] = []*Sabotage{}
	// }

	roomsMutex.Lock()
	rooms[roomCode] = newRoom
	roomsMutex.Unlock()

	log.Printf("Room %s created by %s (%s)\n", roomCode, client.Name, client.ID)

	err := conn.WriteJSON(map[string]interface{}{
		"type":     "room_created",
		"roomCode": roomCode,
		"id":       client.ID,
		"name":     client.Name,
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

	// client.RoomCode = msg.Room
	client.IsHost = false
	client.Health = 5 // Reset health when joining a room
	room.Players = append(room.Players, client)

	// Fix: Initialize both available sabotages and effects
	room.AvailableSabotages[client.ID] = GenerateInitialSabotageList()
	room.PlayerEffects[client.ID] = []*Sabotage{} // Initialize empty effects

	// roomsMutex.RLock()
	// room := rooms[msg.Room]
	// roomsMutex.RUnlock()

	conn.WriteJSON(map[string]interface{}{
		"type": "joined",
	})

	log.Printf("%s (%s) joined room %s\n", client.Name, client.ID, msg.Room)
	broadcastPlayerCount(room)
}

func handleFindMatch(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	err := conn.WriteJSON(map[string]string{
		"type": "searching",
	})
	if err != nil {
		log.Printf("Error sending searching message: %v\n", err)
	}
	client.IsHost = false
	client.Health = 5 // Reset health when searching for a match
	addToMatchQueue(client)
}

func handleStartGame(client *Client, conn *websocket.Conn) {

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	room := client.Room
	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Not in any room"})
		return
	}

	// roomClients, exists := clientsPerRoom[client.RoomCode]
	// if !exists {
	// 	conn.WriteJSON(map[string]string{"error": "Room does not exist"})
	// 	clientsMutex.Unlock()
	// 	return
	// }

	if !client.IsHost {
		conn.WriteJSON(map[string]string{"error": "Only the host can start the game"})
		clientsMutex.Unlock()
		return
	}

	if len(room.Players) < 2 {
		conn.WriteJSON(map[string]string{"error": "Need at least 2 players to start"})
		clientsMutex.Unlock()
		return
	}

	startGame(room)
	log.Printf("Host %s started the game in room %s\n", client.Name, room.RoomCode)
}

func handleCancelFindMatch(client *Client, conn *websocket.Conn) {
	removePlayerFromQueue(client)

	client.Room = nil
	client.IsHost = false
	err := conn.WriteJSON(map[string]string{
		"type": "cancelled",
	})
	if err != nil {
		log.Printf("Error sending cancelled message: %v\n", err)
	}
	log.Printf("Cancelled find match for %s\n", client.Name)
	conn.WriteJSON(map[string]string{"type": "find_match_cancelled"})
}

func handleLeaveRoom(client *Client, conn *websocket.Conn) {
	if client.Room != nil {
		removeClientFromRoom(client)
		client.IsHost = false
		client.Room = nil
		conn.WriteJSON(map[string]string{"type": "left_room"})
		log.Printf("%s left the room\n", client.Name)
	}
}

func handleAnswer(client *Client, msg Message, conn *websocket.Conn) {
	if client.Health <= 0 {
		conn.WriteJSON(map[string]string{
			"error": "You've been eliminated and cannot answer anymore.",
		})
		return
	}

	// roomsMutex.RLock()
	room := rooms[msg.Room]
	// roomsMutex.RUnlock()

	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Room not found"})
		return
	}

	// Check correctness
	correct := false
	currentQuestion := room.Question
	log.Printf("Current question: %+v", room.Question)

	log.Printf("Player %s answered: %s", client.Name, msg.Answer)
	log.Printf("Current question answer: %s", currentQuestion.Answer)

	if msg.Answer == currentQuestion.Answer {
		correct = true
	}

	room.AnswerLog = append(room.AnswerLog, &PlayerAnswer{
		Client:     client,
		Answer:     msg.Answer,
		AnswerTime: msg.AnswerTime,
		Correct:    correct,
	})

	log.Printf("Player %s answered: %s (correct: %t)", client.Name, msg.Answer, correct)
	answerTimeoutMs := int64(30000) // 30 seconds
	now := time.Now().UnixMilli()

	if now-room.QuestionStart > answerTimeoutMs || len(room.AnswerLog) == len(room.Players) {

		for _, player := range room.Players {
			found := false
			for _, ans := range room.AnswerLog {
				if ans.Client == player {
					found = true
					break
				}
			}
			if !found {
				room.AnswerLog = append(room.AnswerLog, &PlayerAnswer{
					Client:     player,
					Answer:     "",
					Correct:    false,
					AnswerTime: 1<<63 - 1, // Max int64, so always slowest
				})
			}
		}
		if len(room.AnswerLog) == len(room.Players) {
			result := room.EvaluateRoundResults()
			if result.Winner == nil || result.Winner.Client == nil {
				log.Printf("No winner found in this round")
			}
			log.Printf("Losers: %+v", result.Losers)
			room.CalculateHealth(result.Winner, result.Losers)

			loserNames := []string{}
			for _, l := range result.Losers {
				if l != nil && l.Client != nil {
					loserNames = append(loserNames, l.Client.Name)
				}
			}
			winnerName := ""
			if result.Winner != nil && result.Winner.Client != nil {
				winnerName = result.Winner.Client.Name
			}

			for _, p := range room.Players {
				// p.connMutex.Lock()
				err := p.Conn.WriteJSON(map[string]interface{}{
					"type":   "round_result",
					"winner": winnerName,
					"losers": loserNames,
				})
				// p.connMutex.Unlock()
				if err != nil {
					log.Printf("Error sending round_result to %s: %v", p.ID, err)
				}

			}
			go func() {
				time.Sleep(3 * time.Second)
				room.AssignSabotagesToLosers(result)
			}()
		}

		// }
	}
}

func handleUseSabotage(winner *Client, msg Message, conn *websocket.Conn) {
	room := winner.Room
	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Room not found"})
		return
	}

	// room.RoundMutex.Lock()
	// defer room.RoundMutex.Unlock()

	// Verify sabotage selection is in progress
	if room.SabotageSelection == nil || room.SabotageSelection.WinnerID != winner.ID {
		conn.WriteJSON(map[string]string{"error": "Not allowed to use sabotage"})
		return
	}

	sabotageName := msg.Name
	targetInfos := []map[string]string{}

	// Apply to all losers (those in the choices map)
	for playerID := range room.SabotageSelection.Choices {
		targetName := "Unknown"
		for _, client := range room.Players {
			if client.ID == playerID {
				targetName = client.Name
				break
			}
		}

		targetInfos = append(targetInfos, map[string]string{
			"id":   playerID,
			"name": targetName,
		})

		// Store the sabotage effect
		s := &Sabotage{
			Name:     sabotageName,
			Used:     true,
			TargetID: playerID,
			UsedByID: winner.ID,
		}

		room.PlayerEffects[playerID] = append(room.PlayerEffects[playerID], s)
		// Remove the used sabotage from the player's available sabotages
		for i, sab := range room.AvailableSabotages[playerID] {
			if sab.Name == sabotageName {
				room.AvailableSabotages[playerID] = slices.Delete(room.AvailableSabotages[playerID], i, i+1)
				break
			}
		}
		log.Printf("HANDLE player effect: %+v", room.PlayerEffects[playerID])
		log.Printf("available sabotages: %+v", room.AvailableSabotages[playerID])
		log.Printf("Applied sabotage %s from %s to %s", sabotageName, winner.ID, playerID)
	}

	// Notify all players in the room
	for _, c := range room.Players {

		c.Conn.WriteJSON(map[string]interface{}{
			"type":     "sabotage_applied",
			"sabotage": sabotageName,
			"usedBy":   winner.Name,
			"targets":  targetInfos,
		})
	}

	go func() {
		time.Sleep(3 * time.Second)
		room.StartQuestion()
	}()
}
