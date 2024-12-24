editorTemplate = `<div id="app" class="container">
<editor-auth-window v-on:auth-ready="join()" v-if="!joined"></editor-auth-window>
<ShowSelect v-if="joined"></ShowSelect>
</div>`

new Vue({
    el: "#editor-window",
    template: editorTemplate,
    data: {
        joined: false,
    },
    methods: {
        join: function() {
            var baseVue = this;
            $("#connect-button").addClass("d-none");
            this.ws = new WebSocket('ws://' + window.location.host + '/editor_ws');
            this.ws.onopen = function (e) {
                baseVue.joined = true;
                errorMessages = "";
            };
            this.ws.onmessage = function (e) {
                var msg = JSON.parse(e.data);
                console.log(msg);
                if (msg["Type"] == "BoardOverview") {
                    baseVue.board = msg["Data"];
                }
            };
            this.ws.onerror = function (e) {
                $("#join-button").removeClass("d-none");
                baseVue.errorMessages = "Couldn't establish connection to the server.";
                return;
            };
        },
        sendNewGame: function() {

        },
    },
    created: function() {
        //this.join();
    }
});