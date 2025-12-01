<script setup lang="ts">
import Client from '@/components/Client.vue';

</script>

<template>
  <main>
    <Client :host="true" :spectator="false"></Client>
  </main>
</template>