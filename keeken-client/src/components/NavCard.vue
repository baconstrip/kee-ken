<script setup lang="ts">
import { computed } from 'vue';


const { btntext, link, color } = defineProps<{
  btntext: string;
  link: string;
  color: string;
}>();

// prevent tailwind from purging dynamic classes
const unused = [
  'bg-primary', 'bg-secondary', 'bg-accent', 'bg-neutral', 'bg-info', 'bg-success', 'bg-warning', 'bg-error',
  'text-primary-content', 'text-secondary-content', 'text-accent-content', 'text-neutral-content', 'text-info-content',
  'text-success-content', 'text-warning-content', 'text-error-content',
  'btn-primary', 'btn-secondary', 'btn-accent', 'btn-neutral', 'btn-info', 'btn-success', 'btn-warning', 'btn-error',
  'btn-soft-primary', 'btn-soft-secondary', 'btn-soft-accent', 'btn-soft-neutral', 'btn-soft-info', 'btn-soft-success',
  'btn-soft-warning', 'btn-soft-error',
];

const buttonColor = computed(() => {
  return `btn btn-${color} btn-soft`;
});

const cardClasses = computed(() => {
  return `card bg-${color} text-${color}-content w-80 mx-3`;
});
</script>

<template>
  <div :class="cardClasses">
    <div class="card-body">
      <div class="card-title justify-center">
        <slot name="icon"></slot>
        <slot name="title"></slot>
      </div>
      <slot name="description"></slot>
      <div class="card-actions mt-2 justify-center">
        <RouterLink :to="link" type="button" class="btn mr-5" :class="buttonColor">
          {{ btntext }}
        </RouterLink>
      </div>
    </div>
  </div>
</template>