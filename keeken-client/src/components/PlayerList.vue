<script setup lang="ts">
import Player from './Player.vue';

defineProps<{
  players: any
}>();
</script>

<template>
  <div class="row">
    <Player v-bind:name="n.Name" v-bind:money="n.Money" v-bind:connected="n.Connected" v-bind:selecting="n.Selecting"
      v-for="(n, p) in players" v-bind:key="n.Name">
    </Player>
  </div>
</template>