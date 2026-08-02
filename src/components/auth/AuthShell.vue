<script setup lang="ts">
import { useRouter } from 'vue-router'

defineProps<{
  title: string
  description?: string
}>()

const router = useRouter()
</script>

<template>
  <main class="relative flex min-h-100dvh items-center justify-center overflow-hidden px-4 py-10">
    <!-- 背景装饰：像素方块点缀 -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      <div class="absolute -left-16 top-10 opacity-10">
        <svg viewBox="0 0 32 32" class="h-40 w-40">
          <polygon points="16,1 31,9.5 16,18 1,9.5" fill="#8BD27C" />
          <polygon points="1,9.5 16,18 16,31 1,22.5" fill="#5CA85C" />
          <polygon points="31,9.5 16,18 16,31 31,22.5" fill="#3F8E4B" />
        </svg>
      </div>
      <div class="absolute -right-20 bottom-10 opacity-10">
        <svg viewBox="0 0 32 32" class="h-56 w-56">
          <polygon points="16,1 31,9.5 16,18 1,9.5" fill="#8BD27C" />
          <polygon points="1,9.5 16,18 16,31 1,22.5" fill="#5CA85C" />
          <polygon points="31,9.5 16,18 16,31 31,22.5" fill="#3F8E4B" />
        </svg>
      </div>
    </div>

    <NCard class="relative w-full max-w-100 shadow-[0_24px_80px_-24px_color-mix(in_srgb,var(--n-primary-color)_30%,transparent)]" :bordered="true" size="large">
      <template #header>
        <NFlex vertical align="center" :size="10" class="mb-2">
          <button class="group" aria-label="返回首页" @click="router.push({ name: 'Home' })">
            <svg viewBox="0 0 32 32" class="h-12 w-12 drop-shadow-md transition-transform duration-200 group-hover:-translate-y-1" aria-hidden="true">
              <polygon points="16,1 31,9.5 16,18 1,9.5" fill="#8BD27C" />
              <polygon points="1,9.5 16,18 16,31 1,22.5" fill="#5CA85C" />
              <polygon points="31,9.5 16,18 16,31 31,22.5" fill="#3F8E4B" />
            </svg>
          </button>
          <NFlex vertical align="center" :size="2">
            <NText class="font-minecraft text-4 text-[var(--n-primary-color)]">
              Akari
            </NText>
            <NH1 class="m-0! text-5! font-700!">
              {{ title }}
            </NH1>
            <NText v-if="description" depth="3" class="text-center text-3.5">
              {{ description }}
            </NText>
          </NFlex>
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
