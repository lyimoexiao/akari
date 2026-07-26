import type { GlobalTheme, GlobalThemeOverrides } from 'naive-ui'
import {
  darkTheme,
  dateEnUS,
  dateZhCN,
  enUS,
  lightTheme,
  zhCN,
} from 'naive-ui'
import { computed } from 'vue'
import { useAppStore } from '@/stores/app'

export function useNaiveTheme() {
  const appStore = useAppStore()

  const isDark = computed(() => appStore.isDark)

  const theme = computed<GlobalTheme | null>(() => {
    return isDark.value ? darkTheme : lightTheme
  })

  const themeOverrides = computed<GlobalThemeOverrides>(() => ({
    common: {
      baseColor: '#FFFFFF80',
    },
  }))

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

  return {
    theme,
    themeOverrides,
    locale,
    dateLocale,
  }
}
