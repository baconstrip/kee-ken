gameTemplate = `<div id="app" class="container">
<div class="alert alert-warning" role="alert" v-if="errorMessages">
  <span v-html="errorMessages"></span>
</div>
<gameboard v-if="ws" v-bind:board="board">
</gameboard>
<questionWindow v-bind:question="question">
</questionWindow>
<auth-window v-bind:host="host" v-on:auth-ready="join()" v-if="!joined">
</auth-window>
<div class="row">
  <player v-bind:name="p.Name" v-bind:money="p.Money" v-for="p in players">
  </player>
</div>
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
        question: '',
        players: [],
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
                var msg = JSON.parse(e.data)
                console.log(msg)
                baseVue.serverContent += msg
                if (msg["Type"] == "BoardOverview") {
                    baseVue.board = msg["Data"];
                }
                if (msg["Type"] == "QuestionPrompt") {
                    baseVue.question = msg["Data"].Question;
                }
                if (msg["Type"] == "PlayerAdded") {
                    baseVue.players.push(msg["Data"])
                }

                baseVue.ws.send(
                    JSON.stringify({
                        Type: "ClientTestMessage",
                        Data: {
                            Example: "Data retreieved",
                            Number: 5,
                        },
                    })
                );
            };
            this.ws.onerror = function(e) {
                $("#join-button").removeClass("d-none");
                baseVue.errorMessages = "Couldn't establish connection to the server.";
                return;
            };
        }
    },
    created: function() {
        if ($("#is-host").val()) {
            this.host = true;
        }
    }
});

