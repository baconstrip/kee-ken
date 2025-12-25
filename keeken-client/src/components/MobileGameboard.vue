<script setup lang="ts">
import { computed } from 'vue';
import eventBus from '../eventbus';

const { board } = defineProps<{
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

const rows = computed(() => {
    if (!board) {
        return;
    }
    var boardLength = board.Categories.length
    return [...Array(5).keys()].map(i => [...Array(boardLength).keys()].map(j => board.Categories[j].Questions[i]));
});
</script>

<template>
  <div class="row">
    <div class="col col-lg-12" id="gameboard">
      <template v-if="board">
        <table class="table text-center">
          <thead>
            <tr>
              <th scope="col" v-for="category in board.Categories" class="align-middle">
                {{ category.Name }}
              </th>
            </tr>
            <tr v-for="row in rows">
              <td v-for="question in row" v-bind:qid="question.ID" v-bind:played="question.Played">
                <span v-if="!question.Played">{{ question.Value }}</span>
              </td>
            </tr>
          </thead>
        </table>
      </template>
    </div>
  </div>
</template>
