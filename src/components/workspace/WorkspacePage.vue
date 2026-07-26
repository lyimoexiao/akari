<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <section class="mx-auto max-w-280 px-4 py-6 sm:px-6 sm:py-8">
    <NFlex justify="space-between" align="start" class="mb-6" :wrap="true">
      <div>
        <NH2 class="m-0! text-5! font-600!">
          {{ title }}
        </NH2>
        <NText v-if="description" depth="3" class="mt-1 block text-3.5">
          {{ description }}
        </NText>
      </div>
      <slot name="actions" />
    </NFlex>
    <slot />
  </section>
</template>
