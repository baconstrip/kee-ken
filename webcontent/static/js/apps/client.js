gameTemplate = `<div id="app" class="container">
<div class="alert alert-warning" role="alert" v-if="errorMessages">
  <span v-html="errorMessages"></span>
</div>
<hostplayer v-if="ws" v-bind:name="hostPlayerName">
</hostplayer>
<gameboard
    v-if="ws"
    v-bind:board="board"
    v-bind:host="host"
    v-on:selectQuestion="sendSelect"
    v-on:nextRound="sendNextRound">
</gameboard>
<component
    v-bind:is="questionComponent"
    v-bind:question="question"
    :key="question"
    v-bind:answer="answer"
    v-bind:duration="duration"
    v-bind:begin="beginCountdown"
    v-bind:answeringPlayer="answeringPlayer"
    v-bind:responsesClosed="responsesClosed"
    v-on:markCorrect='sendMarkAnswer(true)'
    v-on:markIncorrect='sendMarkAnswer(false)'
    v-on:finishReading="sendFinishReading"
    v-on:moveOn="sendMoveOn"
    v-on:buzz="sendBuzz">
</component>
<auth-window v-bind:host="host" v-on:auth-ready="join()" v-if="!joined">
</auth-window>
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
        board: '',
        host: false,
        hostPlayerName: "",
        question: '',
        answer: '',
        players: {},

        questionComponent: null,

        // Increment this number to indicate to the progress bar to start
        // counting.
        beginCountdown: 0,
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
        join: function() {
            var baseVue = this;
            $("#join-button").addClass("d-none");
            this.ws = new WebSocket('ws://' + window.location.host + '/player_game');
            this.ws.onopen = function(e) {
                baseVue.joined = true;
                errorMessages = "";
            };
            this.ws.onmessage = function(e) {
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
                    baseVue.beginCountdown++;
                }
                if (msg["Type"] == "PlayerAnswering") {
                    baseVue.duration = msg["Data"].Interval;
                    baseVue.answeringPlayer = msg["Data"].Name;
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
                    baseVue.beginCountdown = 0;
                    baseVue.question = '';
                    baseVue.answer = '';
                }
            };
            this.ws.onerror = function(e) {
                $("#join-button").removeClass("d-none");
                baseVue.errorMessages = "Couldn't establish connection to the server.";
                return;
            };
        },
        sendSelect: function(e) {
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
        sendFinishReading: function(e) {
            this.ws.send(
                JSON.stringify({
                    Type: "FinishReading",
                    Data: {},
                })
            );
        },
        sendBuzz: function(e) {
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
        sendMarkAnswer: function(e) {
            this.ws.send(
                JSON.stringify({
                    Type: "MarkAnswer",
                    Data: {
                        Correct: e == true,
                    },
                })
            );
        },
        sendMoveOn: function(e) {
            this.ws.send(
                JSON.stringify({
                    Type: "MoveOn",
                    Data: {},
                })
            );
        },
        sendNextRound: function(e) {
            this.ws.send(
                JSON.stringify({
                    Type: "NextRound",
                    Data: {},
                })
            );
        },
    },
    created: function() {
        if ($("#is-host").val()) {
            this.host = true;
        }
    }
});

