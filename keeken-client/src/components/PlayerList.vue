<script setup lang="ts">
import Player from './Player.vue';

defineProps<{
  players: any,
  host: boolean,
}>();
</script>

<template>
  <div class="mx-auto mt-8 w-full max-w-7xl md:px-0 px-4">
    <div class="grid lg:grid-cols-4 md:grid-cols-3 grid-cols-1 gap-4">
      <Player :name="n.Name" :money="n.Money" :connected="n.Connected" :selecting="n.Selecting" :host="host"
        v-for="(n, p) in players" :key="n.Name">
      </Player>
    </div>
  </div>
</template>