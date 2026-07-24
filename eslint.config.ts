import antfu from '@antfu/eslint-config'

export default antfu({
  vue: true,
  typescript: true,
  formatters: {
    css: true,
    html: true,
  },
  stylistic: {
    indent: 2,
    quotes: 'single',
  },
  ignores: [
    '**/dist/**',
    '**/dist-ssr/**',
    '**/coverage/**',
    '**/bin/**',
    '**/data/**',
  ],
})
