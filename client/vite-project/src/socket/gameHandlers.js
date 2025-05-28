// gameHandlers.js

export class GameHandler {
    constructor(socket, roomId, callbacks) {
        this.socket = socket;
        this.roomId = roomId;
        this.setPlayers = callbacks.setPlayers;
        this.setQuestion = callbacks.setQuestion;
        this.setShowPopup = callbacks.setShowPopup;
        this.setTimeLeft = callbacks.setTimeLeft;
        this.onTimeExpired = callbacks.onTimeExpired;
        this.bugEatActiveRef = bugEatActiveRef;
        this.setChooseSabotage = callbacks.setChooseSabotage;
        this.setSabotageNoti = callbacks.setSabotageNoti;
        this.setPlayerEffects = callbacks.setPlayerEffects;
        this.setCdeRainActive = callbacks.setCodeRainActive;

        this.timer = null;

        this.messageHandlers = {
            player_list: this.handlePlayerList.bind(this),
            question: this.handleNewQuestion.bind(this),
            update_health: this.handleUpdateHealth.bind(this),
            choose_sabotage: this.handleChooseSabotage.bind(this),
            sabotage_applied: this.handleSabotageNoti.bind(this),
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

    handleNewQuestion(msg) {
        const question = {
            id: msg.id,
            question: msg.text,
            options: msg.options,
        };
        const effects = msg.effect || [];
        
        this.setQuestion(question);
        this.setPlayerEffects(effects);
        this.setShowPopup(true);

        if (effects.includes("CodeRain")) {
            this.setCdeRainActive(true);
        }

        const questionTime = effects.includes("BugEat") ? 20 : 30;
        this.startTimer(questionTime);
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
            type: "player_answer",
            room: this.roomId,
            answer: answerId,
            answerTime: answerTime,
        };

        this.socket.send(JSON.stringify(msg));
    }

    handleChooseSabotage(msg) {
        this.setChooseSabotage({
            choices: msg.choices[loserIds[0]],
            choose: this.sendSabotageChoice
        });
    }
    sendSabotageChoice(sabotageName) {
        const msg = {
            action: "use_sabotage",
            name: sabotageName,
        };
        this.socket.send(JSON.stringify(msg));
    }

    handleSabotageNoti(msg) {
        const { sabotage, usedBy, targets } = msg;
        const message = "";

        const targetNames = targets.map(t => t.name).join(', ');
        if (usedBy === "system") {
            message = "All players have been sabotaged";
        } else {
            message = `${usedBy} used ${sabotage} on ${targetNames}`;
        }

        this.setSabotageNoti?.(message);
        setTimeout(() => this.setSabotageNoti?.(null), 5000);
    }

    handleUpdateHealth(msg) {
        this.setPlayers(msg.players);
    }
}
