<script setup lang="ts">
import type { ClosetItem } from '@/types/skinlib'
import { useDebounceFn } from '@vueuse/core'
import { useMessage } from 'naive-ui'
import { onMounted, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { listCloset, removeFromCloset, updateTexture } from '@/api/skinlib'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()

const items = shallowRef<readonly ClosetItem[]>([])
const total = shallowRef(0)
const loading = shallowRef(false)
const error = shallowRef<string | null>(null)
const page = shallowRef(1)
const pageSize = 20
const search = shallowRef('')
const category = shallowRef<'skin' | 'cape' | ''>('')
const renameTID = shallowRef<number | null>(null)
const renameName = shallowRef('')
const renameDialogOpen = shallowRef(false)

async function loadCloset() {
  if (!authStore.token)
    return
  loading.value = true
  error.value = null
  try {
    const result = await listCloset(authStore.token, {
      page: page.value,
      page_size: pageSize,
      type: category.value || undefined,
      search: search.value || undefined,
    })
    items.value = result.items as readonly ClosetItem[]
    total.value = result.total
  }
  catch (err) {
    error.value = err instanceof ApiRequestError ? err.message : '获取衣橱失败'
  }
  finally {
    loading.value = false
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  void loadCloset()
}, 300)

function onSearch() {
  page.value = 1
  void loadCloset()
}

function onPageChange(p: number) {
  page.value = p
  void loadCloset()
}

function onCategoryChange(cat: string) {
  category.value = cat as 'skin' | 'cape' | ''
  page.value = 1
  void loadCloset()
}

async function removeItem(tid: number) {
  if (!authStore.token)
    return
  try {
    await removeFromCloset(authStore.token, tid)
    message.success('已从衣橱移除')
    void loadCloset()
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '移除失败')
  }
}

function openRename(item: ClosetItem) {
  renameTID.value = item.texture_tid
  renameName.value = item.item_name
  renameDialogOpen.value = true
}

async function doRename() {
  if (!authStore.token || renameTID.value === null)
    return
  try {
    const { renameClosetItem } = await import('@/api/skinlib')
    await renameClosetItem(authStore.token, renameTID.value, { name: renameName.value })
    message.success('已重命名')
    renameDialogOpen.value = false
    void loadCloset()
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '重命名失败')
  }
}

function textureTypeLabel(type: string): string {
  switch (type) {
    case 'cape': return '披风'
    case 'alex': return 'Alex 皮肤'
    default: return 'Steve 皮肤'
  }
}

function isUploader(item: ClosetItem): boolean {
  return !!item.texture && item.texture.uploader === authStore.user?.id
}

function canToggleVisibility(item: ClosetItem): boolean {
  return isUploader(item) && !item.texture!.public
}

async function setToPublic(item: ClosetItem) {
  if (!authStore.token || !item.texture)
    return
  try {
    await updateTexture(authStore.token, item.texture_tid, {
      name: item.texture.name,
      public: true,
    })
    message.success('已设为公开')
    void loadCloset()
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '操作失败')
  }
}

const previewUrl = (hash: string, itemUrl?: string) => itemUrl || `${import.meta.env.VITE_API_BASE_URL ?? ''}/api/v1/raw/${hash}`

onMounted(() => {
  void loadCloset()
})
</script>

<template>
  <WorkspacePage title="我的衣橱" description="管理你收藏的皮肤和披风">
    <NButton class="mb-4" @click="router.push({ name: 'Skinlib' })">
      浏览皮肤库
    </NButton>

    <NAlert v-if="error" type="error" class="mb-4">
      {{ error }}
      <template #action>
        <NButton text type="error" @click="loadCloset()">
          重试
        </NButton>
      </template>
    </NAlert>

    <!-- Filters -->
    <NFlex class="mb-4" :wrap="true" align="center">
      <NInput
        v-model:value="search"
        clearable
        placeholder="搜索名称..."
        class="w-full sm:max-w-60"
        aria-label="搜索"
        @update:value="debouncedSearch"
        @keyup.enter="onSearch"
      />
      <NButton @click="onSearch">
        搜索
      </NButton>
      <NTabs v-model:value="category" type="segment" class="ml-auto" @update:value="onCategoryChange">
        <NTab key="" title="全部" />
        <NTab key="skin" title="皮肤" />
        <NTab key="cape" title="披风" />
      </NTabs>
    </NFlex>

    <!-- Closet grid -->
    <div v-if="loading" class="flex justify-center py-12">
      <NSpin />
    </div>

    <template v-else-if="items.length === 0">
      <NEmpty description="衣橱还是空的，去皮肤库逛逛吧" class="py-12">
        <template #action>
          <NButton type="primary" @click="router.push({ name: 'Skinlib' })">
            浏览皮肤库
          </NButton>
        </template>
      </NEmpty>
    </template>

    <template v-else>
      <div class="grid gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
        <NCard
          v-for="item in items"
          :key="item.texture_tid"
          size="small"
          hoverable
          class="closet-card"
        >
          <template #cover>
            <div class="flex aspect-1 items-center justify-center bg-gray-50 p-4 dark:bg-gray-800">
              <img
                v-if="item.texture"
                :src="previewUrl(item.texture.hash, item.texture.url)"
                :alt="item.item_name"
                class="max-h-32 object-contain image-render-pixel"
              >
            </div>
          </template>
          <template #header>
            <div class="truncate text-sm font-medium">
              {{ item.item_name }}
            </div>
          </template>
          <div class="flex items-center justify-between text-xs text-gray-500">
            <span v-if="item.texture">
              {{ textureTypeLabel(item.texture.type) }}
            </span>
            <span v-if="item.texture">{{ item.texture.likes }} ♥</span>
          </div>
          <template #footer>
            <NSpace justify="end">
              <NButton v-if="canToggleVisibility(item)" size="tiny" @click="setToPublic(item)">
                设为公开
              </NButton>
              <NButton v-if="isUploader(item)" size="tiny" @click="openRename(item)">
                重命名
              </NButton>
              <NButton size="tiny" type="error" @click="removeItem(item.texture_tid)">
                移除
              </NButton>
            </NSpace>
          </template>
        </NCard>
      </div>

      <NFlex class="mt-4" justify="end">
        <NPagination
          v-model:page="page"
          :page-size="pageSize"
          :item-count="total"
          @update:page="onPageChange"
        />
      </NFlex>
    </template>

    <!-- Rename dialog -->
    <NModal v-model:show="renameDialogOpen" title="重命名" preset="card" style="max-width: 400px">
      <NInput
        v-model:value="renameName"
        placeholder="输入新名称"
        @keyup.enter="doRename"
      />
      <template #footer>
        <NButton type="primary" @click="doRename">
          确定
        </NButton>
      </template>
    </NModal>
  </WorkspacePage>
</template>

<style scoped>
.image-render-pixel {
  image-rendering: pixelated;
}
</style>
