questionTemplate = `<div class="modal" id="question-modal" tabindex="-1" role="dialog" data-backdrop="static">
  <div class="modal-dialog modal-dialog-centered" role="document">
    <div class="modal-content" style="color: black" id="prompt-modal">
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
      <div class="modal-footer" v-if="host">
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
    props: ['question', 'answer',  'duration', 'answeringPlayer', 'responsesClosed', 'host'],
    data: function() {
        return {
            lastStart: '',
            progressComponent: "progressbar",
            penalty: null,
            lockedInterval: null,
        }
    },
    template: questionTemplate,
    methods: {
        recordStart: function() {
            this.lastStart = new Date();
        },
        spacePress: function() {
            var baseVue = this;
            if (this.responsesClosed || this.answeringPlayer){
                return;
            }
            if (!this.penalty && this.lastStart && !this.answeringPlayer) {
                this.finishTimer();
                return;
            } else if (new Date() > this.penalty && this.lastStart && !this.answeringPlayer) {
                this.finishTimer();
                return;
            }
            console.log("Locked out due to penalty");
            $('#question-modal').addClass('penalty').delay(1000).queue(function() {
                $(this).removeClass('penalty');
                $(this).dequeue();
            });

            var baseVue = this;
            if (!this.lockedInterval) {
                $('#prompt-modal').addClass('lockedout');
                baseVue.lockedInterval = setInterval(function() {
                    if (new Date() > baseVue.penalty) {
                        $('#prompt-modal').removeClass('lockedout');
                        clearInterval(baseVue.lockedInterval);
                        baseVue.lockedInterval = null;
                    }
                }, 80);
            }
           
            if (!this.penalty) {
                console.log("starting penalty");
                var dt = new Date()
                dt.setSeconds(dt.getSeconds() + 1);
                this.penalty = dt;
            } else {
                this.penalty.setSeconds(this.penalty.getSeconds() + 1.5);
            }
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
        EventBus.$on("spacePress", this.spacePress);
    },
    beforeDestroy: function() {
        console.log("cleaning up");
        EventBus.$off("spacePress", this.spacePress);
    },
    mounted: function() {
        $("#question-modal").modal("show");
    },
});
