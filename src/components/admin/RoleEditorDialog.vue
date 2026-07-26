<script setup lang="ts">
import type { FormInst, FormRules } from 'naive-ui'
import type { RoleItem, SaveRoleReq } from '@/types/admin'
import { reactive, useTemplateRef, watch } from 'vue'

const props = defineProps<{
  role: RoleItem | null
  permissions: readonly string[]
  loading: boolean
}>()
const emit = defineEmits<{
  submit: [value: SaveRoleReq]
}>()
const show = defineModel<boolean>({ required: true })
const formRef = useTemplateRef<FormInst>('form')
const form = reactive({ name: '', description: '', permissions: [] as string[] })
const rules: FormRules = {
  name: [
    { required: true, message: '请输入角色名称', trigger: ['blur', 'input'] },
    { min: 2, max: 64, message: '角色名称应为 2–64 个字符', trigger: 'blur' },
  ],
}

watch(
  [show, () => props.role],
  ([isOpen, role]) => {
    if (!isOpen)
      return
    form.name = role?.name ?? ''
    form.description = role?.description ?? ''
    form.permissions = [...(role?.permissions ?? [])]
  },
  { immediate: true },
)

async function submit(): Promise<void> {
  try {
    await formRef.value?.validate()
  }
  catch (error) {
    if (Array.isArray(error))
      return
    throw error
  }
  emit('submit', {
    name: form.name.trim(),
    description: form.description.trim(),
    permissions: form.permissions,
  })
}
</script>

<template>
  <NModal v-model:show="show" preset="card" class="w-[calc(100vw-32px)]! max-w-140" :title="role ? '编辑角色' : '创建角色'">
    <NForm ref="form" :model="form" :rules="rules" @submit.prevent="submit">
      <NFormItem label="角色名称" path="name">
        <NInput v-model:value="form.name" placeholder="例如 auditor" :disabled="loading || role?.name === 'super_admin'" />
      </NFormItem>
      <NFormItem label="说明" path="description">
        <NInput v-model:value="form.description" type="textarea" placeholder="这个角色负责什么" :autosize="{ minRows: 2, maxRows: 4 }" />
      </NFormItem>
      <NFormItem label="权限" path="permissions">
        <NCheckboxGroup v-model:value="form.permissions">
          <div class="grid gap-2 sm:grid-cols-2">
            <NCheckbox v-for="permission in permissions" :key="permission" :value="permission" :label="permission" />
          </div>
        </NCheckboxGroup>
      </NFormItem>
      <NButton type="primary" attr-type="submit" :loading="loading">
        {{ role ? '保存修改' : '创建角色' }}
      </NButton>
    </NForm>
  </NModal>
</template>
