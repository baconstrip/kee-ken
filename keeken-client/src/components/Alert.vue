<script setup lang="ts">
defineProps<{
    message: string
}>();
</script>

<template>
    <div class="mx-auto w-full content-center text-center max-w-3xl md:px-0 px-4">
        <div class="alert alert-warning max-w-3/4 block mx-auto">
            {{ message }}
        </div>
    </div>
</template>