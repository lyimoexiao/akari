<script setup lang="ts">
import { darkTheme, useOsTheme } from 'naive-ui'

const osTheme = useOsTheme()
</script>

<template>
  <n-config-provider :theme="osTheme === 'dark' ? darkTheme : null" :theme-overrides="{ common: { borderRadius: '8px' } }">
    <n-global-style />
    <n-message-provider>
      <n-dialog-provider>
        <router-view />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style>
html,
body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  height: 100%;
}

#app {
  height: 100%;
}

*,
*::before,
*::after {
  box-sizing: border-box;
}
</style>
