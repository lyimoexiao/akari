<script setup lang="ts">
import type { TagProps } from 'naive-ui'
import { computed } from 'vue'
import { BUILT_IN_ROLES } from '@/types/auth'

const props = defineProps<{ role: string }>()

const label = computed(() => {
  switch (props.role) {
    case BUILT_IN_ROLES.SUPER_ADMIN:
      return '超级管理员'
    case BUILT_IN_ROLES.STAFF:
      return '职员'
    case BUILT_IN_ROLES.USER:
      return '用户'
    default:
      return props.role
  }
})

const type = computed<TagProps['type']>(() => {
  switch (props.role) {
    case BUILT_IN_ROLES.SUPER_ADMIN:
      return 'warning'
    case BUILT_IN_ROLES.STAFF:
      return 'info'
    default:
      return 'default'
  }
})
</script>

<template>
  <NTag :type="type" size="small" :bordered="false">
    {{ label }}
  </NTag>
</template>
