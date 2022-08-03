alertTemplate = `<div>
<div class="alert alert-warning">
    {{ message }}
</div>
</div>
`

Vue.component('alertbox', {
    props: ['message'],
    template: alertTemplate,
});
