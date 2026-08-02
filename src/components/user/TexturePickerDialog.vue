<script setup lang="ts">
import { watch } from 'vue'
import TextureCard from '@/components/skin/TextureCard.vue'
import { useClosetStore } from '@/stores/closet'

const props = defineProps<{
  show: boolean
  title: string
  type: 'skin' | 'cape'
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  'select': [tid: number, name: string]
}>()

const closetStore = useClosetStore()

function select(tid: number, name: string) {
  emit('select', tid, name)
  emit('update:show', false)
}

function close() {
  emit('update:show', false)
}

// 每次打开时重新加载对应类型的衣橱项
watch(() => props.show, (val) => {
  if (val)
    void closetStore.fetchList({ page: 1, category: props.type, search: '' })
})
</script>

<template>
  <NModal
    :show="show"
    :title="title || ' '"
    preset="card"
    style="max-width: 640px"
    @update:show="close"
  >
    <template v-if="closetStore.loading">
      <div class="flex justify-center py-8">
        <NSpin />
      </div>
    </template>

    <template v-else-if="closetStore.error">
      <NAlert type="error">
        {{ closetStore.error }}
      </NAlert>
    </template>

    <template v-else-if="closetStore.items.length === 0">
      <NEmpty :description="`衣橱中没有${type === 'cape' ? '披风' : '皮肤'}，先去皮肤库添加吧`" />
    </template>

    <template v-else>
      <div class="grid grid-cols-3 gap-3 sm:grid-cols-4">
        <TextureCard
          v-for="item in closetStore.items"
          v-show="item.texture"
          :key="item.texture_tid"
          :data="item.texture!"
          size="sm"
          :auto-rotate="false"
          class="cursor-pointer"
          @click="select(item.texture_tid, item.item_name)"
        />
      </div>

      <NFlex v-if="closetStore.total > closetStore.pageSize" class="mt-3" justify="center">
        <NPagination
          :page="closetStore.page"
          :page-size="closetStore.pageSize"
          :item-count="closetStore.total"
          @update:page="(p: number) => closetStore.fetchList({ page: p })"
        />
      </NFlex>
    </template>
  </NModal>
</template>
