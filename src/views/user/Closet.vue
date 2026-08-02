<script setup lang="ts">
import type { ClosetItem } from '@/types/skinlib'
import { useDebounceFn } from '@vueuse/core'
import { useMessage } from 'naive-ui'
import { computed, onMounted, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { updateTexture } from '@/api/skinlib'
import TextureCard from '@/components/skin/TextureCard.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useYggdrasilProfile } from '@/composables/useYggdrasilProfile'
import { useAuthStore } from '@/stores/auth'
import { useClosetStore } from '@/stores/closet'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const closetStore = useClosetStore()
const profile = useYggdrasilProfile()

const renameTID = shallowRef<number | null>(null)
const renameName = shallowRef('')
const renameDialogOpen = shallowRef(false)

const equipLoadingTID = shallowRef<number | null>(null)

const debouncedSearch = useDebounceFn(() => {
  closetStore.resetPage()
  void closetStore.fetchList()
}, 300)

function onSearch() {
  closetStore.resetPage()
  void closetStore.fetchList()
}

function onPageChange(p: number) {
  void closetStore.fetchList({ page: p })
}

function onCategoryChange(cat: string) {
  void closetStore.fetchList({ page: 1, category: cat as 'skin' | 'cape' | '' })
}

async function removeItem(tid: number) {
  try {
    await closetStore.remove(tid)
    message.success('已从衣橱移除')
    void closetStore.fetchList()
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
  if (renameTID.value === null)
    return
  try {
    await closetStore.rename(renameTID.value, renameName.value)
    message.success('已重命名')
    renameDialogOpen.value = false
    void closetStore.fetchList()
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '重命名失败')
  }
}

function isUploader(item: ClosetItem): boolean {
  return !!item.texture && item.texture.uploader === authStore.user?.id
}

function canToggleVisibility(item: ClosetItem): boolean {
  return isUploader(item) && !item.texture!.public
}

async function setToPublic(item: ClosetItem) {
  if (!item.texture)
    return
  try {
    await updateTexture(authStore.token!, item.texture_tid, {
      name: item.texture.name,
      public: true,
    })
    message.success('已设为公开')
    void closetStore.fetchList()
  }
  catch (err) {
    message.error(err instanceof ApiRequestError ? err.message : '操作失败')
  }
}

/** 装备：皮肤类型 → 设为皮肤；披风类型 → 设为披风 */
async function equip(item: ClosetItem) {
  if (!item.texture)
    return
  equipLoadingTID.value = item.texture_tid
  const isCape = item.texture.type === 'cape'
  const ok = isCape
    ? await profile.setCape(item.texture_tid)
    : await profile.setSkin(item.texture_tid)
  equipLoadingTID.value = null
  if (ok) {
    // 展示当前装备效果，供用户确认
    message.success(isCape ? '披风已装备' : '皮肤已装备')
  }
}

const itemsWithTexture = computed(() => closetStore.items.filter(item => item.texture))

onMounted(() => {
  void closetStore.fetchList()
})
</script>

<template>
  <WorkspacePage title="我的衣橱" description="管理你收藏的皮肤和披风，点击「设为皮肤 / 披风」即可装备">
    <template #actions>
      <NButton @click="router.push({ name: 'Skinlib' })">
        浏览皮肤库
      </NButton>
    </template>

    <NAlert v-if="closetStore.error" type="error" class="mb-4">
      {{ closetStore.error }}
      <template #action>
        <NButton text type="error" @click="closetStore.fetchList()">
          重试
        </NButton>
      </template>
    </NAlert>

    <!-- Filters -->
    <NFlex class="mb-4" :wrap="true" align="center">
      <NInput
        v-model:value="closetStore.search"
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
      <NTabs v-model:value="closetStore.category" type="segment" size="small" class="ml-auto w-fit! shrink-0" @update:value="onCategoryChange">
        <NTab key="all" name="" label="全部" />
        <NTab key="skin" name="skin" label="皮肤" />
        <NTab key="cape" name="cape" label="披风" />
      </NTabs>
    </NFlex>

    <!-- Closet grid -->
    <div v-if="closetStore.loading" class="flex justify-center py-12">
      <NSpin />
    </div>

    <template v-else-if="itemsWithTexture.length === 0">
      <NEmpty description="衣橱还是空的，去皮肤库逛逛吧" class="py-12">
        <template #action>
          <NButton type="primary" @click="router.push({ name: 'Skinlib' })">
            浏览皮肤库
          </NButton>
        </template>
      </NEmpty>
    </template>

    <template v-else>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
        <TextureCard
          v-for="item in itemsWithTexture"
          :key="item.texture_tid"
          :data="item.texture!"
        >
          <template #actions>
            <NSpace justify="between" align="center">
              <NButton
                size="tiny"
                type="primary"
                secondary
                :loading="equipLoadingTID === item.texture_tid"
                @click="equip(item)"
              >
                {{ item.texture!.type === 'cape' ? '设为披风' : '设为皮肤' }}
              </NButton>
              <NButton v-if="canToggleVisibility(item)" size="tiny" @click="setToPublic(item)">
                设公开
              </NButton>
              <NButton v-if="isUploader(item)" size="tiny" @click="openRename(item)">
                重命名
              </NButton>
              <NButton size="tiny" type="error" @click="removeItem(item.texture_tid)">
                移除
              </NButton>
            </NSpace>
          </template>
        </TextureCard>
      </div>

      <NFlex class="mt-4" justify="end">
        <NPagination
          :page="closetStore.page"
          :page-size="closetStore.pageSize"
          :item-count="closetStore.total"
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
