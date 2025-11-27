<script setup lang="ts">
defineProps<{
  name: string  
}>()
</script>

<template>
    <div class="col-lg-4 col-m-6 col-sm-12 hostbox">
        <h2 v-if="name">Host: {{ name }}</h2>
        <h2 v-else><em>No Host!</em></h2>
    </div>
</template>