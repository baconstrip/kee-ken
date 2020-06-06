hostPlayerTemplate = `<div class="row text-center justify-content-end" >
    <div class="col-4">
    <h2 v-if="name">Host: {{ name }}</h2>
    <h2 v-else><em>No Host!</em></h2>
    </div>
</div>
`

Vue.component('hostplayer', {
    props: ['name'],
    template: hostPlayerTemplate,
});
