package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"slices"
	"time"
)

// LoadQuestions reads quiz.json and populates the questions slice
func LoadQuestions(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&questions); err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())
	log.Printf("Loaded %d questions from %s\n", len(questions), filename)
	return nil
}

// GetRandomQuestion returns a random question from the pool
func GetRandomQuestion() *Question {
	questionMutex.Lock()
	defer questionMutex.Unlock()

	if len(questions) == 0 {
		return nil
	}
	return &questions[rand.Intn(len(questions))]
}

func GenerateInitialSabotageList() []*Sabotage {
	var initialSabotage = []string{
		"BugSwarm",
		"BugEat",
		"BugLamp",
		"FakePopup",
		"CodeRain",
		"FLicker",
		"GlitchText",
		"BackwardText",
		"Blurry",
		"MouseDrift",
	}
	sabotages := make([]*Sabotage, len(initialSabotage))
	for i, name := range initialSabotage {
		sabotages[i] = &Sabotage{
			Name: name,
			Used: false,
		}
	}
	return sabotages
}

func (room *Room) EvaluateRoundResults() *RoundResult {
	room.RoundMutex.Lock()
	defer room.RoundMutex.Unlock()

	var correctAnswers []*PlayerAnswer
	var incorrectAnswers []*PlayerAnswer
	var winner *PlayerAnswer

	for _, pa := range room.AnswerLog {
		if pa.Correct {
			correctAnswers = append(correctAnswers, pa)
			if winner == nil || pa.AnswerTime < winner.AnswerTime {
				winner = pa
			}
		} else {
			incorrectAnswers = append(incorrectAnswers, pa)
		}
	}

	var losers []*PlayerAnswer
	losers = append(losers, incorrectAnswers...)
	var slowest *PlayerAnswer
	for _, pa := range correctAnswers {
		if pa.Client != nil && pa.Client.id != winner.Client.id {
			if slowest == nil || pa.AnswerTime > slowest.AnswerTime {
				slowest = pa
			}
		}
	}
	losers = append(losers, slowest)

	return &RoundResult{
		CorrectPlayers:   correctAnswers,
		IncorrectPlayers: incorrectAnswers,
		Winner:           winner,
		Losers:           losers,
	}
}
func (room *Room) CalculateHealth(winner *PlayerAnswer, losers []*PlayerAnswer) {
	if winner != nil && winner.Client != nil {
		client := winner.Client
		if client.Health < 5 {
			client.Health++
			log.Printf("Player %s gains 1 health. Total: %d", client.id, client.Health)
		}

		client.conn.WriteJSON(map[string]interface{}{
			"type":   "player_info",
			"id":     client.id,
			"name":   client.name,
			"health": client.Health,
		})
	}

	for _, client := range losers {
		if client.Client == nil {
			continue
		}
		c := client.Client
		c.Health--

		if c.Health <= 0 {
			log.Printf("Player %s is out of the game!", c.id)
			c.conn.WriteJSON(map[string]interface{}{
				"type":   "player_info",
				"id":     c.id,
				"name":   c.name,
				"health": c.Health,
			})
		} else {
			log.Printf("Player %s loses 1 health. Remaining: %d", c.id, c.Health)
			c.conn.WriteJSON(map[string]interface{}{
				"type":   "player_info",
				"id":     c.id,
				"name":   c.name,
				"health": c.Health,
			})
		}
	}

	room.CheckGameOver()
}

