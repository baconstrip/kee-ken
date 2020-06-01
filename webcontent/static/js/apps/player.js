playerTemplate = `<div class="col text-center" style="border: 2px black solid; background-color: #000">
    <h2>{{ name }}</h2>
    <h3>{{ money }}</h3>
</div>
`

Vue.component('player', {
    props: ['name', 'money'],
    template: playerTemplate,
});
