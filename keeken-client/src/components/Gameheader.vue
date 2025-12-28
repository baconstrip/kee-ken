<script setup lang="ts">
import { computed } from 'vue';
import eventBus from '../eventbus';

const {board, hostName, host } = defineProps<{
  board: any,
  hostName: string,
  host: boolean,
}>();

const roundName = computed(() => {
  if (board && board["Round"] !== undefined) {
    switch (board["Round"]) {
      case "1":
        return "Round 1 - 一番";
      case "2":
        return "Round 2 - 二番";
      case "3":
        return "Round 3 - 終わり";
      default:
        return "uh oh";
    }
  }
  return "Waiting for game!";
});

const hostdisplay = computed(() => {
  if (hostName) {
    return hostName;
  }
  return "No Host!";
});
</script>

<template>
  <div class="mx-auto w-full max-w-7xl">
    <div class="mb-8 flex items-center justify-between gap-6 rounded-lg border border-primary/30 bg-primary-content/70 px-6 py-4 backdrop-blur-sm md:px-8 md:py-5 mt-6">
      <div class="flex items-center gap-3">
        <div class="h-2 w-2 animate-pulse rounded-full bg-accent"></div>
        <p class="text-xs font-medium uppercase tracking-widest text-muted md:text-sm">
          {{ roundName }} 
        </p>
      </div>
      <div class="flex-1 text-center">
        <h1 class="text-2xl font-bold tracking-tight text-primary md:text-4xl">{{ hostdisplay }}</h1>
        <p class="mt-0.5 text-xs font-medium uppercase tracking-wider text-base-content md:text-sm">
          Host
        </p>
      </div>
      <div class="w-32 text-right text-xs text-md:w-40 md:text-sm text-base-content/50">
        <span class="font-mono">00:00</span>
      </div>
    </div>
  </div>
</template>