func (room *Room) AssignSabotagesToLosers(result *RoundResult) {
	if result == nil || len(result.Losers) == 0 {
		return
	}

	if result.Winner != nil && result.Winner.Client != nil {
		// Collect all available sabotages for each loser
		availablePerLoser := make(map[string]map[string]bool)
		for _, loser := range result.Losers {
			sabotageSet := make(map[string]bool)
			for _, s := range room.AvailableSabotages[loser.Client.id] {
				if !s.Used {
					sabotageSet[s.Name] = true
				}
			}
			availablePerLoser[loser.Client.id] = sabotageSet
		}

		// Calculate intersection of all sabotage names
		var intersection []string
		for name := range availablePerLoser[result.Losers[0].Client.id] {
			isCommon := true
			for _, sset := range availablePerLoser {
				if !sset[name] {
					isCommon = false
					break
				}
			}
			if isCommon {
				intersection = append(intersection, name)
			}
		}

		if len(intersection) == 0 {
			// No common sabotages available, assign randomly
			RandomSabotage(result.Losers, room)
			return
		}

		// Build a map where each loser maps to the same intersected sabotage list
		sabotageChoices := make(map[string][]string)
		for _, loser := range result.Losers {
			sabotageChoices[loser.Client.id] = intersection
		}

		// Store sabotage selection state
		room.SabotageSelection = &SabotageSelection{
			WinnerID: result.Winner.Client.id,
			Choices:  sabotageChoices,
			Pending:  make(map[string]bool),
		}
		for _, loser := range result.Losers {
			room.SabotageSelection.Pending[loser.Client.id] = true
		}

		//Notify winner
		err := result.Winner.Client.conn.WriteJSON(map[string]interface{}{
			"type":    "choose_sabotage",
			"choices": sabotageChoices,
		})
		if err != nil {
			log.Printf("error sending sabotage choices to winner: %v", err)
		}

	} else {
		// No winner: assign sabotages randomly
		RandomSabotage(result.Losers, room)
	}
}

func RandomSabotage(losers []*PlayerAnswer, room *Room) {
	for _, loser := range losers {
		var available []*Sabotage
		for _, s := range room.AvailableSabotages[loser.Client.id] {
			if !s.Used {
				available = append(available, s)
			}
		}
		if len(available) == 0 {
			continue
		}

		// Randomly choose one sabotage
		chosenIdx := rand.Intn(len(available))
		chosen := available[chosenIdx]

		// Move to PlayerEffects and remove from AvailableSabotages
		room.PlayerEffects[loser.Client.id] = append(room.PlayerEffects[loser.Client.id], chosen)
		room.AvailableSabotages[loser.Client.id] = slices.Delete(
			room.AvailableSabotages[loser.Client.id], chosenIdx,
			chosenIdx+1,
		)

		// Update metadata
		chosen.Used = true
		chosen.UsedByID = "system"
		chosen.TargetID = loser.Client.id

		loser.Client.conn.WriteJSON(map[string]interface{}{
			"type":     "sabotage_applied",
			"sabotage": chosen.Name,
			"usedBy":   "system",
			"targets":  loser.Client.name,
		})
	}
}

func (room *Room) StartQuestion() {

	question := GetRandomQuestion()
	if question == nil {
		log.Println("No question returned")
		return
	}

	// Reset previous round's answers, effects, etc.
	room.Question = question
	room.QuestionStart = time.Now().UnixMilli()
	room.AnswerLog = []*PlayerAnswer{}
	room.SabotageSelection = nil

	// Collect sabotage effects per player
	playerEffects := make(map[string][]string)
	for _, player := range room.Players {
		var effects []string
		for _, sabotage := range room.PlayerEffects[player.id] {
			if sabotage.Used && sabotage.TargetID == player.id {
				effects = append(effects, sabotage.Name)
			}
		}
		playerEffects[player.id] = effects
	}

	// Broadcast question to all players with their effects
	for _, player := range room.Players {
		err := player.conn.WriteJSON(map[string]interface{}{
			"type":    "question",
			"id":      question.ID,
			"text":    question.Text,
			"options": question.Options,
			"effect":  playerEffects[player.id],
		})
		if err != nil {
			log.Printf("error sending question to client %s: %v", player.id, err)
		}
	}
}

func (room *Room) CheckGameOver() {
	activePlayers := 0
	var lastPlayer *Client

	for _, client := range room.Players {
		if client.Health > 0 {
			activePlayers++
			lastPlayer = client
		}
	}

	if activePlayers <= 1 {
		log.Println("Game over!")

		// Broadcast winner (if any)
		if activePlayers == 1 && lastPlayer != nil {
			lastPlayer.conn.WriteJSON(map[string]interface{}{
				"type": "game_over",
				"note": "You win!",
			})
		}

		// Notify all clients
		for _, client := range room.Players {
			if client != lastPlayer {
				client.conn.WriteJSON(map[string]interface{}{
					"type": "game_over",
					"note": lastPlayer.name + " wins!",
				})
			}
		}
	}
}
