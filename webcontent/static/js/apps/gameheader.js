gameHeaderTemplate = `<div>
    <button type="button" class="btn btn-danger" @click="$emit('cancelGame');" v-if="host && started">Cancel Game!</button>
    <div class="row text-center justify-content-end" >
    <roundcounter v-bind:board="board"></roundcounter>
    <hostplayer v-bind:name="hostName"></hostplayer>
    </div>
</div>
`

Vue.component('gameheader', {
    props: ['board', 'hostName', 'host', 'started'],
    template: gameHeaderTemplate,
});