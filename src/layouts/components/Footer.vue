<script setup lang="ts">
import { useNow } from '@vueuse/core'
import { computed, onMounted, shallowRef } from 'vue'
import { getYggdrasilMetadata } from '@/api/auth'

const now = useNow({ interval: 60_000 })
const year = computed(() => now.value.getFullYear())

// 站点标题与描述：用户可在后端配置（yggdrasil.server_name / description），未获取到时回退默认值
const siteTitle = shallowRef('Akari')
const siteDescription = shallowRef('Minecraft 皮肤站 · 支持 Yggdrasil 与 authlib-injector')

onMounted(async () => {
  try {
    const metadata = await getYggdrasilMetadata()
    const serverName = metadata.meta.serverName
    if (typeof serverName === 'string' && serverName.trim())
      siteTitle.value = serverName.trim()
    const description = metadata.meta.description
    if (typeof description === 'string' && description.trim())
      siteDescription.value = description.trim()
  }
  catch {
    // 保持默认标题与描述
  }
})
</script>

<template>
  <NLayoutFooter bordered>
    <div class="mx-auto max-w-320 px-4 py-6 sm:px-6">
      <div class="divider-glow mb-5 w-24" />
      <NFlex justify="space-between" align="center" :wrap="true" class="gap-3">
        <NFlex vertical :size="2">
          <NText class="font-minecraft text-4 text-[var(--n-primary-color)]">
            {{ siteTitle }}
          </NText>
          <NText depth="3" class="text-3">
            {{ siteDescription }}
          </NText>
        </NFlex>
        <NText depth="3" class="text-3">
          Copyright &copy; {{ year }} {{ siteTitle }} · Powered by
          <NButton text tag="a" type="primary" target="_blank" rel="noopener noreferrer" href="https://github.com/lyimoexiao/akari">
            Akari Skin Server
          </NButton>
        </NText>
      </NFlex>
    </div>
  </NLayoutFooter>
</template>
