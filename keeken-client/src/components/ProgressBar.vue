<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue';
import eventBus from '../eventbus';

const { duration } = defineProps<{
    duration: any,
}>();

const start = ref(0);
const style = ref({
    width: "100%",
    "transition-duration": "0.1s",
});
const color = ref('danger');

const beginCountdownListener = (e: string) => {
    console.log("began countdown with duration: " + duration);
    if (e == "answer") {
        color.value = "success";
    }
    else if (e == "buzz") {
        color.value = "warning";
    }

    start.value = new Date().getTime();
    const timer = setInterval(shrink, 50);
    eventBus.emit("countdownStart");
    function shrink() {
        const delta = new Date().getTime() - start.value;
        if (delta > duration) {
            clearInterval(timer);
            style.value.width = "0%";
            return;
        }

        const frac = 1.0 - (delta / duration);
        style.value.width = (frac * 100) + "%";
    }
};  

const classes = ref(['progress-bar', 'bg-' + color.value]);

eventBus.on("beginCountdown", beginCountdownListener);
onBeforeUnmount(() => {
    eventBus.off("beginCountdown", beginCountdownListener);
});

</script>

<template>
<div class="progress" style="height: 20px;">
  <div v-bind:class="classes" role="progressbar" v-bind:style="style"></div>
</div>
</template>