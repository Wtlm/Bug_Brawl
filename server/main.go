package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Room   string `json:"room,omitempty"`
}

type Client struct {
	conn     *websocket.Conn
	name     string
	roomCode string
	isHost   bool
}

var (
	clientsPerRoom = make(map[string][]*Client)
	clientsMutex   sync.Mutex
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

var (
	matchQueue []*Client
	queueMutex sync.Mutex
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

func handleWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New WebSocket connection attempt")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()
	log.Println("WebSocket connection established")

	client := &Client{conn: conn}

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message:", err)
			break
		}

		log.Printf("Received raw message: %s", string(msgBytes))

		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Println("Invalid JSON:", err)
			conn.WriteJSON(map[string]string{"error": "Invalid JSON"})
			continue
		}

		log.Printf("Parsed message: %+v", msg)
		if msg.Name != "" {
			client.name = msg.Name
		}

		clientsMutex.Lock()

		switch msg.Action {
		case "create":
			queueMutex.Lock()
			for i, queuedClient := range matchQueue {
				if queuedClient == client {
					matchQueue = append(matchQueue[:i], matchQueue[i+1:]...)
					log.Printf("Removed %s from match queue (switching to create/join)\n", client.name)
					break
				}
			}
			queueMutex.Unlock()

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

		case "join":
			queueMutex.Lock()
			for i, queuedClient := range matchQueue {
				if queuedClient == client {
					matchQueue = append(matchQueue[:i], matchQueue[i+1:]...)
					log.Printf("Removed %s from match queue (switching to create/join)\n", client.name)
					break
				}
			}
			queueMutex.Unlock()

			if msg.Room == "" {
				conn.WriteJSON(map[string]string{"error": "Room code required to join"})
				clientsMutex.Unlock()
				continue
			}
			roomClients, exists := clientsPerRoom[msg.Room]
			if !exists {
				conn.WriteJSON(map[string]string{"error": "Room does not exist"})
				clientsMutex.Unlock()
				continue
			}
			if len(roomClients) >= 4 {
				conn.WriteJSON(map[string]string{"error": "Room is full"})
				clientsMutex.Unlock()
				continue
			}

			client.roomCode = msg.Room
			client.isHost = false
			clientsPerRoom[msg.Room] = append(clientsPerRoom[msg.Room], client)

			conn.WriteJSON(map[string]interface{}{
				"type": "joined",
			})

			log.Printf("%s joined room %s\n", client.name, msg.Room)
			broadcastPlayerCount(client.roomCode, len(clientsPerRoom[client.roomCode]))
		case "find_match":
			queueMutex.Lock()
			for i, queuedClient := range matchQueue {
				if queuedClient == client {
					matchQueue = append(matchQueue[:i], matchQueue[i+1:]...)
					log.Printf("Removed %s from match queue (switching to create/join)\n", client.name)
					break
				}
			}
			queueMutex.Unlock()

			client.isHost = false
			addToMatchQueue(client)
			err := conn.WriteJSON(map[string]string{
				"type": "searching",
			})
			if err != nil {
				log.Printf("Error sending searching message: %v\n", err)
			}
		case "start_game":
			if client.roomCode == "" {
				conn.WriteJSON(map[string]string{"error": "Not in any room"})
				clientsMutex.Unlock()
				continue
			}

			roomClients, exists := clientsPerRoom[client.roomCode]
			if !exists {
				conn.WriteJSON(map[string]string{"error": "Room does not exist"})
				clientsMutex.Unlock()
				continue
			}

			if !client.isHost {
				conn.WriteJSON(map[string]string{"error": "Only the host can start the game"})
				clientsMutex.Unlock()
				continue
			}

			if len(roomClients) < 2 {
				conn.WriteJSON(map[string]string{"error": "Need at least 2 players to start"})
				clientsMutex.Unlock()
				continue
			}

			startGame(client.roomCode)
			log.Printf("Host %s started the game in room %s\n", client.name, client.roomCode)

		case "cancel_find_match":
			queueMutex.Lock()
			for i, queuedClient := range matchQueue {
				if queuedClient == client {
					matchQueue = append(matchQueue[:i], matchQueue[i+1:]...)
					log.Printf("Removed %s from match queue (canceling find match)\n", client.name)
					break
				}
			}
			queueMutex.Unlock()
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

		case "leave_room":
			if client.roomCode != "" {
				removeClientFromRoom(client)
				client.roomCode = ""
				client.isHost = false
				conn.WriteJSON(map[string]string{"type": "left_room"})
				log.Printf("%s left the room\n", client.name)
			}

		default:
			conn.WriteJSON(map[string]string{"error": "Invalid action"})
		}

		clientsMutex.Unlock()
	}

	// Clean up when client disconnects
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	// Remove from match queue if they were searching
	queueMutex.Lock()
	for i, queuedClient := range matchQueue {
		if queuedClient == client {
			matchQueue = append(matchQueue[:i], matchQueue[i+1:]...)
			log.Printf("Removed %s from match queue\n", client.name)
			break
		}
	}
	queueMutex.Unlock()

	if client.roomCode != "" {
		removeClientFromRoom(client)
	}
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

	if len(matchQueue) >= limit {
		// Take the first 2 clients
		matched := matchQueue[:limit]
		matchQueue = matchQueue[limit:]

		roomCode := generateRoomCode()

		// Ensure room code is unique
		clientsMutex.Lock()
		for {
			if _, exists := clientsPerRoom[roomCode]; !exists {
				break
			}
			roomCode = generateRoomCode()
		}
		clientsPerRoom[roomCode] = matched
		clientsMutex.Unlock()

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

func removeClientFromRoom(client *Client) {
	roomClients, exists := clientsPerRoom[client.roomCode]
	if !exists {
		return
	}

	// If the host is leaving, notify all other players that the room is being destroyed
	if client.isHost && len(roomClients) > 1 {
		// Notify all other players (excluding the host)
		for _, c := range roomClients {
			if c != client {
				c.conn.WriteJSON(map[string]interface{}{
					"type":    "room_destroyed",
					"message": "The host has left the room. Room has been closed.",
				})
			}
		}
		log.Printf("Host %s left room %s - notifying %d other players\n", client.name, client.roomCode, len(roomClients)-1)
	}

	// Remove client from room
	for i, c := range roomClients {
		if c == client {
			clientsPerRoom[client.roomCode] = append(roomClients[:i], roomClients[i+1:]...)
			break
		}
	}

	// If room is empty or host left, delete it
	if len(clientsPerRoom[client.roomCode]) == 0 || client.isHost {
		delete(clientsPerRoom, client.roomCode)
		log.Printf("Room %s deleted\n", client.roomCode)
	} else {
		// If a non-host left, make someone else host and broadcast updated player count
		if len(clientsPerRoom[client.roomCode]) > 0 {
			clientsPerRoom[client.roomCode][0].isHost = true
			log.Printf("New host assigned in room %s\n", client.roomCode)
		}
		// Broadcast updated player count
		broadcastPlayerCount(client.roomCode, len(clientsPerRoom[client.roomCode]))
	}
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

func main() {
	// router := mux.NewRouter()

	// // Enable CORS for development
	// router.Use(func(next http.Handler) http.Handler {
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		w.Header().Set("Access-Control-Allow-Origin", "*")
	// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// 		if r.Method == "OPTIONS" {
	// 			w.WriteHeader(http.StatusOK)
	// 			return
	// 		}

	// 		next.ServeHTTP(w, r)
	// 	})
	// })

	// router.HandleFunc("/ws", handleWS)

	// api := router.PathPrefix("/api").Subrouter()
	// api.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "text/plain")
	// 	w.Write([]byte("Server running..."))
	// }).Methods("GET")

	http.HandleFunc("/ws", handleWS)

	log.Println("Server running on http://0.0.0.0:8080")
	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
