<script setup lang="ts">
import eventBus from '@/eventbus';
import { computed } from 'vue';

const { name, money, connected, selecting } = defineProps<{
    name: string,
    money: Number,
    connected: boolean,
    selecting: boolean,
    host: boolean,
}>()

const connectionStyle = computed(() => {
    return {
        "background-color": connected ? "green" : "red",
    };
});

const selectingStyle = computed(() => {
    return {
        "background-color": selecting ? "orange" : "gray",
    };
});

const openAdjustScore = () => {
    eventBus.emit("openAdjustScore", name);
};
</script>

<template>
    <div class="col text-center" style="border: 2px black solid; background-color: #000">
        <div class="connection" :style="connectionStyle">
        </div>
        <h2>{{ name }}</h2>
        <h3>{{ money }}</h3>
        <div class="selecting mx-auto" :style="selectingStyle">
        </div>
        <button v-if="host" class="btn btn-sm btn-warning mt-2" @click="openAdjustScore()">Adjust Score</button>
    </div>
</template>
