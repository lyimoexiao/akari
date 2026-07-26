import { usePreferredDark, useStorage } from '@vueuse/core'
import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'

export const THEME_MODES = ['auto', 'light', 'dark'] as const
export const LANGUAGES = ['zh-CN', 'en-US'] as const

export type ThemeMode = (typeof THEME_MODES)[number]
export type Language = (typeof LANGUAGES)[number]

export const useAppStore = defineStore('app', () => {
  const isScrolled = shallowRef(false)
  const themeMode = useStorage<ThemeMode>('akari_theme_mode', 'auto')
  const lang = useStorage<Language>('akari_language', 'zh-CN')
  const prefersDark = usePreferredDark()

  const isDark = computed(() => themeMode.value === 'auto'
    ? prefersDark.value
    : themeMode.value === 'dark')

  function updateThemeMode(mode?: ThemeMode): void {
    if (mode) {
      themeMode.value = mode
      return
    }

    const currentIndex = THEME_MODES.indexOf(themeMode.value)
    themeMode.value = THEME_MODES[(currentIndex + 1) % THEME_MODES.length] ?? 'auto'
  }

  return {
    isScrolled,
    themeMode,
    lang,
    isDark,
    updateThemeMode,
  }
})
