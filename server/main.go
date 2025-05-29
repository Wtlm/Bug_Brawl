package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

type Message struct {
	Action     string `json:"action"`
	Name       string `json:"name"`
	Room       string `json:"room,omitempty"`
	Answer     string `json:"answer,omitempty"`
	AnswerTime int64  `json:"answerTime,omitempty"`
}

type Option struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Question struct {
	ID      int   `json:"id"`
	Text    string   `json:"question"`
	Options []Option `json:"options"`
	Answer  string   `json:"answer"`
}

type Client struct {
	id     string
	conn   *websocket.Conn
	name   string
	room   *Room
	isHost bool
	Health int 
}

type PlayerAnswer struct {
	Client     *Client
	Answer     string
	AnswerTime int64
	Correct    bool
}

type Sabotage struct {
	Name     string
	Used     bool
	TargetID string
	UsedByID string
}

type Room struct {
	Players []*Client
	// Host          *Client
	RoomCode      string
	Question      *Question
	QuestionStart int64
	AnswerLog     []*PlayerAnswer
	RoundMutex    sync.Mutex
	// SabotageLog   map[string]string
	// AllSabotages  map[string]*Sabotage
	AvailableSabotages map[string][]*Sabotage
	PlayerEffects      map[string][]*Sabotage
	SabotageSelection  *SabotageSelection
}

type RoundResult struct {
	CorrectPlayers   []*PlayerAnswer
	IncorrectPlayers []*PlayerAnswer
	Winner           *PlayerAnswer
	Losers           []*PlayerAnswer
}

type SabotageSelection struct {
	WinnerID string
	Choices  map[string][]string
	Pending  map[string]bool
}

var (
	// clientsPerRoom = make(map[string][]*Client)
	clientsMutex   sync.Mutex
	matchQueue     []*Client
	queueMutex     sync.Mutex
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	rooms         = make(map[string]*Room)
	roomsMutex    sync.RWMutex
	questions     []Question
	questionMutex sync.Mutex
)

func main() {
	err := LoadQuestions("quiz.json")
	if err != nil {
		log.Fatal("Failed to load questions:", err)
	}
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
	err = http.ListenAndServe("0.0.0.0:8080", nil)
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

	client := &Client{conn: conn, id: generateClientID()}

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

		case "player_answer":
			handleAnswer(client, msg, conn)

		case "use_sabotage":
			handleUseSabotage(client, msg, conn)

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
			log.Printf("Removed %s (%s) from match queue\n", client.name, client.id)
			break
		}
	}
	queueMutex.Unlock()

	if client.room.RoomCode != "" {
		removeClientFromRoom(client)
	}
}
