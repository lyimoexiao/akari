<script setup lang="ts">
import type { ClosetItem } from '@/types/skinlib'
import { shallowRef, watch } from 'vue'
import { ApiRequestError } from '@/api'
import { listCloset } from '@/api/skinlib'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  show: boolean
  title: string
  type: 'skin' | 'cape'
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'select': [tid: number, name: string]
}>()

const authStore = useAuthStore()
const items = shallowRef<readonly ClosetItem[]>([])
const loading = shallowRef(false)
const error = shallowRef<string | null>(null)
const page = shallowRef(1)
const total = shallowRef(0)
const pageSize = 20

async function loadItems() {
  if (!authStore.token)
    return
  loading.value = true
  error.value = null
  try {
    const result = await listCloset(authStore.token, {
      page: page.value,
      page_size: pageSize,
      type: props.type,
    })
    items.value = result.items as readonly ClosetItem[]
    total.value = result.total
  }
  catch (err) {
    error.value = err instanceof ApiRequestError ? err.message : '加载失败'
  }
  finally {
    loading.value = false
  }
}

function select(tid: number, name: string) {
  emit('select', tid, name)
  emit('update:show', false)
}

function close() {
  emit('update:show', false)
}

const previewUrl = (hash: string) => `${import.meta.env.VITE_API_BASE_URL ?? ''}/api/v1/raw/${hash}`

// Reload items each time the dialog opens
watch(() => props.show, (val) => {
  if (val) {
    page.value = 1
    void loadItems()
  }
})
</script>

<template>
  <NModal
    :show="show"
    :title="title || ' '"
    preset="card"
    style="max-width: 600px"
    @update:show="close"
  >
    <template v-if="loading">
      <div class="flex justify-center py-8">
        <NSpin />
      </div>
    </template>

    <template v-else-if="error">
      <NAlert type="error">
        {{ error }}
      </NAlert>
    </template>

    <template v-else-if="items.length === 0">
      <NEmpty :description="`衣橱中没有${type === 'cape' ? '披风' : '皮肤'}，先去皮肤库添加吧` || ' '" />
    </template>

    <template v-else>
      <div class="grid grid-cols-3 gap-3 sm:grid-cols-4">
        <NCard
          v-for="item in items"
          :key="item.texture_tid"
          size="small"
          hoverable
          class="cursor-pointer"
          @click="select(item.texture_tid, item.item_name)"
        >
          <div class="flex justify-center bg-gray-50 p-2 dark:bg-gray-800">
            <img
              v-if="item.texture"
              :src="previewUrl(item.texture.hash)"
              :alt="item.item_name"
              class="h-16 w-16 object-contain image-render-pixel"
            >
          </div>
          <div class="mt-1 truncate text-center text-xs">
            {{ item.item_name }}
          </div>
        </NCard>
      </div>

      <NFlex v-if="total > pageSize" class="mt-3" justify="center">
        <NPagination
          v-model:page="page"
          :page-size="pageSize"
          :item-count="total"
          @update:page="loadItems"
        />
      </NFlex>
    </template>
  </NModal>
</template>

<style scoped>
.image-render-pixel {
  image-rendering: pixelated;
}
</style>
