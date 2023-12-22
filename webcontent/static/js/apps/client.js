gameTemplate = `<div id="app" class="container">
<div class="alert alert-warning" role="alert" v-if="errorMessages">
  <span v-html="errorMessages"></span>
</div>
<gameheader v-if="ws"
    v-bind:hostName="hostPlayerName"
    v-bind:board="board"
    v-bind:host="host"
    v-bind:started="!!board"
    v-on:cancelGame="sendCancelGame">
</gameheader>
<gameboard
    v-if="board"
    v-bind:board="board"
    v-bind:host="host"
    v-on:selectQuestion="sendSelect"
    v-on:nextRound="sendNextRound">
</gameboard>
<alertbox message="Waiting for the host to start the game..." v-if="!board && joined">
</alertbox>
<button v-if="joined && host && !board" type="button" class="btn btn-success" @click="sendStartGame">Start Game</button>
<alertbox message="Connecting to the server..." v-if="ws == null && joined">
</alertbox>
<component
    v-bind:is="questionComponent"
    v-bind:question="question"
    v-if="questionComponent"
    :key="question"
    v-bind:answer="answer"
    v-bind:host="host"
    v-bind:duration="duration"
    v-bind:answeringPlayer="answeringPlayer"
    v-bind:responsesClosed="responsesClosed"
    v-on:markCorrect='sendMarkAnswer(true)'
    v-on:markIncorrect='sendMarkAnswer(false)'
    v-on:finishReading="sendFinishReading"
    v-on:moveOn="sendMoveOn"
    v-on:buzz="sendBuzz">
</component>
<owari
    v-if="owari"
    v-on:submitBid="sendBid"
    v-on:submitAnswer="sendOwariAnswer"
    v-bind:category="owari"
    v-bind:prompt="owariPrompt"
    v-bind:answers="owariAnswers"
    v-bind:bids="owariBids"
    v-bind:money="yourMoney"
    v-bind:host="host">
</owari>
<category-browser v-if="host && joined && !board">
</category-browser>
<auth-window v-bind:host="host" v-on:auth-ready="join()" v-if="!joined">
</auth-window>
<alertbox v-bind:message="gameErrors" v-if="gameErrors">
</alertbox>
<playerlist v-if="players" v-bind:players="players">
</playerlist>
</div>`

new Vue({
    el: '#game-app',

    data: {
        ws: null,
        serverResp: '',
        joined: false,
        name: '',
        serverContent: '',
        errorMessages: '',
        gameErrors: '',
        board: '',
        host: false,
        hostPlayerName: "",
        question: '',
        answer: '',
        players: {},

        owari: null,
        owariPrompt: null,
        owariAnswers: null,
        owariBids: null,
        yourMoney: -1,

        questionComponent: null,

        // Duration of the current countdown.
        duration: 0,

        // Player currently trying to answer the question
        answeringPlayer: 0,
        // Whether the current question is closed for responses.
        responsesClosed: false,
    },
    template: gameTemplate,
    methods: {
        send: function () {
        },
        join: function () {
            var baseVue = this;
            $("#join-button").addClass("d-none");
            this.ws = new WebSocket('ws://' + window.location.host + '/player_game');
            this.ws.onopen = function (e) {
                baseVue.joined = true;
                errorMessages = "";
            };
            this.ws.onmessage = function (e) {
                var msg = JSON.parse(e.data);
                console.log(msg);
                baseVue.serverContent += msg;
                if (msg["Type"] == "BoardOverview") {
                    baseVue.board = msg["Data"];
                }
                if (msg["Type"] == "QuestionPrompt") {
                    baseVue.questionComponent = "questionWindow";
                    baseVue.question = msg["Data"].Question;
                    baseVue.answer = msg["Data"].Answer || "";
                }
                if (msg["Type"] == "UpdatePlayers") {
                    baseVue.players = msg["Data"].Plys;
                }
                if (msg["Type"] == "HostAdd") {
                    baseVue.hostPlayerName = msg["Data"].Name;
                }
                if (msg["Type"] == "OpenResponses") {
                    baseVue.answeringPlayer = '';
                    baseVue.duration = msg["Data"].Interval;
                    EventBus.$emit("beginCountdown", "buzz");
                }
                if (msg["Type"] == "PlayerAnswering") {
                    baseVue.duration = msg["Data"].Interval;
                    baseVue.answeringPlayer = msg["Data"].Name;
                    EventBus.$emit("beginCountdown", "answer");
                }
                if (msg["Type"] == "CloseResponses") {
                    baseVue.responsesClosed = true;
                }
                if (msg["Type"] == "HideQuestion") {
                    $("#question-modal").modal("hide");
                    baseVue.questionComponent = null;
                    baseVue.answeringPlayer = '';
                    baseVue.duration = 0;
                    baseVue.responsesClosed = false;
                    baseVue.question = '';
                    baseVue.answer = '';
                }
                if (msg["Type"] == "ServerError") {
                    baseVue.gameErrors = msg["Data"].Error;
                    setTimeout(() => { baseVue.gameErrors = '' }, 10000);
                }
                if (msg["Type"] == "BeginOwari") {
                    baseVue.owari = msg["Data"].Category;
                    baseVue.yourMoney = msg["Data"].Money;
                }
                if (msg["Type"] == "ShowOwariPrompt") {
                    baseVue.owariPrompt = msg["Data"].Prompt;
                }
                if (msg["Type"] == "ShowOwariResults") {
                    baseVue.owariBids = msg["Data"].Bids;
                    baseVue.owariAnswers = msg["Data"].Answers;
                }
                if (msg["Type"] == "ClearBoard") {
                    baseVue.serverContent = '';
                    baseVue.errorMessages= '';
                    baseVue.gameErrors = '';
                    baseVue.board = '';
                    baseVue.owari = null;
                    baseVue.owariPrompt = null;
                    baseVue.owariAnswers = null;
                    baseVue.owariBids = null;
                    baseVue.question = '';
                }
            };
            this.ws.onerror = function (e) {
                $("#join-button").removeClass("d-none");
                baseVue.errorMessages = "Couldn't establish connection to the server.";
                return;
            };
        },
        sendSelect: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "SelectQuestion",
                    Data: {
                        ID: e.value,
                    },
                })
            );
            console.log(e);
        },
        sendFinishReading: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "FinishReading",
                    Data: {},
                })
            );
        },
        sendBuzz: function (e) {
            console.log(e);
            this.ws.send(
                JSON.stringify({
                    Type: "AttemptAnswer",
                    Data: {
                        ResponseTime: e,
                    },
                })
            );
        },
        sendMarkAnswer: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "MarkAnswer",
                    Data: {
                        Correct: e == true,
                    },
                })
            );
        },
        sendMoveOn: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "MoveOn",
                    Data: {},
                })
            );
        },
        sendNextRound: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "NextRound",
                    Data: {},
                })
            );
        },
        sendStartGame: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "StartGame",
                    Data: {},
                })
            );
        },
        sendBid: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "EnterBid",
                    Data: { Money: Number(e) },
                })
            );
        },
        sendOwariAnswer: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "FreeformAnswer",
                    Data: { Message: e },
                })
            );
        },
        sendCancelGame: function (e) {
            this.ws.send(
                JSON.stringify({
                    Type: "CancelGame",
                    Data: {  },
                })
            );
        },
    },
    created: function () {
        if ($("#is-host").val()) {
            this.host = true;
        }
    }
});

