package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Room struct {
	Code    string
	Clients []*Client
}

type Message struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Room   string `json:"room,omitempty"`
}

type Client struct {
	conn   *websocket.Conn
	id     string
	name   string
	room   *Room
	isHost bool
}

var (
	rooms        = make(map[string]*Room)
	clientsMutex sync.Mutex
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

var (
	matchQueue []*Client
	queueMutex sync.Mutex
)

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

func handleWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New WebSocket connection attempt")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()
	log.Println("WebSocket connection established")

	client := &Client{
		conn: conn,
		id:   uuid.NewString()}

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
			handleCreateRoom(client, conn)

		case "join":
			handleJoinRoom(client, conn, msg)

		case "find_match":
			handleFindMatch(client, conn)

		case "start_game":
			handleStartGame(client, conn)

		case "cancel_find_match":
			handleCancelFindMatch(client, conn)

		case "leave_room":
			handleLeaveRoom(client, conn)

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

	if client.room.Code != "" {
		removeClientFromRoom(client)
	}
}
