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

/**
 * Akari 主题：Minecraft 翡翠绿主色 + 基岩灰暗色系。
 * 亮/暗两套 themeOverrides 共享相同圆角与字体。
 */
const baseOverrides: GlobalThemeOverrides = {
  common: {
    fontFamily: 'var(--font-basic)',
    borderRadius: '8px',
    borderRadiusSmall: '6px',
    fontSize: '14px',
    fontSizeSmall: '13px',
  },
  Card: {
    borderRadius: '12px',
  },
  Button: {
    borderRadiusMedium: '8px',
    borderRadiusSmall: '6px',
  },
  Input: {
    borderRadius: '8px',
  },
  Select: {
    borderRadius: '8px',
  },
  Menu: {
    borderRadius: '8px',
  },
  Tabs: {
    // segment 类型默认 padding 为 6px 0，胶囊紧贴文字，补左右留白
    tabPaddingSmallSegment: '4px 14px',
    tabPaddingMediumSegment: '6px 16px',
  },
  Table: {
    borderRadius: '10px',
  },
  DataTable: {
    borderRadius: '10px',
  },
}

const lightOverrides: GlobalThemeOverrides = {
  ...baseOverrides,
  common: {
    ...baseOverrides.common,
    primaryColor: '#3FB950',
    primaryColorHover: '#4CC763',
    primaryColorPressed: '#35A648',
    primaryColorSuppl: '#3FB950',
    successColor: '#2E9E5B',
    successColorHover: '#35AE65',
    successColorPressed: '#278C51',
    warningColor: '#D99A2B',
    warningColorHover: '#E3A83C',
    warningColorPressed: '#C08A24',
    errorColor: '#E5484D',
    errorColorHover: '#EE5A5F',
    errorColorPressed: '#D13A3F',
    bodyColor: '#F5F7F6',
    cardColor: '#FFFFFF',
    modalColor: '#FFFFFF',
    popoverColor: '#FFFFFF',
    tableColor: '#FFFFFF',
    tableHeaderColor: '#EEF1EF',
    inputColor: '#FBFCFB',
  },
  Card: {
    ...baseOverrides.Card,
    borderColor: 'rgba(24, 40, 32, 0.08)',
  },
  Layout: {
    headerColor: 'rgba(255, 255, 255, 0.88)',
    siderColor: '#F5F7F6',
    footerColor: 'transparent',
  },
}

const darkOverrides: GlobalThemeOverrides = {
  ...baseOverrides,
  common: {
    ...baseOverrides.common,
    primaryColor: '#4ADE80',
    primaryColorHover: '#63E695',
    primaryColorPressed: '#34C96B',
    primaryColorSuppl: '#4ADE80',
    successColor: '#4ADE80',
    successColorHover: '#63E695',
    successColorPressed: '#34C96B',
    warningColor: '#E3B341',
    warningColorHover: '#ECC257',
    warningColorPressed: '#C99B2F',
    errorColor: '#F0454A',
    errorColorHover: '#F45C61',
    errorColorPressed: '#D93338',
    bodyColor: '#0F1417',
    borderColor: 'rgba(255, 255, 255, 0.09)',
    cardColor: '#161D21',
    modalColor: '#161D21',
    popoverColor: '#1C252A',
    tableColor: '#161D21',
    tableHeaderColor: '#11181C',
    inputColor: '#0C1113',
    hoverColor: 'rgba(255, 255, 255, 0.06)',
  },
  Card: {
    ...baseOverrides.Card,
    borderColor: 'rgba(255, 255, 255, 0.08)',
  },
  Layout: {
    headerColor: 'rgba(15, 20, 23, 0.88)',
    siderColor: '#0F1417',
    footerColor: 'transparent',
  },
}

export function useNaiveTheme() {
  const appStore = useAppStore()

  const isDark = computed(() => appStore.isDark)

  const theme = computed<GlobalTheme | null>(() => {
    return isDark.value ? darkTheme : lightTheme
  })

  const themeOverrides = computed<GlobalThemeOverrides>(() => {
    return isDark.value ? darkOverrides : lightOverrides
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

  return {
    theme,
    themeOverrides,
    locale,
    dateLocale,
  }
}
