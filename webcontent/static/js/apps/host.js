gameTemplate = `<div id="app" class="container">
<div class="alert alert-warning" role="alert" v-if="errorMessages">
  <span v-html="errorMessages"></span>
</div>
<gameboard v-if="ws" v-bind:board="board">
</gameboard>
<auth-window host="true" v-on:auth-ready="join()">
</auth-window>
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
        board: ''
    },
    template: gameTemplate,
    methods: {
        send: function () {
        },
        join: function() {
            

            var baseVue = this;
            this.ws = new WebSocket('ws://' + window.location.host + '/player_game');
            this.ws.onopen = function(e) {
                baseVue.ws.send(
                    JSON.stringify({
                        test: "message",
                    })
                );
            };
            this.ws.onmessage = function(e) {
                var msg = JSON.parse(e.data)
                console.log(msg)
                baseVue.serverContent += msg
                baseVue.board = msg
            };
            this.ws.onerror = function(e) {
                $("#join-button").removeClass("d-none");
                baseVue.errorMessages = "Couldn't establish connection to the server.";
                return;
            };
        }
    },
});

