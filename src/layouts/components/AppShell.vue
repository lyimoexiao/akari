<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import Footer from './Footer.vue'
import Header from './Header.vue'

const appStore = useAppStore()
</script>

<template>
  <NLayout :native-scrollbar="false" position="absolute">
    <NBackTop :visibility-height="1" class="z-1000" @update:show="appStore.isScrolled = $event" />
    <Header />
    <main class="min-h-100vh">
      <slot />
    </main>
    <Footer />
  </NLayout>
</template>
