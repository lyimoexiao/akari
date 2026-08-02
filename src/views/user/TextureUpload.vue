<script setup lang="ts">
import type { UploadFileInfo } from 'naive-ui'
import { useMessage } from 'naive-ui'
import { computed, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ApiRequestError } from '@/api'
import { uploadTexture } from '@/api/skinlib'
import SkinViewer3D from '@/components/skin/SkinViewer3D.vue'
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
const sizeError = shallowRef<string | null>(null)
const objectUrl = shallowRef<string | null>(null)

function readImageSize(f: File): Promise<{ width: number, height: number }> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(f)
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(url)
      resolve({ width: img.naturalWidth, height: img.naturalHeight })
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('无法读取图片尺寸'))
    }
    img.src = url
  })
}

async function validateSize(f: File) {
  sizeError.value = null
  let size: { width: number, height: number }
  try {
    size = await readImageSize(f)
  }
  catch {
    sizeError.value = '无法读取图片，请确认是有效的 PNG 文件'
    return
  }
  const { width, height } = size
  if (textureType.value === 'cape') {
    const ok = (width % 64 === 0 && height % 32 === 0) || (width % 22 === 0 && height % 17 === 0)
    if (!ok)
      sizeError.value = `披风尺寸需为 64×32 或 22×17 的倍数，当前 ${width}×${height}`
  }
  else {
    const ok = width % 64 === 0 && height % 32 === 0
    if (!ok)
      sizeError.value = `皮肤尺寸需为 64×32 的倍数，当前 ${width}×${height}`
  }
}

function onFileChange(data: { file: UploadFileInfo }) {
  const selectedFile = data.file.file
  if (!selectedFile)
    return
  file.value = selectedFile
  if (objectUrl.value) {
    URL.revokeObjectURL(objectUrl.value)
    objectUrl.value = null
  }
  objectUrl.value = URL.createObjectURL(selectedFile)
  if (!name.value) {
    name.value = selectedFile.name.replace(/\.[^.]+$/, '')
  }
  void validateSize(selectedFile)
}

function discardPreview() {
  if (objectUrl.value) {
    URL.revokeObjectURL(objectUrl.value)
    objectUrl.value = null
  }
}

// 类型切换时按新规则重新校验尺寸
watch(textureType, () => {
  if (file.value)
    void validateSize(file.value)
})

const previewSkinUrl = computed(() => (textureType.value === 'cape' ? null : objectUrl.value))
const previewCapeUrl = computed(() => (textureType.value === 'cape' ? objectUrl.value : null))
const previewModel = computed<'default' | 'slim'>(() => (textureType.value === 'alex' ? 'slim' : 'default'))

async function submit() {
  if (!authStore.token) {
    error.value = '请先登录'
    return
  }
  if (!file.value) {
    error.value = '请选择文件'
    return
  }
  if (sizeError.value) {
    error.value = sizeError.value
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
    message.success('上传成功，已加入你的衣橱')
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
  <WorkspacePage title="上传纹理" description="上传 Minecraft 皮肤或披风，右侧可实时预览 3D 效果">
    <div class="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
      <!-- 表单 -->
      <NCard>
        <NSpace vertical :size="16">
          <div>
            <NText depth="3" class="mb-1 block text-sm">
              选择文件（PNG）
            </NText>
            <NUpload
              accept="image/png"
              :max="1"
              :show-file-list="false"
              :default-upload="false"
              draggable
              @change="onFileChange"
            >
              <NUploadDragger>
                <div class="flex flex-col items-center gap-2 py-8">
                  <span class="i-ph-cloud-arrow-up-duotone text-6 text-[var(--n-primary-color)]" />
                  <NText depth="3" class="text-3.5">
                    点击或拖拽 PNG 文件到此处
                  </NText>
                  <NText depth="3" class="text-3">
                    皮肤：64×32 倍数（如 64×64）· 披风：64×32 或 22×17 倍数
                  </NText>
                </div>
              </NUploadDragger>
            </NUpload>
            <NText v-if="file" depth="2" class="mt-2 flex items-center gap-1.5 text-3">
              <span class="i-ph-file-image-duotone text-4 text-[var(--n-primary-color)]" />
              {{ file.name }}（{{ (file.size / 1024).toFixed(1) }} KB）
            </NText>
          </div>

          <NAlert v-if="sizeError" type="warning" :bordered="false">
            {{ sizeError }}
          </NAlert>

          <NInput
            v-model:value="name"
            placeholder="纹理名称（默认取文件名）"
            :disabled="uploading"
            maxlength="64"
            show-count
          />

          <div>
            <NText depth="3" class="mb-2 block text-sm">
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

          <NAlert v-if="error" type="error" :bordered="false">
            {{ error }}
          </NAlert>

          <NButton
            type="primary"
            :loading="uploading"
            :disabled="!file || uploading"
            @click="submit"
          >
            <template #icon>
              <span class="i-ph-upload-simple-duotone" />
            </template>
            上传
          </NButton>
        </NSpace>
      </NCard>

      <!-- 3D 预览 -->
      <div class="w-full">
        <NCard size="small" title="实时预览" class="h-full">
          <div class="flex h-full min-h-100 flex-col">
            <div class="skin-stage relative flex-1 w-full overflow-hidden rounded-lg">
              <SkinViewer3D
                v-if="objectUrl"
                :skin-url="previewSkinUrl"
                :cape-url="previewCapeUrl"
                :model="previewModel"
                :lazy="false"
                :interactive="true"
                :auto-rotate="true"
              />
              <div v-else class="flex h-full flex-col items-center justify-center gap-2 text-[var(--n-text-color-3)]">
                <span class="i-ph-cube-duotone text-6 opacity-40" />
                <NText depth="3" class="text-3">
                  选择文件后在此预览
                </NText>
              </div>
            </div>
            <NText depth="3" class="mt-2 block text-center text-3">
              拖动角色可旋转查看
            </NText>
          </div>
        </NCard>
      </div>
    </div>
  </WorkspacePage>
</template>
