<script setup lang="ts">
import { useTitle } from '@vueuse/core'
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Provider from './components/Provider.vue'

const route = useRoute()
useTitle(computed(() => `${route.meta.title ?? 'Akari'} · Akari`))
</script>

<template>
  <Provider>
    <RouterView />
  </Provider>
</template>
