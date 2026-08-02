<script setup lang="ts">
import { computed } from 'vue'
import SkinViewer3D from '@/components/skin/SkinViewer3D.vue'
import { texturePreviewUrl } from '@/composables/useSkinUrl'

export interface TextureCardData {
  readonly name: string
  readonly type: 'steve' | 'alex' | 'cape'
  readonly hash: string
  readonly url?: string | null
  readonly likes?: number
}

const props = withDefaults(defineProps<{
  data: TextureCardData
  /** sm：选择器等小卡片；md：常规网格卡片 */
  size?: 'sm' | 'md'
  /** 是否自动旋转（卡片默认关闭，手动拖拽查看） */
  autoRotate?: boolean
  /** 是否允许鼠标/触摸拖拽旋转 */
  interactive?: boolean
  /** 停止操作多少秒后复位视角（默认 3 秒） */
  idleReset?: number
}>(), {
  size: 'md',
  autoRotate: false,
  interactive: true,
  idleReset: 3,
})

const isCape = computed(() => props.data.type === 'cape')
const previewUrl = computed(() => texturePreviewUrl(props.data))
const skinUrl = computed(() => (isCape.value ? null : previewUrl.value))
const capeUrl = computed(() => (isCape.value ? previewUrl.value : null))
const model = computed<'default' | 'slim'>(() => props.data.type === 'alex' ? 'slim' : 'default')

const typeLabel = computed(() => {
  switch (props.data.type) {
    case 'cape': return '披风'
    case 'alex': return 'Alex'
    default: return 'Steve'
  }
})
</script>

<template>
  <NCard size="small" hoverable class="texture-card h-full">
    <template #cover>
      <div
        class="relative w-full overflow-hidden"
        :class="size === 'sm' ? 'aspect-4/5' : 'aspect-3/4'"
      >
        <SkinViewer3D
          :skin-url="skinUrl"
          :cape-url="capeUrl"
          :model="model"
          :auto-rotate="autoRotate"
          :interactive="interactive"
          :idle-reset="idleReset"
          :fallback-url="previewUrl"
          :lazy="true"
        />
        <span
          class="absolute left-2 top-2 rounded-full px-2 py-0.5 text-2.5 font-medium backdrop-blur-md"
          :class="isCape
            ? 'bg-amber-500/85 text-white'
            : 'bg-[var(--n-primary-color)]/85 text-white'"
        >
          {{ typeLabel }}
        </span>
      </div>
    </template>
    <template #header>
      <div class="truncate text-sm font-medium" :title="data.name">
        {{ data.name }}
      </div>
    </template>
    <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
      <span class="flex items-center gap-1">
        <span class="i-ph-heart-fill text-3 text-[var(--n-error-color)]" />
        {{ data.likes ?? 0 }}
      </span>
      <span class="i-ph-cube text-3 opacity-60" title="3D 预览" />
    </div>
    <template v-if="$slots.actions" #footer>
      <slot name="actions" />
    </template>
  </NCard>
</template>

<style scoped>
.texture-card :deep(.n-card__content) {
  padding-top: 10px;
  padding-bottom: 4px;
}
</style>
