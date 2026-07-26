<script setup lang="ts">
import {
  NConfigProvider,
  NDialogProvider,
  NGlobalStyle,
  NLoadingBarProvider,
  NMessageProvider,
  NModalProvider,
  NNotificationProvider,
  useDialog,
  useLoadingBar,
  useMessage,
  useModal,
  useNotification,
} from 'naive-ui'

import { defineComponent, h } from 'vue'
import { useNaiveTheme } from '@/composables/useNaiveTheme'

const { theme, themeOverrides, locale, dateLocale } = useNaiveTheme()

function setupNaiveTools() {
  window.$loadingBar = useLoadingBar()
  window.$notification = useNotification()
  window.$message = useMessage()
  window.$dialog = useDialog()
  window.$modal = useModal()
}

const NaiveProviderContent = defineComponent({
  setup() {
    setupNaiveTools()
  },
  render() {
    return h('div', { className: 'naive-tools' })
  },
})
</script>

<template>
  <NConfigProvider :theme="theme" :theme-overrides="themeOverrides" :locale="locale" :date-locale="dateLocale">
    <!-- <NBackTop :visibility-height="1" class="z-1000" @update:show="isScrolled = $event" /> -->
    <NGlobalStyle />
    <NLoadingBarProvider>
      <NDialogProvider>
        <NNotificationProvider>
          <NMessageProvider>
            <NModalProvider>
              <slot />
              <NaiveProviderContent />
            </NModalProvider>
          </NMessageProvider>
        </NNotificationProvider>
      </NDialogProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>
