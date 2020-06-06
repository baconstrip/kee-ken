playerListTemplate = `<div class="row">
  <player v-bind:name="n.Name" v-bind:money="n.Money" v-for="(n, p) in players" v-bind:key="n.Name">
  </player>
</div>
`

Vue.component('playerlist', {
    props: ['players'],
    template: playerListTemplate,
});
