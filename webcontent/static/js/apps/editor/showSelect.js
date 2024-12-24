showSelectTemplate = `<div class="row">
<div v-if="activeShow">
    <h3 class="show-header">Editing show {{ activeShow }}</h3>
</div>
<div v-if="!activeShow">
    <h3 class="show-header">Pick a show to begin editing</h3>
</div>

</div>
`

Vue.component('ShowSelect', {
    props: [],
    data: function() {
        return {
            "activeShow": null,
            "activeShowID": null,
        };
    },
    methods: {

    },
    template: showSelectTemplate,
});