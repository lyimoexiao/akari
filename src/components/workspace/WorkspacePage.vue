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
        <div class="mb-2 flex items-center gap-2">
          <span class="h-1.5 w-1.5 rounded-sm bg-[var(--n-primary-color)] shadow-[0_0_8px_var(--n-primary-color)]" />
          <NH2 class="m-0! text-5! font-700!">
            {{ title }}
          </NH2>
        </div>
        <NText v-if="description" depth="3" class="mt-1 block text-3.5">
          {{ description }}
        </NText>
      </div>
      <NFlex v-if="$slots.actions" :size="10" :wrap="true" justify="end">
        <slot name="actions" />
      </NFlex>
    </NFlex>
    <slot />
  </section>
</template>
