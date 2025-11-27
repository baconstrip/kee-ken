<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, useTemplateRef } from 'vue';
import eventBus from '../eventbus';
import ProgressBar from './ProgressBar.vue';

const { question, answer, duration, answeringPlayer, responsesClosed, host } = defineProps<{
    question: string,
    answer: string,
    duration: number,
    answeringPlayer: string | null,
    responsesClosed: boolean,
    host: boolean,
}>();


const progressComponent = ref("progressbar");
const lastStart = ref<Date | null>(null);
const penalty = ref<Date | null>(null);
const lockedInterval = ref<number | null>(null);

const lockedout = ref(false);
const showPenalty = ref(false);

const questionModal = useTemplateRef<HTMLDivElement>('question-modal');

const recordStart = () => {
    lastStart.value = new Date();
};

const finishTimer = () => {
    if (!lastStart.value) {
        return;
    }
    const delta = new Date().getTime() - lastStart.value.getTime();
    if (delta > duration) {
        return;
    }

    eventBus.emit("buzz", delta);
};

const spacePress = () => {
  if (responsesClosed || answeringPlayer) {
    return;
  }

  if (!penalty.value && lastStart.value && !answeringPlayer) {
    finishTimer();
    return;
  } else if (new Date() > penalty.value! && lastStart.value && !answeringPlayer) {
    finishTimer();
    return;
  }

  const now = new Date();
  if (!penalty.value || now > penalty.value) {
    finishTimer();
    return;
  }

  console.log("Locked out due to penalty");

  // TODO this needs to be validated for behaviour
  showPenalty.value = true;
  window.setTimeout(() => {
    showPenalty.value = false;
  }, 1000);
  // $('#question-modal').addClass('penalty').delay(1000).queue(function () {
  //   $(this).removeClass('penalty');
  //   $(this).dequeue();
  // });

  if (!lockedInterval.value) {
    lockedout.value = true;
    lockedInterval.value = window.setInterval(function () {
      if (new Date() > penalty.value!) {
        lockedout.value = false;
        clearInterval(lockedInterval.value!);
        lockedInterval.value = null;
      }
    }, 80);
  }

  if (!penalty.value) {
    console.log("starting penalty");
    const dt = new Date();
    dt.setSeconds(dt.getSeconds() + 1);
    penalty.value = dt;
  } else {
    penalty.value.setSeconds(penalty.value.getSeconds() + 1.5);
  }
};

const hideQuestionListener = () => {
  // potentially need this
  // if (lockedInterval.value) {
  //   clearInterval(lockedInterval.value);
  //   lockedInterval.value = null;
  // }
  // lockedout.value = false;
  // penalty.value = null;
  
  (questionModal as any).modal("hide");
};

const countDownstartListener = () => {
  console.log("countdown started");
  recordStart();
};

eventBus.on("spacePress", spacePress);
eventBus.on("hideQuestion", hideQuestionListener);
eventBus.on("countdownStart", countDownstartListener);

onBeforeUnmount(() => {
  console.log("cleaning up");
  eventBus.off("spacePress", spacePress);
  eventBus.off("hideQuestion", hideQuestionListener);
  eventBus.off("countdownStart", countDownstartListener);
});

onMounted(() => {
  (questionModal as any).modal("show");
});

</script>

<template>
  <div class="modal" id="question-modal" tabindex="-1" role="dialog" data-backdrop="static" :class="{'penalty': showPenalty }" ref="question-modal">
    <div class="modal-dialog modal-dialog-centered" role="document" style="max-width: 95vw">
      <div class="modal-content" style="color: black" id="prompt-modal" :class="{ 'lockedout': lockedout }">
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
          <button type="button" class="btn btn-info" @click="eventBus.emit('finishReading')" v-if="!responsesClosed">Finish Reading</button>
          <button type="button" class="btn btn-success" @click="eventBus.emit('moveOn')" v-if="responsesClosed">Next Question</button>
        </div>
        <component 
          v-if="!responsesClosed"
          v-bind:duration="duration"
          v-bind:is="progressComponent">
        </component>
        <div class="modal-body" v-if="answeringPlayer && !responsesClosed">
          <h2>Player answering: {{ answeringPlayer }}</h2>
          <template v-if="answer">
            <button type="button" class="btn btn-success" @click="eventBus.emit('markCorrect')">Correct</button>
            <button type="button" class="btn btn-danger" @click="eventBus.emit('markIncorrect')">Incorrect</button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>