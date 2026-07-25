import { defineConfig, presetIcons, presetWind3 } from 'unocss'

export default defineConfig({
  presets: [
    presetWind3(),
    presetIcons(),
  ],
  shortcuts: [
    {
      'color-base': 'color-neutral-800 dark:color-neutral-200',
      'bg-base': 'bg-white dark:bg-#111',
      'bg-secondary': 'bg-#eee dark:bg-#222',
      'border-base': 'border-#8882',

      'bg-active': 'bg-#8881',
      'color-active': 'color-blue-600 dark:color-blue-300',
      'border-active': 'border-blue-600/25 dark:border-blue-400/25',

      'btn-action': 'inline-flex items-center gap-2 rounded border border-base px2 py1 op75 hover:op100 hover:bg-active disabled:pointer-events-none disabled:op30!',
      'btn-action-sm': 'btn-action text-sm',
      'btn-action-icon': 'inline-flex h-8 w-8 items-center justify-center rounded border border-base op75 hover:op100 hover:bg-active disabled:pointer-events-none disabled:op30!',

      'op-fade': 'op65 dark:op55',
      'op-mute': 'op30 dark:op25',

      'sider-width': 'w-56',
      'sider-collapsed-width': 'w-14',
      'h-header': 'h-12',

      'z-sider': 'z-50',
      'z-header': 'z-40',
      'z-overlay': 'z-60',
    },
  ],
  rules: [],
  safelist: [],
})
