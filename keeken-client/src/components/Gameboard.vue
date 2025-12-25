<script setup lang="ts">
import { computed } from 'vue';
import eventBus from '../eventbus';

const { board, host } = defineProps<{
    board: any,
    host: boolean
}>();

const select = (e: MouseEvent) => {
    if (!host) {
        return;
    }
    var elem = e.target as HTMLElement;
    if (elem.tagName == "SPAN") {
        elem = elem.parentElement as HTMLElement;
    }
    var id = elem.getAttribute("qid");
    if (elem.getAttribute("played")?.toLowerCase() == "true") {
      return;
    }
    eventBus.emit("selectQuestion", id as string);
};

const rows = computed(() => {
    if (!board) {
        return;
    }
    var boardLength = board.Categories.length
    return [...Array(5).keys()].map(i => [...Array(boardLength).keys()].map(j => board.Categories[j].Questions[i]));
});
</script>

<template>
  <div class="w-full">
    <div class="" id="gameboard" v-if="board">
      <div class="overflow-hidden rounded-3xl border-4 border-primary/40 bg-primary-content/70 shadow-2xl">
        <div class="grid grid-cols-6 border-b-4 border-primary/40 bg-primary-content/50">
          <div v-for="category in board.Categories" class="border-r-4 border-primary/40 p-4 last:border-r-0 md:p-6">
            <h2 class="text-center text-lg font-bold uppercase tracking-wide text-primary hyphens-auto text-pretty ">
              {{ category.Name }}
            </h2>
          </div>
        </div>
        <div v-for="row in rows" class="grid grid-cols-6 border-b4 borger-primary/40 last:border-b-0">
          <div v-for="question in row" @click="select" v-bind:qid="question.ID" v-bind:played="question.Played" 
          :disabled="question.Played"
          class="group relative border-r-4 border-primary/40 bg-primary-content/70 p-4 transition-all 
          hover:bg-primary/20 disabled:cursor-not-allowed disabled:bg-base-200/50 last:border-r-0 md:p-6">
            <div class="flex h-full items-center justify-content text-center" :qid="question.ID" :played="question.Played">
              <span v-if="!question.Played" 
              class="text-3xl mx-auto font-bold text-primary transition-transform group-hover:scale-110 md:text-5xl lg:text-6xl">
                {{ question.Value }}
              </span>
              <span v-else class="text-xl text-base-content/50 mx-auto">
                ---
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
    </div>
</template>
