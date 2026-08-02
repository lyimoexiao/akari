import type { YggdrasilStatusResp } from '@/types/auth'
import { useMessage } from 'naive-ui'
import { shallowRef } from 'vue'
import { ApiRequestError } from '@/api'
import {
  clearProfileCape,
  clearProfileSkin,
  getYggdrasilStatus,
  setProfileCape,
  setProfileSkin,
} from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

/**
 * Yggdrasil 档案操作：读取状态、装备/清除皮肤与披风。
 * 供账户中心、衣橱等页面复用，统一 loading 与错误提示。
 */
export function useYggdrasilProfile() {
  const authStore = useAuthStore()
  const message = useMessage()

  const status = shallowRef<YggdrasilStatusResp | null>(null)
  const loading = shallowRef(false)
  const mutating = shallowRef(false)
  const error = shallowRef<string | null>(null)

  async function refresh(): Promise<void> {
    if (!authStore.token)
      return
    loading.value = true
    error.value = null
    try {
      status.value = await getYggdrasilStatus(authStore.token)
    }
    catch (err) {
      error.value = err instanceof ApiRequestError ? err.message : '获取 Yggdrasil 状态失败'
    }
    finally {
      loading.value = false
    }
  }

  async function setSkin(tid: number): Promise<boolean> {
    if (!authStore.token)
      return false
    mutating.value = true
    try {
      await setProfileSkin(authStore.token, tid)
      message.success('皮肤已设置')
      await refresh()
      return true
    }
    catch (err) {
      message.error(err instanceof ApiRequestError ? err.message : '设置皮肤失败')
      return false
    }
    finally {
      mutating.value = false
    }
  }

  async function setCape(tid: number): Promise<boolean> {
    if (!authStore.token)
      return false
    mutating.value = true
    try {
      await setProfileCape(authStore.token, tid)
      message.success('披风已设置')
      await refresh()
      return true
    }
    catch (err) {
      message.error(err instanceof ApiRequestError ? err.message : '设置披风失败')
      return false
    }
    finally {
      mutating.value = false
    }
  }

  async function clearSkin(): Promise<void> {
    if (!authStore.token)
      return
    mutating.value = true
    try {
      await clearProfileSkin(authStore.token)
      message.success('皮肤已清除')
      await refresh()
    }
    catch (err) {
      message.error(err instanceof ApiRequestError ? err.message : '清除皮肤失败')
    }
    finally {
      mutating.value = false
    }
  }

  async function clearCape(): Promise<void> {
    if (!authStore.token)
      return
    mutating.value = true
    try {
      await clearProfileCape(authStore.token)
      message.success('披风已清除')
      await refresh()
    }
    catch (err) {
      message.error(err instanceof ApiRequestError ? err.message : '清除披风失败')
    }
    finally {
      mutating.value = false
    }
  }

  return {
    status,
    loading,
    mutating,
    error,
    refresh,
    setSkin,
    setCape,
    clearSkin,
    clearCape,
  }
}
