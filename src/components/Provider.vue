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
  useThemeVars,
} from 'naive-ui'

import { defineComponent, h, watch } from 'vue'
import { useNaiveTheme } from '@/composables/useNaiveTheme'

/**
 * naive-ui 的 --n-* CSS 变量只存在于使用它们的组件根元素上，
 * 自定义元素（Header、页面 div 等）无法直接引用。
 * 这里把常用主题变量提升到 :root，全局可用且跟随明暗主题切换。
 */
const GLOBAL_THEME_VARS = {
  '--n-primary-color': 'primaryColor',
  '--n-border-color': 'borderColor',
  '--n-text-color': 'textColor',
  '--n-text-color-3': 'textColor3',
  '--n-hover-color': 'hoverColor',
  '--n-error-color': 'errorColor',
  '--n-body-color': 'bodyColor',
  '--n-card-color': 'cardColor',
} as const

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
    const themeVars = useThemeVars()
    watch(themeVars, (vars) => {
      const style = document.documentElement.style
      for (const [cssVar, key] of Object.entries(GLOBAL_THEME_VARS)) {
        style.setProperty(cssVar, vars[key as keyof typeof vars])
      }
    }, { immediate: true })
  },
  render() {
    return h('div', { className: 'naive-tools' })
  },
})
</script>

<template>
  <NConfigProvider :theme="theme" :theme-overrides="themeOverrides" :locale="locale" :date-locale="dateLocale">
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
