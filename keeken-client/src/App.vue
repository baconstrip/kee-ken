<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="h-screen w-screen">
    <div id="appcontainer" class="h-full w-full mx-auto md:px-4 xl:w-[1200px] flex flex-col">
      <RouterView />
    </div>
  </div>
</template>