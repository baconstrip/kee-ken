gameHeaderTemplate = `<div class="row text-center justify-content-end" >
<roundcounter v-bind:board="board"></roundcounter>
<hostplayer v-bind:name="hostName"></hostplayer>
</div>
`

Vue.component('gameheader', {
    props: ['board', 'hostName'],
    template: gameHeaderTemplate,
});