playerTemplate = `<div class="col text-center" style="border: 2px black solid; background-color: #000">
    <div class="connection" v-bind:style="connectionStyle">
    </div>
    <h2>{{ name }}</h2>
    <h3>{{ money }}</h3>
    <div class="selecting mx-auto" v-bind:style="selectingStyle">
    </div>
</div>
`

Vue.component('player', {
    props: ['name', 'money', 'connected', 'selecting'],
    template: playerTemplate,
    computed: {
        connectionStyle: function () {
            return {
                "background-color": this.connected ? "green" : "red",
            };
        },
        selectingStyle: function () {
            return {
                "background-color": this.selecting ? "orange" : "gray",
            };
        },
    },
});
