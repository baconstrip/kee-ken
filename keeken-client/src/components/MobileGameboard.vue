<script setup lang="ts">
const { board } = defineProps<{
    board: any,
}>();
</script>

<template>
  <div class="w-full" v-if="board && board.Round != '3'">
    <div class="" id="gameboard">
      <div class="overflow-hidden rounded-xl border-4 border-primary/40 bg-primary-content/70 shadow-2xl mx-4 max-w-full flex flex-col">
        <div class="grid grid-cols-5 gap-0 mt-6 first:mt-2" v-for="category in board.Categories">
          <div class="border-b-1 border-primary/40 text-center text-bold text-primary text-2xl col-span-5 pb-2 font-bold uppercase tracking-wide px-4">
           {{ category.Name }} 
          </div>
          <div v-for="question in category.Questions" class="group relative border-l-1 first:border-l-0 border-primary/40 bg-primary-content/70">
            <div class="h-full items-center justify-content text-center">
              <span v-if="!question.Played"
              class="text-2xl mx-auto font-bold text-primary">
                {{ question.Value }}
              </span>
              <span v-else class="text-xl text-base-content/50 mx-auto">
                ---
              </span>
            </div>
          </div>
          <div class="w-full border-b-4 border-primary/40 h-1 col-span-5"></div>
        </div>
      </div>
    </div>
  </div>
</template>