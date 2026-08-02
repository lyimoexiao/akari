<script setup lang="ts">
import { useClipboard } from '@vueuse/core'
import { useMessage } from 'naive-ui'
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getYggdrasilMetadata } from '@/api/auth'
import { getTexture } from '@/api/skinlib'
import AccountProfileCard from '@/components/user/AccountProfileCard.vue'
import TexturePickerDialog from '@/components/user/TexturePickerDialog.vue'
import YggdrasilStatusCard from '@/components/user/YggdrasilStatusCard.vue'
import WorkspacePage from '@/components/workspace/WorkspacePage.vue'
import { rawTextureUrl } from '@/composables/rawTextureUrl'
import { useYggdrasilProfile } from '@/composables/useYggdrasilProfile'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const profile = useYggdrasilProfile()
const canReadYggdrasil = computed(() => authStore.hasPermission('yggdrasil.status.read'))

const pickerType = shallowRef<'skin' | 'cape'>('skin')
const pickerTitle = shallowRef('')
const pickerShow = shallowRef(false)

// ── Yggdrasil API 地址（基于后端 server.base_url）──
const yggdrasilBaseUrl = shallowRef<string | null>(null)

const yggdrasilApiUrl = computed(() => {
  const base = yggdrasilBaseUrl.value
  return base ? `${base.replace(/\/+$/, '')}/api/v1/yggdrasil/` : null
})

const { copy } = useClipboard({ source: () => yggdrasilApiUrl.value ?? '' })

async function loadYggdrasilMetadata() {
  try {
    const metadata = await getYggdrasilMetadata()
    yggdrasilBaseUrl.value = metadata.base_url || window.location.origin
  }
  catch {
    yggdrasilBaseUrl.value = window.location.origin
  }
}

function copyYggdrasilUrl() {
  if (!yggdrasilApiUrl.value)
    return
  void copy()
    .then(() => message.success('已复制到剪贴板'))
    .catch(() => message.error('复制失败，请手动选择复制'))
}

function onYggdrasilDragStart(event: DragEvent) {
  if (!yggdrasilApiUrl.value)
    return
  const dataTransfer = event.dataTransfer
  if (!dataTransfer)
    return
  // 启动器（HMCL / PCL2 等）通过 text/uri-list 或 text/plain 识别拖入的认证服务器地址
  dataTransfer.setData('text/uri-list', yggdrasilApiUrl.value)
  dataTransfer.setData('text/plain', yggdrasilApiUrl.value)
  dataTransfer.effectAllowed = 'copy'
}

// 当前装备纹理的展示 URL
const equippedSkinUrl = shallowRef<string | null>(null)
const equippedCapeUrl = shallowRef<string | null>(null)
const equippedModel = shallowRef<'default' | 'slim'>('default')

async function loadEquipped() {
  equippedSkinUrl.value = null
  equippedCapeUrl.value = null
  const status = profile.status.value
  if (!status || !authStore.token)
    return
  try {
    if (status.texture_skin_id) {
      const skin = await getTexture(authStore.token, status.texture_skin_id)
      equippedSkinUrl.value = rawTextureUrl(skin.hash, skin.url)
      equippedModel.value = skin.type === 'alex' ? 'slim' : 'default'
    }
    if (status.texture_cape_id) {
      const cape = await getTexture(authStore.token, status.texture_cape_id)
      equippedCapeUrl.value = rawTextureUrl(cape.hash, cape.url)
    }
  }
  catch {
    // 纹理读取失败时展示默认模型
  }
}

watch(profile.status, () => {
  void loadEquipped()
})

onMounted(() => {
  void profile.refresh()
  void loadYggdrasilMetadata()
})

function openSetSkin() {
  pickerType.value = 'skin'
  pickerTitle.value = '选择皮肤'
  pickerShow.value = true
}

function openSetCape() {
  pickerType.value = 'cape'
  pickerTitle.value = '选择披风'
  pickerShow.value = true
}

async function onTextureSelected(tid: number) {
  const ok = pickerType.value === 'cape'
    ? await profile.setCape(tid)
    : await profile.setSkin(tid)
  if (ok)
    void loadEquipped()
}

async function onClearSkin() {
  await profile.clearSkin()
  void loadEquipped()
}

async function onClearCape() {
  await profile.clearCape()
  void loadEquipped()
}
</script>

<template>
  <WorkspacePage title="账户概览" :description="`欢迎回来，${authStore.username ?? '用户'}`">
    <template #actions>
      <NButton secondary @click="router.push({ name: 'UserCloset' })">
        我的衣橱
      </NButton>
      <NButton type="primary" @click="router.push({ name: 'UserTextureUpload' })">
        上传皮肤
      </NButton>
    </template>

    <NAlert v-if="profile.error.value" type="error" class="mb-4">
      {{ profile.error.value }}
      <template #action>
        <NButton text type="error" @click="profile.refresh()">
          重试
        </NButton>
      </template>
    </NAlert>

    <div class="grid gap-4 lg:grid-cols-3">
      <AccountProfileCard
        v-if="authStore.user"
        :user="authStore.user"
        @verify-email="router.push({ name: 'AuthVerifyEmail' })"
      />
      <div class="lg:col-span-2">
        <YggdrasilStatusCard
          :status="profile.status.value"
          :loading="profile.loading.value"
          :available="canReadYggdrasil"
          :skin-url="equippedSkinUrl"
          :cape-url="equippedCapeUrl"
          :model="equippedModel"
          @set-skin="openSetSkin"
          @set-cape="openSetCape"
          @clear-skin="onClearSkin"
          @clear-cape="onClearCape"
        />

        <!-- Yggdrasil API 地址 -->
        <NCard size="small" class="mt-4">
          <template #header>
            <NFlex align="center" :size="8">
              <span class="i-ph-link-duotone text-4 text-[var(--n-primary-color)]" />
              <NText strong>
                Yggdrasil API 地址
              </NText>
            </NFlex>
          </template>

          <NFlex v-if="yggdrasilApiUrl" :wrap="true" align="center" :size="12">
            <div
              draggable="true"
              class="cursor-grab select-all rounded-lg border border-dashed border-[var(--n-border-color)] bg-[var(--n-input-color)] px-3 py-2 font-mono text-3 text-[var(--n-primary-color)] transition-colors hover:border-[var(--n-primary-color)] active:cursor-grabbing"
              title="拖拽到启动器（HMCL / PCL2 等）窗口即可添加"
              @dragstart="onYggdrasilDragStart"
            >
              <span class="i-ph-cube-duotone mr-1.5 opacity-60" />
              {{ yggdrasilApiUrl }}
            </div>
            <NButton size="small" secondary @click="copyYggdrasilUrl">
              <template #icon>
                <span class="i-ph-copy-duotone" />
              </template>
              复制
            </NButton>
          </NFlex>

          <NText depth="3" class="mt-2 block text-3">
            将该地址填入启动器的 Yggdrasil 认证服务器，或直接拖拽到 HMCL / PCL2 窗口自动添加
          </NText>
        </NCard>
      </div>
    </div>

    <TexturePickerDialog
      v-model:show="pickerShow"
      :title="pickerTitle"
      :type="pickerType"
      @select="onTextureSelected"
    />
  </WorkspacePage>
</template>
