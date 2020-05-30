alertTemplate = `<div class="alert alert-warning">
    {{ message }}
</div>
`

Vue.component('alertbox', {
    props: ['message'],
    template: alertTemplate,
});
