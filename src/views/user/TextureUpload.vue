<script setup lang="ts">
import { useMessage } from 'naive-ui'
import { shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { uploadTexture } from '@/api/skinlib'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()

const name = shallowRef('')
const textureType = shallowRef<'steve' | 'alex' | 'cape'>('steve')
const file = shallowRef<File | null>(null)
const uploading = shallowRef(false)
const error = shallowRef<string | null>(null)
const previewUrlLocal = shallowRef<string | null>(null)

function onFileChange(data: { file: { file?: File } }) {
  const selectedFile = data.file?.file
  if (!selectedFile)
    return
  file.value = selectedFile
  if (!name.value) {
    name.value = selectedFile.name.replace(/\.[^.]+$/, '')
  }
  previewUrlLocal.value = URL.createObjectURL(selectedFile)
}

function discardPreview() {
  if (previewUrlLocal.value) {
    URL.revokeObjectURL(previewUrlLocal.value)
    previewUrlLocal.value = null
  }
}

async function submit() {
  if (!authStore.token) {
    error.value = '请先登录'
    return
  }
  if (!file.value) {
    error.value = '请选择文件'
    return
  }
  if (!name.value.trim()) {
    error.value = '请输入名称'
    return
  }

  uploading.value = true
  error.value = null
  try {
    await uploadTexture(authStore.token, file.value, name.value.trim(), textureType.value)
    discardPreview()
    message.success('上传成功')
    await router.push({ name: 'UserCloset' })
  }
  catch (err) {
    error.value = err instanceof ApiRequestError ? err.message : '上传失败'
  }
  finally {
    uploading.value = false
  }
}
</script>

<template>
  <WorkspacePage title="上传纹理" description="上传 Minecraft 皮肤或披风到你的衣橱">
    <NAlert v-if="error" type="error" class="mb-4">
      {{ error }}
    </NAlert>

    <NCard style="max-width: 500px">
      <NSpace vertical :size="16">
        <!-- File upload -->
        <div>
          <NText depth="3" class="mb-1 block text-sm">
            选择文件
          </NText>
          <NUpload
            accept="image/png"
            :max="1"
            :show-file-list="false"
            @change="onFileChange"
          >
            <NButton>
              选择 PNG 文件
            </NButton>
          </NUpload>
        </div>

        <!-- Preview -->
        <div v-if="previewUrlLocal" class="flex justify-center">
          <div class="bg-gray-100 p-4 dark:bg-gray-800">
            <img
              :src="previewUrlLocal"
              alt="预览"
              class="max-h-40 object-contain image-render-pixel"
            >
          </div>
        </div>

        <!-- Name -->
        <NInput
          v-model:value="name"
          placeholder="纹理名称"
          :disabled="uploading"
        />

        <!-- Type -->
        <div>
          <NText depth="3" class="mb-1 block text-sm">
            类型
          </NText>
          <NRadioGroup v-model:value="textureType">
            <NSpace>
              <NRadio value="steve">
                Steve 皮肤
              </NRadio>
              <NRadio value="alex">
                Alex 皮肤
              </NRadio>
              <NRadio value="cape">
                披风
              </NRadio>
            </NSpace>
          </NRadioGroup>
        </div>

        <!-- Submit -->
        <NButton
          type="primary"
          :loading="uploading"
          :disabled="!file || uploading"
          @click="submit"
        >
          上传
        </NButton>
      </NSpace>
    </NCard>
  </WorkspacePage>
</template>

<style scoped>
.image-render-pixel {
  image-rendering: pixelated;
}
</style>
