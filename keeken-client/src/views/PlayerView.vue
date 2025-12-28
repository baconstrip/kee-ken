<script setup lang="ts">
import client from '@/components/Client.vue';
</script>

<template>
  <client :host="false" :spectator="false"></client>
</template>