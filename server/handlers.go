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

	// client.roomCode = roomCode
	client.isHost = true
	// clientsPerRoom[roomCode] = []*Client{client}
	newRoom := &Room{
		Players:           []*Client{client},
		RoomCode:          roomCode,
		Question:          nil,
		SabotageSelection: nil,
		AnswerLog:         []*PlayerAnswer{},
		AvailableSabotages: map[string][]*Sabotage{
			client.id: GenerateInitialSabotageList(),
		},
		PlayerEffects: map[string][]*Sabotage{
			client.id: {},
		},
	}
	// for _, c := range newRoom.Players{
	// 	newRoom.Players,
	// 	newRoom.AvailableSabotages[c.id] = GenerateInitialSabotageList(),
	// 	newRoom.PlayerEffects[c.id] = []*Sabotage{}
	// }

	roomsMutex.Lock()
	rooms[roomCode] = newRoom
	roomsMutex.Unlock()

	log.Printf("Room %s created by %s (%s)\n", roomCode, client.name, client.id)

	err := conn.WriteJSON(map[string]interface{}{
		"type":     "room_created",
		"roomCode": roomCode,
		"id":       client.id,
		"name":     client.name,
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

	// client.roomCode = msg.Room
	client.isHost = false
	room.Players = append(room.Players, client)
	room.PlayerEffects[client.id] = GenerateInitialSabotageList()

	// roomsMutex.RLock()
	// room := rooms[msg.Room]
	// roomsMutex.RUnlock()

	conn.WriteJSON(map[string]interface{}{
		"type": "joined",
	})

	log.Printf("%s (%s) joined room %s\n", client.name, client.id, msg.Room)
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
	client.isHost = false
	addToMatchQueue(client)
}

func handleStartGame(client *Client, conn *websocket.Conn) {

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	room := client.room
	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Not in any room"})
		return
	}

	// roomClients, exists := clientsPerRoom[client.roomCode]
	// if !exists {
	// 	conn.WriteJSON(map[string]string{"error": "Room does not exist"})
	// 	clientsMutex.Unlock()
	// 	return
	// }

	if !client.isHost {
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
		client.isHost = false
		client.room = nil
		conn.WriteJSON(map[string]string{"type": "left_room"})
		log.Printf("%s left the room\n", client.name)
	}
}

func handleAnswer(client *Client, msg Message, conn *websocket.Conn) {
	if client.Health <= 0 {
		conn.WriteJSON(map[string]string{
			"error": "You've been eliminated and cannot answer anymore.",
		})
		return
	}

	roomsMutex.RLock()
	room := client.room
	roomsMutex.RUnlock()

	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Room not found"})
		return
	}

	room.RoundMutex.Lock()
	defer room.RoundMutex.Unlock()

	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now-room.QuestionStart > 30000 { // 30 seconds timeout
		conn.WriteJSON(map[string]string{"error": "Answer time expired"})
		return
	}

	// Check if client already answered
	for _, ans := range room.AnswerLog {
		if ans.Client == client {
			conn.WriteJSON(map[string]string{"error": "Already answered"})
			return
		}
	}

	// Check correctness
	correct := false
	currentQuestion := room.Question
	if msg.Answer == currentQuestion.Answer {
		correct = true
	}

	room.AnswerLog = append(room.AnswerLog, &PlayerAnswer{
		Client:     client,
		Answer:     msg.Answer,
		AnswerTime: msg.AnswerTime,
		Correct:    correct,
	})
	if len(room.AnswerLog) == len(room.Players) {
		result := room.EvaluateRoundResults()
		room.CalculateHealth(result.Winner, result.Losers)

		// Broadcast round result
		loserNames := []string{}
		for _, l := range result.Losers {
			if l.Client != nil {
				loserNames = append(loserNames, l.Client.name)
			}
		}

		for _, p := range room.Players {
			p.conn.WriteJSON(map[string]interface{}{
				"type":   "round_result",
				"winner": result.Winner.Client.name,
				"losers": loserNames,
			})
		}

		room.AssignSabotagesToLosers(result)
	}

}

func handleUseSabotage(winner *Client, msg Message, conn *websocket.Conn) {
	room := winner.room
	if room == nil {
		conn.WriteJSON(map[string]string{"error": "Room not found"})
		return
	}

	room.RoundMutex.Lock()
	defer room.RoundMutex.Unlock()

	// Verify sabotage selection is in progress
	if room.SabotageSelection == nil || room.SabotageSelection.WinnerID != winner.id {
		conn.WriteJSON(map[string]string{"error": "Not allowed to use sabotage"})
		return
	}

	sabotageName := msg.Name
	targetInfos := []map[string]string{}

	// Apply to all losers (those in the choices map)
	for playerID := range room.SabotageSelection.Choices {
		targetName := "Unknown"
		for _, client := range room.Players {
			if client.id == playerID {
				targetName = client.name
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
			UsedByID: winner.id,
		}

		room.PlayerEffects[playerID] = append(room.PlayerEffects[playerID], s)
		// Remove the used sabotage from the player's available sabotages
		for i, sab := range room.AvailableSabotages[playerID] {
			if sab.Name == sabotageName {
				room.AvailableSabotages[playerID] = slices.Delete(room.AvailableSabotages[playerID], i, i+1)
				break
			}
		}

		log.Printf("Applied sabotage %s from %s to %s", sabotageName, winner.id, playerID)
	}

	// Notify all players in the room
	for _, c := range room.Players {
		c.conn.WriteJSON(map[string]interface{}{
			"type":     "sabotage_applied",
			"sabotage": sabotageName,
			"usedBy":   winner.name,
			"targets":  targetInfos,
		})
	}

	room.StartQuestion()
}
