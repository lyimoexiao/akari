import type { ClosetItem } from '@/types/skinlib'
import { defineStore } from 'pinia'
import { shallowRef } from 'vue'
import { ApiRequestError } from '@/api'
import { addToCloset, listCloset, removeFromCloset, renameClosetItem } from '@/api/skinlib'
import { useAuthStore } from './auth'

export type ClosetCategory = 'skin' | 'cape' | ''

export const useClosetStore = defineStore('closet', () => {
  const authStore = useAuthStore()

  const items = shallowRef<readonly ClosetItem[]>([])
  const total = shallowRef(0)
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const page = shallowRef(1)
  const pageSize = shallowRef(20)
  const search = shallowRef('')
  const category = shallowRef<ClosetCategory>('')

  async function fetchList(overrides?: Partial<{
    page: number
    pageSize: number
    search: string
    category: ClosetCategory
  }>) {
    if (overrides) {
      if (overrides.page !== undefined)
        page.value = overrides.page
      if (overrides.pageSize !== undefined)
        pageSize.value = overrides.pageSize
      if (overrides.search !== undefined)
        search.value = overrides.search
      if (overrides.category !== undefined)
        category.value = overrides.category
    }
    if (!authStore.token)
      return
    loading.value = true
    error.value = null
    try {
      const result = await listCloset(authStore.token, {
        page: page.value,
        page_size: pageSize.value,
        type: category.value || undefined,
        search: search.value || undefined,
      })
      items.value = result.items
      total.value = result.total
    }
    catch (err) {
      error.value = err instanceof ApiRequestError ? err.message : '获取衣橱失败'
    }
    finally {
      loading.value = false
    }
  }

  /** 重置到第一页（搜索/分类变化后调用） */
  function resetPage() {
    page.value = 1
  }

  async function add(tid: number, name: string) {
    if (!authStore.token)
      return
    await addToCloset(authStore.token, { tid, name })
  }

  async function remove(tid: number) {
    if (!authStore.token)
      return
    await removeFromCloset(authStore.token, tid)
  }

  async function rename(tid: number, name: string) {
    if (!authStore.token)
      return
    await renameClosetItem(authStore.token, tid, { name })
  }

  return {
    items,
    total,
    loading,
    error,
    page,
    pageSize,
    search,
    category,
    fetchList,
    resetPage,
    add,
    remove,
    rename,
  }
})
