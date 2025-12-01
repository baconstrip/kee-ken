<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue';
import eventBus from '../eventbus';

// Prevent tailwind from purging these classes
const unused = ['bg-error', 'bg-warning', 'bg-success'];

const { duration } = defineProps<{
    duration: any,
}>();

const start = ref(0);
const style = ref({
    width: "100%",
    "transition-duration": "0.1s",
    height: "100%",
});
const color = ref('error');

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

const colorClass = () => {
    return 'bg-' + color.value;
};

// const colorClass = ref('bg-' + color.value);

eventBus.on("beginCountdown", beginCountdownListener);
onBeforeUnmount(() => {
    eventBus.off("beginCountdown", beginCountdownListener);
});

</script>

<template>
<div class="progress" style="height: 20px;">
  <div class="progress-bar" :class="colorClass()" role="progressbar" :style="style"></div>
</div>
</template>