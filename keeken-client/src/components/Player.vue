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

const connectionClasses = computed(() => {
    return {
        'bg-green-500': connected,
        'animate-pulse': connected,

        'bg-red-500': !connected,
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

const elmeentClasses = computed(() => {
    return {
        'border-accent': selecting,
        'shadow-lg': selecting,
        'shadow-accent/40': selecting,

        'border-base-content/20': !selecting,
    };
});
</script>

<template>
    <div class="rounded-lg border-2 px-6 py-5 transition-all bg-primary-content/70" :class="elmeentClasses">
        <div class="flex flex-col gap-3">
            <div class="flex items-center justify-between">
                <div class="h-2.5 w-4 rounded-full" :class="connectionClasses"></div>
                <div class="badge badge-accent text-sm tracking-wide" v-if="selecting">Picking</div>
            </div>
            <div class="flex items-baseline justify-between gap-3">
                <p class="text-lg font-bold text-primary">{{ name }}</p>
                <p class="text-2xl font-bold">{{ money }}</p>
            </div>
            <button v-if="host" class="btn btn-sm btn-warning mt-2" @click="openAdjustScore()">Adjust Score</button>
        </div>
    </div>
</template>
