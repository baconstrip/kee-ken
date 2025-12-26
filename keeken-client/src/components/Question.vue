<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, useTemplateRef } from 'vue';
import eventBus from '../eventbus';
import ProgressBar from './ProgressBar.vue';

const { question, answer, duration, answeringPlayer, responsesClosed, host, spectator, responsesOpen } = defineProps<{
  question: string,
  answer: string,
  duration: number,
  answeringPlayer: string | null,
  responsesClosed: boolean,
  host: boolean,
  spectator: boolean,
  responsesOpen: boolean,
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
    if (!responsesClosed && !responsesOpen) {
      eventBus.emit('finishReading');
    }

    if (responsesClosed) {
      eventBus.emit('moveOn');
    }
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
  <dialog class="modal modal-open" id="questionModal" tabindex="-1" role="dialog" data-backdrop="static"
    :class="{ 'penalty': showPenalty }" ref="questionModal">
    <div class="modal-box overflow-visible border-4 border-secondary/100 rounded-4xl md:max-w-3xl max-w-[90sw]"
      :class="{ 'lockedout': lockedout }" role="document">
      <div class="flex flex-col gap-4">
        <div class="py-5 text-2xl text-center md:leading-14 md:text-4xl md:py-10 md:mx-6" v-html="question"></div>
        <div class="content-right text-center -mt-4 mb-4" v-if="answer">
          <div class="badge badge-soft bg-secondary/20 badge-secondary badge-xl text-3xl py-6">
            <p>{{ answer }}</p>
          </div>
        </div>
        <ProgressBar v-if="!responsesClosed" :duration="duration">
        </ProgressBar>
        <div class="pt-5" v-if="host && ((!responsesClosed || !responsesOpen) && !answeringPlayer)">
          <button type="button" class="btn btn-info" @click="eventBus.emit('finishReading')"
            v-if="!responsesClosed && !responsesOpen">Finish Reading</button>
          <button type="button" class="btn btn-success" @click="eventBus.emit('moveOn')" v-if="responsesClosed">Next
            Question</button>
        </div>
        <div v-if="answeringPlayer && !responsesClosed">
          <div class="flex flex-row gap-4 pt-4" v-if="host">
            <div class="flex-grow"></div>
            <button type="button" class="btn btn-success btn-lg basis-3xs"
              @click="eventBus.emit('markCorrect')">Correct</button>
            <button type="button" class="btn btn-error btn-lg basis-3xs"
              @click="eventBus.emit('markIncorrect')">Incorrect</button>
            <div class="flex-grow"></div>
          </div>

        </div>
      </div>
      <div v-if="answeringPlayer && !responsesClosed"
        class="absolute top-100% left-0 rounded-full border-4 p-4 px-6 translate-x-4 translate-y-4 bg-primary-content border-accent md:max-w-xl max-w-[90sw]">
        <span class="text-xs uppercase tracking-wide">Buzzed</span>&emsp;<span class="text-lg">{{ answeringPlayer
          }}</span>
      </div>
    </div>

    <div class="modal-backdrop bg-black/40 backdrop-blur-sm transition-all"></div>
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
  color: black;
}

#gameboard {
  margin-bottom: 1rem;
}
</style>