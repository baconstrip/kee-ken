<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, useTemplateRef } from 'vue';
import eventBus from '../eventbus';
import ProgressBar from './ProgressBar.vue';

const { question, answer, duration, answeringPlayer, responsesClosed, host, spectator } = defineProps<{
    question: string,
    answer: string,
    duration: number,
    answeringPlayer: string | null,
    responsesClosed: boolean,
    host: boolean,
    spectator: boolean,
}>();


const lastStart = ref<Date | null>(null);
const penalty = ref<Date | null>(null);
const lockedInterval = ref<number | null>(null);

const lockedout = ref(false);
const showPenalty = ref(false);

const questionModal = useTemplateRef<HTMLDivElement>('questionModal');

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
  if (host) {
    return;
  }

  if (spectator) {
    return;
  }

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
  console.log("Locked out due to penalty");

  showPenalty.value = true;
  window.setTimeout(() => {
    showPenalty.value = false;
  }, 1000);

  if (!lockedInterval.value) {
    lockedout.value = true;
    lockedInterval.value = window.setInterval(function () {
      if (new Date() > penalty.value!) {
        lockedout.value = false;
        clearInterval(lockedInterval.value!);
        lockedInterval.value = null;
        penalty.value = null;
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

const countDownstartListener = () => {
  console.log("countdown started");
  recordStart();
};

eventBus.on("spacePress", spacePress);
eventBus.on("countdownStart", countDownstartListener);

onBeforeUnmount(() => {
  console.log("cleaning up");
  eventBus.off("spacePress", spacePress);
  eventBus.off("countdownStart", countDownstartListener);
});

</script>

<template>
  <dialog class="modal modal-open" id="questionModal" tabindex="-1" role="dialog" data-backdrop="static" :class="{'penalty': showPenalty }" ref="questionModal">
    <div class="modal-box" role="document" style="max-width: 95vw">
      <div class="modal-content" id="prompt-modal" :class="{ 'lockedout': lockedout }">
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
        <ProgressBar
          v-if="!responsesClosed"
          :duration="duration">
        </ProgressBar>
        <div class="modal-body" v-if="answeringPlayer && !responsesClosed">
          <h2>Player answering: {{ answeringPlayer }}</h2>
          <template v-if="answer">
            <button type="button" class="btn btn-success" @click="eventBus.emit('markCorrect')">Correct</button>
            <button type="button" class="btn btn-danger" @click="eventBus.emit('markIncorrect')">Incorrect</button>
          </template>
        </div>
      </div>
    </div>
  </dialog>
</template>

<style lang="postcss" scoped>
@keyframes flash {
  from {
    background-color: red
  }

  to {}
}

.penalty {
  animation-name: flash;
  animation-duration: 1s;
}

.lockedout {
  background-color: orange;
}

#gameboard {
  margin-bottom: 1rem;
}
</style>