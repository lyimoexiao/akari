<script setup lang="ts">
import { useNow } from '@vueuse/core'
import { computed } from 'vue'

const now = useNow({ interval: 60_000 })
const year = computed(() => now.value.getFullYear())
</script>

<template>
  <NLayoutFooter bordered>
    <NFlex justify="space-between" align="center" class="mx-auto max-w-320 px-4 py-3 text-3" :wrap="true">
      <NText depth="3">
        Copyright &copy; {{ year }} Akari.
      </NText>
      <NText depth="3">
        Powered by
        <NButton text tag="a" type="primary" target="_blank" rel="noopener noreferrer" href="https://github.com/lyimoexiao/akari">
          Akari
        </NButton>
      </NText>
    </NFlex>
  </NLayoutFooter>
</template>
