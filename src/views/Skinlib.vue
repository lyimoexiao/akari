<script setup lang="ts">
import type { TextureItem } from '@/types/skinlib'
import { useAsyncState, useDebounceFn } from '@vueuse/core'
import { useDialog, useMessage } from 'naive-ui'
import { computed, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { addToCloset, deleteTexture, listTextures, updateTexture } from '@/api/skinlib'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const authStore = useAuthStore()

const page = shallowRef(1)
const pageSize = shallowRef(20)
const search = shallowRef('')
const typeFilter = shallowRef<'skin' | 'cape' | ''>('')
const order = shallowRef<'latest' | 'likes'>('latest')

async function loadTextures() {
  const token = authStore.token
  const result = await listTextures(token, {
    page: page.value,
    page_size: pageSize.value,
    type: typeFilter.value || undefined,
    search: search.value || undefined,
    order: order.value,
  })
  return result
}

const {
  state: result,
  isLoading,
  error,
  execute: refresh,
} = useAsyncState(loadTextures, { items: [] as readonly TextureItem[], total: 0, page: 1, page_size: 20, total_pages: 0 })

const errorMessage = computed(() => error.value instanceof ApiRequestError
  ? error.value.message
  : error.value ? '获取皮肤列表失败' : null)

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  void refresh(0)
}, 300)

function onSearch() {
  page.value = 1
  void refresh(0)
}

function onPageChange(p: number) {
  page.value = p
  void refresh(0)
}

function onTypeFilterChange(type: string) {
  typeFilter.value = type as 'skin' | 'cape' | ''
  page.value = 1
  void refresh(0)
}

function onOrderChange(o: string) {
  order.value = o as 'latest' | 'likes'
  void refresh(0)
}

const previewUrl = (texture: TextureItem) => texture.url || `${import.meta.env.VITE_API_BASE_URL ?? ''}/api/v1/raw/${texture.hash}`

function goUpload() {
  void router.push({ name: 'UserTextureUpload' })
}

async function addItemToCloset(texture: TextureItem) {
  if (!authStore.token)
    return
  try {
    await addToCloset(authStore.token, { tid: texture.tid, name: texture.name })
    message.success('已添加到衣橱')
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '添加到衣橱失败')
  }
}

function canDelete(texture: TextureItem): boolean {
  if (!authStore.token)
    return false
  return texture.uploader === authStore.user?.id || authStore.hasPermission('skinlib.manage')
}

function canEdit(texture: TextureItem): boolean {
  return canDelete(texture)
}

async function toggleVisibility(texture: TextureItem) {
  if (!authStore.token)
    return
  try {
    await updateTexture(authStore.token, texture.tid, {
      name: texture.name,
      public: !texture.public,
    })
    message.success(texture.public ? '已设为私有' : '已设为公开')
    void refresh(0)
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '修改失败')
  }
}

function confirmDelete(texture: TextureItem) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除「${texture.name}」吗？此操作不可撤销。`,
    positiveText: '删除',
    negativeText: '取消',
    positiveButtonProps: { type: 'error' },
    onPositiveClick: async () => {
      if (!authStore.token)
        return
      try {
        await deleteTexture(authStore.token, texture.tid)
        message.success('已删除')
        void refresh(0)
      }
      catch (err) {
        message.error(err instanceof ApiRequestError ? err.message : '删除失败')
      }
    },
  })
}
</script>

<template>
  <section class="mx-auto max-w-280 px-4 py-6 sm:px-6 sm:py-8">
    <NFlex justify="space-between" align="start" class="mb-6" :wrap="true">
      <div>
        <NH2 class="m-0! text-5! font-600!">
          皮肤库
        </NH2>
        <NText depth="3" class="mt-1 block text-3.5">
          浏览和收集由社区分享的 Minecraft 皮肤与披风
        </NText>
      </div>
      <NButton v-if="authStore.isAuthenticated" type="primary" @click="goUpload()">
        上传皮肤
      </NButton>
    </NFlex>

    <NAlert v-if="errorMessage" type="error" class="mb-4">
      {{ errorMessage }}
      <template #action>
        <NButton text type="error" @click="refresh(0)">
          重试
        </NButton>
      </template>
    </NAlert>

    <!-- Filters -->
    <NFlex class="mb-4" :wrap="true" align="center">
      <NInput
        v-model:value="search"
        clearable
        placeholder="搜索皮肤名称..."
        class="w-full sm:max-w-60"
        aria-label="搜索皮肤名称"
        @update:value="debouncedSearch"
        @keyup.enter="onSearch"
      />
      <NButton @click="onSearch">
        搜索
      </NButton>

      <NSpace class="ml-auto" :wrap="true">
        <NTabs v-model:value="typeFilter" type="segment" @update:value="onTypeFilterChange">
          <NTab key="" title="全部" />
          <NTab key="skin" title="皮肤" />
          <NTab key="cape" title="披风" />
        </NTabs>
        <NSelect
          v-model:value="order"
          :options="[
            { label: '最新上传', value: 'latest' },
            { label: '最多喜欢', value: 'likes' },
          ]"
          class="w-32"
          size="small"
          @update:value="onOrderChange"
        />
      </NSpace>
    </NFlex>

    <!-- Texture Grid -->
    <div v-if="isLoading" class="flex justify-center py-12">
      <NSpin />
    </div>

    <template v-else-if="result.items.length === 0">
      <NEmpty description="暂无皮肤" class="py-12" />
    </template>

    <template v-else>
      <div class="grid gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        <NCard
          v-for="texture in result.items"
          :key="texture.tid"
          size="small"
          hoverable
          class="texture-card"
        >
          <template #cover>
            <div class="flex aspect-1 items-center justify-center bg-gray-50 p-4 dark:bg-gray-800">
              <img
                :src="previewUrl(texture)"
                :alt="texture.name"
                class="max-h-32 object-contain image-render-pixel"
              >
            </div>
          </template>
          <template #header>
            <div class="truncate text-sm font-medium">
              {{ texture.name }}
            </div>
          </template>
          <div class="flex items-center justify-between text-xs text-gray-500">
            <span>{{ texture.type === 'cape' ? '披风' : texture.type === 'alex' ? 'Alex 皮肤' : 'Steve 皮肤' }}</span>
            <span>{{ texture.likes }} ♥</span>
          </div>
          <template #footer>
            <NSpace justify="between" align="center">
              <NButton v-if="canEdit(texture)" size="tiny" @click="toggleVisibility(texture)">
                {{ texture.public ? '设为私有' : '设为公开' }}
              </NButton>
              <NButton
                v-if="authStore.isAuthenticated"
                size="tiny"
                secondary
                @click="addItemToCloset(texture)"
              >
                添加到衣橱
              </NButton>
              <NButton
                v-if="canDelete(texture)"
                size="tiny"
                type="error"
                @click="confirmDelete(texture)"
              >
                删除
              </NButton>
            </NSpace>
          </template>
        </NCard>
      </div>

      <NFlex class="mt-4" justify="end">
        <NPagination
          v-model:page="page"
          :page-size="pageSize"
          :item-count="result.total"
          @update:page="onPageChange"
        />
      </NFlex>
    </template>
  </section>
</template>

<style scoped>
.image-render-pixel {
  image-rendering: pixelated;
}
</style>
