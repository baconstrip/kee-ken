<script setup lang="ts">
import { computed } from 'vue';
import eventBus from '../eventbus';

const { host, started, board, joined } = defineProps<{
  host: boolean,
  started: boolean,
  board: any,
  joined: boolean,
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
  <div class="border-primary/40 border-4 bg-primary-content/50 rounded-xl p-4 flex flex-row gap-4 mx-30 mt-10" v-if="joined">
    <div>
      <button type="button" class="btn btn-error" @click="eventBus.emit('cancelGame');" v-if="host && started">
        Cancel Game!
      </button>
    </div>
    <div class="col" v-if="showNext && host">
      <div class="row">
        <button type="button" class="btn btn-success" @click="eventBus.emit('nextRound');">Next round</button>
      </div>
    </div>
    <button v-if="joined && host && !board" type="button" class="btn btn-success" @click="eventBus.emit('startGame')">
      Start Game
    </button>
  </div>
</template>