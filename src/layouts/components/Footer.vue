<script setup lang="ts">
import { useNow } from '@vueuse/core'
import { computed } from 'vue'

const now = useNow({ interval: 60_000 })
const year = computed(() => now.value.getFullYear())
</script>

<template>
  <NLayoutFooter bordered>
    <div class="mx-auto max-w-320 px-4 py-6 sm:px-6">
      <div class="divider-glow mb-5 w-24" />
      <NFlex justify="space-between" align="center" :wrap="true" class="gap-3">
        <NFlex vertical :size="2">
          <NText class="font-minecraft text-4 text-[var(--n-primary-color)]">
            Akari
          </NText>
          <NText depth="3" class="text-3">
            Minecraft 皮肤站 · 支持 Yggdrasil 与 authlib-injector
          </NText>
        </NFlex>
        <NText depth="3" class="text-3">
          Copyright &copy; {{ year }} Akari ·
          <NButton text tag="a" type="primary" target="_blank" rel="noopener noreferrer" href="https://github.com/lyimoexiao/akari">
            GitHub
          </NButton>
        </NText>
      </NFlex>
    </div>
  </NLayoutFooter>
</template>
