// gameHandlers.js
import { getSocket } from "./socket.js";

export class GameHandler {
    constructor(socket, roomId, callbacks) {
        this.socket = socket;
        this.roomId = roomId;
        this.setPlayers = callbacks.setPlayers;
        this.setQuestion = callbacks.setQuestion;
        this.setShowPopup = callbacks.setShowPopup;
        this.setTimeLeft = callbacks.setTimeLeft;
        this.onTimeExpired = callbacks.onTimeExpired;
        // this.bugEatActiveRef = bugEatActiveRef;
        this.setChooseSabotage = callbacks.setChooseSabotage;
        this.setSabotageNoti = callbacks.setSabotageNoti;
        this.setPlayerEffects = callbacks.setPlayerEffects;
        this.setCodeRainActive = callbacks.setCodeRainActive;
        this.setRoundResultNoti = callbacks.setRoundResultNoti;
        this.onGameOver = callbacks.onGameOver;
        this.setWaitWinner = callbacks.setWaitWinner;
        // this.sendSabotageChoice = callbacks.sendSabotageChoice;
        // this.sendSabotageChoice = this.sendSabotageChoice.bind(this);
        this.timer = null;

        this.messageHandlers = {
            player_info: this.handlePlayerList.bind(this),
            player_update: this.handlePlayerUpdate.bind(this),
            question: this.handleNewQuestion.bind(this),
            // update_health: this.handleUpdateHealth.bind(this),
            choose_sabotage: this.handleChooseSabotage.bind(this),
            wait_winner: this.handleWaitForWinner.bind(this),
            sabotage_applied: this.handleSabotageNoti.bind(this),
            round_result: this.handleRoundResult.bind(this),
            game_over: this.handleGameOver.bind(this),
        };

        this.init();
    }

    init() {
        this.socket.onopen = () => {
            console.log("Game socket open. Joining room:", this.roomId);
            if (this.roomId) {
                this.socket.send(JSON.stringify({ type: "join_game", roomId: this.roomId }));
            }
        };
        if (this.socket.readyState === WebSocket.OPEN) {
            console.log("Socket already open. Joining room immediately:", this.roomId);
            if (this.roomId) {
                this.socket.send(JSON.stringify({ type: "join_game", roomId: this.roomId }));
            }
        }
        this.socket.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            console.log("Game socket message:", msg);

            const handler = this.messageHandlers[msg.type];
            if (handler) {
                handler(msg);
            } else {
                console.warn("Unhandled message type:", msg.type);
            }
        };
    }

    handlePlayerList(msg) {
        this.setPlayers(msg.players);
    }

    handlePlayerUpdate(msg) {
        this.setPlayers(msg.players);

    }
    handleNewQuestion(msg) {

        const question = {
            id: msg.id,
            question: msg.question,
            options: msg.options,
        };
    const effects = msg.effect || {};

        this.setQuestion(question);
        this.setPlayerEffects(effects);
        this.setShowPopup(true);

        // if (effects.includes("CodeRain")) {
        //     this.setCodeRainActive(true);
        // }

        // const questionTime = effects.includes("BugEat") ? 20 : 5;
        // this.startTimer(questionTime);
    }

    startTimer(seconds) {
        this.clearTimer();
        let timeLeft = seconds;
        this.setTimeLeft(timeLeft);

        this.timer = setInterval(() => {
            timeLeft -= 1;
            this.setTimeLeft(timeLeft);

            if (timeLeft <= 0) {
                this.clearTimer();
                this.onTimeExpired?.();
            }
        }, 1000);
    }

    clearTimer() {
        if (this.timer) {
            clearInterval(this.timer);
            this.timer = null;
        }
    }

    sendAnswer(answerId) {
        const answerTime = Date.now();

        const msg = {
            action: "player_answer",
            room: this.roomId,
            answer: answerId,
            answerTime: answerTime,
        };

        this.socket.send(JSON.stringify(msg));
    }

    handleChooseSabotage(msg) {
        const choices = msg.choices[Object.keys(msg.choices)[0]];

        this.setChooseSabotage(choices);
        setTimeout(() => this.setChooseSabotage?.(null), 3000);
    }
    sendSabotageChoice(sabotageName) {
        const msg = {
            action: "use_sabotage",
            name: sabotageName,
        };
        this.socket.send(JSON.stringify(msg));
    }
    handleWaitForWinner(msg) {
        this.setWaitWinner(msg.winner);
    }
    handleSabotageNoti(msg) {
        const { sabotage, usedBy, targets } = msg;
        let message = "";

        const targetNames = Array.isArray(targets) ? targets.map(t => t.name).join(', ') : targets;
        if (usedBy === "System") {
            message = "All players have been sabotaged";
        } else {
            message = `Winner used ${sabotage} on ${targetNames}`;
        }

        this.setSabotageNoti?.(message);
        setTimeout(() => this.setSabotageNoti?.(null), 3000);
    }

    handleRoundResult(msg) {
        const { winner, losers } = msg;

        let message = "";
        if (!winner) {
            message = "No correct answers this round.";
        } else { message = `Fatest Correct Answer: \n ${winner} `; }

        if (losers && losers.length > 0) {
            message += `\nIncorrect Answer (-1 heart):\n ${losers.join(', ')}`;
        }

        // this.setShowPopup(true);
        this.setRoundResultNoti?.(message);
        setTimeout(() => this.setRoundResultNoti?.(null), 3000);
    }

    handleGameOver(msg) {
        const { note } = msg;
        this.setRoundResultNoti?.(note);

        setTimeout(() => {
            this.setRoundResultNoti?.(null);
            this.setShowPopup(false);
            this.clearTimer();
            this.onGameOver?.();
        }, 300);
    }

}
