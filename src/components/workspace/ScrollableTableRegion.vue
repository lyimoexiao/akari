<script setup lang="ts">
import { useScroll } from '@vueuse/core'
import { useTemplateRef } from 'vue'

defineProps<{
  label: string
  minWidth: number
}>()

const scrollRegion = useTemplateRef<HTMLDivElement>('scrollRegion')
const { arrivedState } = useScroll(scrollRegion)
</script>

<template>
  <NText depth="3" class="mb-2 flex items-center gap-1 text-3 sm:hidden">
    <span class="i-ph-arrows-left-right-duotone" aria-hidden="true" />
    左右滑动查看更多
  </NText>
  <div class="relative">
    <div
      ref="scrollRegion"
      role="region"
      :aria-label="label"
      tabindex="0"
      class="overflow-x-auto overscroll-x-contain rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--n-color-target)]"
    >
      <div :style="{ minWidth: `${minWidth}px` }">
        <slot />
      </div>
    </div>
    <div
      v-if="!arrivedState.right"
      aria-hidden="true"
      class="pointer-events-none absolute inset-y-0 right-0 w-8 bg-gradient-to-l from-[var(--n-color)] to-transparent"
    />
  </div>
</template>
