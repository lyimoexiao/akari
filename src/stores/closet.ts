import type { ClosetItem, ListClosetReq } from '@/types/skinlib'
import { defineStore } from 'pinia'
import { computed, shallowRef } from 'vue'
import { addToCloset, listCloset, removeFromCloset, renameClosetItem } from '@/api/skinlib'
import { useAuthStore } from './auth'

export const useClosetStore = defineStore('closet', () => {
  const authStore = useAuthStore()
  const items = shallowRef<readonly ClosetItem[]>([])
  const total = shallowRef(0)
  const loading = shallowRef(false)

  const textures = computed(() => items.value)

  async function fetchList(params?: ListClosetReq) {
    if (!authStore.token)
      return
    loading.value = true
    try {
      const result = await listCloset(authStore.token, params)
      items.value = result.items
      total.value = result.total
    }
    finally {
      loading.value = false
    }
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
    textures,
    fetchList,
    add,
    remove,
    rename,
  }
})
