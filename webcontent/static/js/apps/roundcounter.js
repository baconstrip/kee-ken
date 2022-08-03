roundCounterTemplate = `
<div class="col-lg-4 col-m-6 col-sm-12 hostbox">
    <h2 v-if="board">Round {{ board["Round"] }}</h2>
    <h2 v-else><em>Waiting for game!</em></h2>
</div>
`

Vue.component('roundcounter', {
    props: ['board'],
    template: roundCounterTemplate,
});
