<script setup lang="ts">
import { computed } from 'vue';

const { name, money, connected, selecting } = defineProps<{
    name: string,
    money: Number,
    connected: boolean,
    selecting: boolean,
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

</script>

<template>
    <div class="col text-center" style="border: 2px black solid; background-color: #000">
        <div class="connection" v-bind:style="connectionStyle">
        </div>
        <h2>{{ name }}</h2>
        <h3>{{ money }}</h3>
        <div class="selecting mx-auto" v-bind:style="selectingStyle">
        </div>
    </div>
</template>
