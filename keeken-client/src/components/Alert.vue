<script setup lang="ts">
defineProps<{
    message: string
}>();
</script>

<template>
    <div>
        <div class="alert alert-warning">
            {{ message }}
        </div>
    </div>
</template>