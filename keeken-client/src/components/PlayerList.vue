<script setup lang="ts">
import Player from './Player.vue';

defineProps<{
  players: any,
  host: boolean,
}>();
</script>

<template>
  <div class="row">
    <Player :name="n.Name" :money="n.Money" :connected="n.Connected" :selecting="n.Selecting" :host="host"
      v-for="(n, p) in players" :key="n.Name">
    </Player>
  </div>
</template>