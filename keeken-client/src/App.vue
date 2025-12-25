<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="h-screen w-screen">
    <div id="appcontainer" class="h-full w-full mx-auto md:px-4 xl:w-[1400px]">
      <RouterView />
    </div>
  </div>
</template>