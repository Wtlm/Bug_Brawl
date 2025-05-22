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
		client.name = msg.Name

		clientsMutex.Lock()

		switch msg.Action {
		case "create":
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

		default:
			conn.WriteJSON(map[string]string{"error": "Invalid action"})
		}

		clientsMutex.Unlock()
	}

	// Clean up when client disconnects
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	if client.roomCode != "" {
		removeClientFromRoom(client)
	}
}

func removeClientFromRoom(client *Client) {
	roomClients, exists := clientsPerRoom[client.roomCode]
	if !exists {
		return
	}

	// Remove client from room
	for i, c := range roomClients {
		if c == client {
			clientsPerRoom[client.roomCode] = append(roomClients[:i], roomClients[i+1:]...)
			break
		}
	}

	// If room is empty, delete it
	if len(clientsPerRoom[client.roomCode]) == 0 {
		delete(clientsPerRoom, client.roomCode)
		log.Printf("Room %s deleted (empty)\n", client.roomCode)
	} else {
		// If host left, make someone else host
		if client.isHost && len(clientsPerRoom[client.roomCode]) > 0 {
			clientsPerRoom[client.roomCode][0].isHost = true
			log.Printf("New host assigned in room %s\n", client.roomCode)
		}
		// Broadcast updated player count
		broadcastPlayerCount(client.roomCode, len(clientsPerRoom[client.roomCode]))
	}
}

func broadcastPlayerCount(room string, count int) {
	for _, c := range clientsPerRoom[room] {
		c.conn.WriteJSON(map[string]interface{}{
			"type":        "waiting",
			"playerCount": count,
		})
	}
}

func startGame(room string) {
	for _, c := range clientsPerRoom[room] {
		c.conn.WriteJSON(map[string]interface{}{
			"type": "start",
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
