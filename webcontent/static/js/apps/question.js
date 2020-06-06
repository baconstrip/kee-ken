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
        v-bind:begin="startCountdown"
        v-bind:duration="duration"
        v-bind:color="progressColor"
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
    props: ['question', 'answer', 'begin', 'duration', 'answeringPlayer', 'responsesClosed'],
    data: function() {
        return {
            lastStart: '',
            progressComponent: "progressbar",
            startCountdown: 0,
            progressColor: 'danger',
        }
    },
    template: questionTemplate,
    watch: {
        answeringPlayer: function(q, _) {
            this.progressComponent = "";
            this.progressComponent = "progressbar";
            this.lastStart = '';
            
            this.startCountdown++;
            this.progressColor = 'success';
        },
        begin: function(q, _) {
            this.progressColor = 'warning';
            this.startCountdown++; 
        },
    },
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
    },
    created: function() {
        EventBus.$on("spacePress", this.finishTimer);
    },
    mounted: function() {
        $("#question-modal").modal("show");
    },
});
