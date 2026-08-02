import { defineConfig, presetIcons, presetWind4 } from 'unocss'

export default defineConfig({
  presets: [
    presetWind4(),
    presetIcons({
      extraProperties: {
        'display': 'inline-block',
        'vertical-align': 'middle',
      },
    }),
  ],
  shortcuts: [
    ['img-pixel', { 'image-rendering': 'pixelated' }],
  ],
  rules: [
    ['font-minecraft', { 'font-family': 'Minecraft' }],
  ],
  safelist: [],
})
