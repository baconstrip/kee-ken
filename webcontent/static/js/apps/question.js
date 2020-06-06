questionTemplate = `<div class="modal" id="question-modal" tabindex="-1" role="dialog" data-backdrop="static">
  <div class="modal-dialog modal-dialog-centered" role="document">
    <div class="modal-content" style="color: black">
      <div class="modal-header">
        <h5 class="modal-title">Question</h5>
      </div>
      <div class="modal-body" v-html="question">
      </div>
      <div class="modal-header" v-if="answer">
        <h5 class="modal-title">Answer</h5>
      </div>
      <div class="modal-body" v-if="answer">
        <p>{{ answer }}</p>
      </div>
      <div class="modal-footer" v-if="answer">
        <button type="button" class="btn btn-info" @click="$emit('finishReading')" v-if="!responsesClosed">Finish Reading</button>
        <button type="button" class="btn btn-success" @click="$emit('moveOn')" v-if="responsesClosed">Next Question</button>
      </div>
      <component 
        v-if="!responsesClosed"
        v-bind:duration="duration"
        @countdownStart="recordStart"
        v-bind:is="progressComponent">
      </component>
      <div class="modal-body" v-if="answeringPlayer && !responsesClosed">
        <h2>Player answering: {{ answeringPlayer }}</h2>
        <template v-if="answer">
          <button type="button" class="btn btn-success" @click="$emit('markCorrect')">Correct</button>
          <button type="button" class="btn btn-danger" @click="$emit('markIncorrect')">Incorrect</button>
        </template>
      </div>
    </div>
  </div>
</div>
`

Vue.component('questionWindow', {
    props: ['question', 'answer',  'duration', 'answeringPlayer', 'responsesClosed'],
    data: function() {
        return {
            lastStart: '',
            progressComponent: "progressbar",
        }
    },
    template: questionTemplate,
    methods: {
        recordStart: function() {
            this.lastStart = new Date();
        },
        finishTimer: function() {
            console.log("Button press took: ", ( new Date() - this.lastStart));
            var delta = new Date() - this.lastStart;
            if (delta > this.duration) {
                return;
            }

            this.$emit("buzz", delta);
        },
        beginCountdown: function(e) {
            this.lastStart = '';
        },
    },
    created: function() {
        EventBus.$on("spacePress", this.finishTimer);
        EventBus.$on("beginCountdown", this.beginCountdown);
    },
    mounted: function() {
        $("#question-modal").modal("show");
    },
});
