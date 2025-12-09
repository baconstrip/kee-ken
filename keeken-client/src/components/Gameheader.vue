<script setup lang="ts">
import Hostplayer from './Hostplayer.vue';
import RoundCounter from './RoundCounter.vue';

import eventBus from '../eventbus';

defineProps<{
  board: any,
  hostName: string,
  host: boolean,
  started: boolean
}>();
</script>

<template>
  <div>
    <button type="button" class="btn btn-danger" @click="eventBus.emit('cancelGame');" v-if="host && started">
      Cancel Game!
    </button>
    <div class="row text-center justify-content-end">
      <RoundCounter :board="board"></RoundCounter>
      <Hostplayer :name="hostName"></Hostplayer>
    </div>
  </div>
</template>