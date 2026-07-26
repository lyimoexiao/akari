<script setup lang="ts">
import type { RequestLog } from '@/types/admin'
import { useAsyncState, useClipboard, useUrlSearchParams } from '@vueuse/core'
import { useMessage } from 'naive-ui'
import { computed, shallowRef } from 'vue'
import { ApiRequestError } from '@/api'
import { adminGetRequestLog } from '@/api/admin'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const message = useMessage()
const params = useUrlSearchParams<{ request_id?: string }>('history')
const requestId = shallowRef(typeof params.request_id === 'string' ? params.request_id : '')
const { copy, copied, isSupported } = useClipboard({ source: requestId })

async function loadRequestLog(): Promise<RequestLog | null> {
  const value = requestId.value.trim()
  if (!authStore.token || !value)
    return null
  params.request_id = value
  return adminGetRequestLog(authStore.token, value)
}

const {
  state: requestLog,
  isLoading,
  error,
  execute: lookup,
} = useAsyncState(loadRequestLog, null, { immediate: requestId.value.length > 0 })

const errorMessage = computed(() => error.value instanceof ApiRequestError
  ? error.value.message
  : error.value ? '查询请求日志失败' : null)

async function copyValue(value: string): Promise<void> {
  await copy(value)
  message.success('已复制')
}
</script>

<template>
  <WorkspacePage title="请求日志" description="通过响应中的 trace_id / request_id 查询单次请求详情">
    <NFlex class="mb-4" :wrap="true">
      <NInput v-model:value="requestId" class="w-full sm:max-w-110" clearable placeholder="输入 request_id" aria-label="请求 ID" @keyup.enter="lookup(0)" />
      <NButton type="primary" :loading="isLoading" :disabled="!requestId.trim()" @click="lookup(0)">
        查询
      </NButton>
    </NFlex>

    <NAlert v-if="errorMessage" type="error" class="mb-4">
      {{ errorMessage }}
    </NAlert>

    <NEmpty v-if="!requestLog && !isLoading && !errorMessage" description="输入请求 ID 开始查询" />

    <NCard v-else-if="requestLog" size="small" title="请求详情">
      <template #header-extra>
        <NButton v-if="isSupported" text type="primary" @click="copyValue(requestLog.request_id)">
          {{ copied ? '已复制' : '复制请求 ID' }}
        </NButton>
      </template>
      <NDescriptions label-placement="left" :column="1" size="small" bordered>
        <NDescriptionsItem label="请求 ID">
          <NText code class="break-all">
            {{ requestLog.request_id }}
          </NText>
        </NDescriptionsItem>
        <NDescriptionsItem label="模块">
          {{ requestLog.module }}
        </NDescriptionsItem>
        <NDescriptionsItem label="请求">
          <NText code>
            {{ requestLog.method }} {{ requestLog.path }}
          </NText>
        </NDescriptionsItem>
        <NDescriptionsItem label="状态 / 延迟">
          {{ requestLog.status }} / {{ requestLog.latency_ms }} ms
        </NDescriptionsItem>
        <NDescriptionsItem label="用户 / IP">
          {{ requestLog.user_id ?? '匿名' }} / {{ requestLog.ip }}
        </NDescriptionsItem>
        <NDescriptionsItem label="时间">
          {{ requestLog.created_at }}
        </NDescriptionsItem>
        <NDescriptionsItem label="User-Agent">
          <NText class="break-all">
            {{ requestLog.user_agent }}
          </NText>
        </NDescriptionsItem>
      </NDescriptions>

      <NCollapse class="mt-4">
        <NCollapseItem title="请求头" name="request-headers">
          <NCode :code="requestLog.request_headers" language="json" word-wrap />
        </NCollapseItem>
        <NCollapseItem title="请求体" name="request-body">
          <NCode :code="requestLog.request_body" language="json" word-wrap />
        </NCollapseItem>
        <NCollapseItem title="响应头" name="response-headers">
          <NCode :code="requestLog.response_headers" language="json" word-wrap />
        </NCollapseItem>
        <NCollapseItem title="响应体" name="response-body">
          <NCode :code="requestLog.response_body" language="json" word-wrap />
        </NCollapseItem>
      </NCollapse>
    </NCard>
  </WorkspacePage>
</template>
