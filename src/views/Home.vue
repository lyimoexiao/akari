<script setup lang="ts">
import type { TextureItem } from '@/types/skinlib'
import { useAsyncState } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { listTextures } from '@/api/skinlib'
import TextureCard from '@/components/skin/TextureCard.vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// ── 最新上传 ──
const {
  state: latest,
  isLoading: latestLoading,
  execute: refreshLatest,
} = useAsyncState(
  () => listTextures(null, { page: 1, page_size: 4, order: 'latest' }),
  { items: [] as readonly TextureItem[], total: 0, page: 1, page_size: 4, total_pages: 0 },
)

void refreshLatest()
</script>

<template>
  <div class="mx-auto max-w-320 px-4 sm:px-6">
    <!-- Hero -->
    <section class="py-14 text-center sm:py-20">
      <div class="mx-auto max-w-180">
        <h1 class="font-minecraft text-8 leading-tight text-[var(--n-primary-color)] sm:text-9">
          Akari
        </h1>
        <NText class="mt-4 block text-4.5 font-600 sm:text-5">
          分享、收藏你的 Minecraft 皮肤与披风
        </NText>
        <NText depth="3" class="mx-auto mt-3 block max-w-150 text-3.5 leading-6">
          基于 Yggdrasil 协议的皮肤站，支持 authlib-injector 客户端直连。
          上传皮肤、管理衣橱，在网页上实时预览 3D 角色。
        </NText>
        <NFlex class="mt-8" justify="center" :wrap="true" :size="12">
          <NButton type="primary" size="large" @click="router.push({ name: 'Skinlib' })">
            <template #icon>
              <span class="i-ph-t-shirt-duotone" />
            </template>
            浏览皮肤库
          </NButton>
          <NButton v-if="authStore.isAuthenticated" size="large" secondary @click="router.push({ name: 'UserTextureUpload' })">
            <template #icon>
              <span class="i-ph-upload-simple-duotone" />
            </template>
            上传皮肤
          </NButton>
          <NButton v-else size="large" secondary @click="router.push({ name: 'AuthRegister' })">
            <template #icon>
              <span class="i-ph-user-plus-duotone" />
            </template>
            注册账户
          </NButton>
        </NFlex>
      </div>
    </section>

    <!-- 最新上传 -->
    <section class="pb-12">
      <NFlex justify="space-between" align="center" class="mb-4">
        <div>
          <div class="flex items-center gap-2">
            <span class="i-ph-lightning-duotone text-4 text-[var(--n-primary-color)]" />
            <NH2 class="m-0! text-4.5! font-700!">
              最新上传
            </NH2>
          </div>
          <NText depth="3" class="text-3">
            社区刚刚分享的皮肤与披风
          </NText>
        </div>
        <NButton text type="primary" @click="router.push({ name: 'Skinlib' })">
          查看全部 →
        </NButton>
      </NFlex>

      <div v-if="latestLoading" class="flex justify-center py-12">
        <NSpin />
      </div>
      <div v-else-if="latest.items.length === 0" class="py-12">
        <NEmpty description="还没有人上传过皮肤，来做第一个吧" />
      </div>
      <div v-else class="grid grid-cols-2 gap-4 md:grid-cols-4">
        <TextureCard
          v-for="texture in latest.items"
          :key="texture.tid"
          :data="texture"
          class="cursor-pointer"
          @click="router.push({ name: 'Skinlib' })"
        />
      </div>
    </section>

    <!-- 特性 -->
    <section class="grid gap-4 pb-16 sm:grid-cols-3">
      <NCard hoverable class="transition-transform duration-200 hover:-translate-y-1" @click="router.push({ name: 'Skinlib' })">
        <template #header>
          <NFlex align="center" :size="8">
            <span class="i-ph-t-shirt-duotone text-5 text-[var(--n-primary-color)]" />
            <NText strong class="text-4">
              皮肤库
            </NText>
          </NFlex>
        </template>
        <NText depth="3" class="text-3.5">
          浏览社区分享的皮肤与披风，一键收藏进你的衣橱。
        </NText>
      </NCard>

      <NCard hoverable class="transition-transform duration-200 hover:-translate-y-1" @click="router.push({ name: 'UserCloset' })">
        <template #header>
          <NFlex align="center" :size="8">
            <span class="i-iconoir-closet text-5 text-[var(--n-primary-color)]" />
            <NText strong class="text-4">
              我的衣橱
            </NText>
          </NFlex>
        </template>
        <NText depth="3" class="text-3.5">
          管理收藏的皮肤与披风，随时换装，3D 实时预览。
        </NText>
      </NCard>

      <NCard hoverable class="transition-transform duration-200 hover:-translate-y-1" @click="router.push({ name: 'AuthLogin' })">
        <template #header>
          <NFlex align="center" :size="8">
            <span class="i-ph-cube-duotone text-5 text-[var(--n-primary-color)]" />
            <NText strong class="text-4">
              Yggdrasil 直连
            </NText>
          </NFlex>
        </template>
        <NText depth="3" class="text-3.5">
          通过 authlib-injector 在游戏内使用你的皮肤，支持 Slim 模型。
        </NText>
      </NCard>
    </section>
  </div>
</template>
