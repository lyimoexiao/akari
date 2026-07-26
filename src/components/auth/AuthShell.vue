<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<template>
  <main class="min-h-100dvh flex items-center justify-center bg-[var(--n-color-body)] px-4 py-8">
    <NCard class="w-full max-w-100" :bordered="true" size="large">
      <template #header>
        <NFlex vertical align="center" :size="4">
          <NH1 class="m-0! text-6! font-700!">
            {{ title }}
          </NH1>
          <NText v-if="description" depth="3" class="text-center text-3.5">
            {{ description }}
          </NText>
        </NFlex>
      </template>

      <slot />

      <template v-if="$slots.footer" #footer>
        <NFlex justify="center">
          <slot name="footer" />
        </NFlex>
      </template>
    </NCard>
  </main>
</template>
