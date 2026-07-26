<script setup lang="ts">
import type { GlobalTheme } from 'naive-ui'
import {
  darkTheme,
  dateEnUS,
  dateZhCN,
  enUS,
  lightTheme,
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
  zhCN,
} from 'naive-ui'

import { computed, defineComponent, h } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

// 滚动状态：是否显示返回顶部按钮（即页面已滚动）
// const isScrolled = ref(false)

// 提供给子组件使用
// provide('isScrolled', isScrolled)

// 直接使用 store 中的 isDark computed
const isDark = computed(() => appStore.isDark)

const theme = computed<GlobalTheme | null>(() => {
  return isDark.value ? darkTheme : lightTheme
})

const locale = computed(() => {
  const langMap = {
    'zh-CN': zhCN,
    'en-US': enUS,
  }
  return langMap[appStore.lang] || zhCN
})

const dateLocale = computed(() => {
  const langMap = {
    'zh-CN': dateZhCN,
    'en-US': dateEnUS,
  }
  return langMap[appStore.lang] || dateZhCN
})

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
  <NConfigProvider :theme="theme" :locale="locale" :date-locale="dateLocale">
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
