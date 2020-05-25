new Vue({
    el: '#app',

    data: {
        ws: null,
        serverResp: '',
        joined: false,
        name: '',
        serverContent: '',
        errorMessages: '',
    },

    methods: {
        send: function () {
        },
        join: function() {
            var baseVue = this;
            $("#join-button").addClass("d-none");
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
                baseVue.serverContent += msg['Test']
            };
            this.ws.onerror = function(e) {
                $("#join-button").removeClass("d-none");
                baseVue.errorMessages = "Couldn't establish connection to the server.";
                return;
            };
        }
    },
});

