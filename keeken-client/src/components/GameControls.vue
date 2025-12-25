<script setup lang="ts">
import { computed } from 'vue';
import eventBus from '../eventbus';

const { host, started, board } = defineProps<{
  host: boolean,
  started: boolean
  board: any,
}>();

const showNext = computed(() => {
  if (!board) {
    return '';
  }
  var unplayed = board.Categories.filter((c: any) => c.Questions.filter((q: any) => !q.Played).length > 0);
  console.log(unplayed)
  return unplayed.length == 0;
});
</script>

<template>
  <div>
    <button type="button" class="btn btn-danger" @click="eventBus.emit('cancelGame');" v-if="host && started">
      Cancel Game!
    </button>
  </div>
  <div class="col" v-if="showNext && host">
    <div class="row">
      <button type="button" class="btn btn-success" @click="eventBus.emit('nextRound');">Next round</button>
    </div>
  </div>
</template>