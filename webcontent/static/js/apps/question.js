questionTemplate = `<div class="modal" id="question-modal" tabindex="-1" role="dialog" data-backdrop="static">
  <div class="modal-dialog modal-dialog-centered" role="document">
    <div class="modal-content">
      <div class="modal-header">
        <h5 class="modal-title">Question</h5>
      </div>
      <div class="modal-body">
        <p>{{ question }}</p>
      </div>
    </div>
  </div>
</div>
`

Vue.component('questionWindow', {
    props: ['question'],
    template: questionTemplate,
    watch: {
        question: function(q, _) {
            if (q) {
                $("#question-modal").modal("show");
            } else {
                $("#question-modal").modal("hide");
            }
        },
    }
});
