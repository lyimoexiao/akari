<script setup lang="ts">
import type { SkinViewer } from 'skinview3d'
import { useIntersectionObserver } from '@vueuse/core'
import { onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'

type ModelType = 'default' | 'slim'

/**
 * skinview3d 封装的 3D 玩家预览。
 *
 * - 默认懒加载：仅当组件进入视口时才实例化 WebGL 渲染器，
 *   离开视口自动 dispose 以释放 WebGL context（浏览器有 ~16 个上限）。
 * - 纹理加载完成前 / 加载失败时回退到 PNG 缩略图（fallbackUrl）。
 */
const props = withDefaults(defineProps<{
  skinUrl?: string | null
  capeUrl?: string | null
  model?: ModelType | 'auto-detect'
  /** 是否自动缓慢旋转 */
  autoRotate?: boolean
  /** 是否允许鼠标/触摸拖拽旋转 */
  interactive?: boolean
  /** 停止操作多少秒后平滑复位到初始视角（0 = 不复位） */
  idleReset?: number
  /** canvas 像素比上限（性能控制） */
  pixelRatio?: number
  /** 是否仅在进入视口时渲染 */
  lazy?: boolean
  zoom?: number
  fov?: number
  /** 3D 渲染完成/失败前的 PNG 兜底图 */
  fallbackUrl?: string | null
}>(), {
  skinUrl: null,
  capeUrl: null,
  model: 'auto-detect',
  autoRotate: true,
  interactive: true,
  idleReset: 0,
  pixelRatio: 1.5,
  lazy: true,
  zoom: 0.9,
  fov: 50,
  fallbackUrl: null,
})

const containerRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

const viewer = shallowRef<SkinViewer | null>(null)
const renderFailed = ref(false)
const fallbackVisible = ref(false)

let modulePromise: Promise<typeof import('skinview3d')> | null = null
let resizeObserver: ResizeObserver | null = null
let stopObserver: (() => void) | null = null
let resumeRotateTimer: ReturnType<typeof setTimeout> | null = null
let idleResetTimer: ReturnType<typeof setTimeout> | null = null
let resetAnimFrame: number | null = null

type Vec3 = ReturnType<SkinViewer['camera']['position']['clone']>
/** 每个实例独立的初始视角（多实例共享模块变量会互相覆盖） */
const initialPoses = new WeakMap<SkinViewer, { pos: Vec3, target: Vec3 }>()

function loadModule() {
  modulePromise ??= import('skinview3d')
  return modulePromise
}

async function mountViewer() {
  const container = containerRef.value
  const canvas = canvasRef.value
  if (!container || !canvas || viewer.value)
    return

  const { SkinViewer } = await loadModule()
  if (!containerRef.value || !canvasRef.value || viewer.value)
    return // 组件已卸载或已在别处挂载

  fallbackVisible.value = !!props.fallbackUrl
  try {
    const instance = new SkinViewer({
      canvas,
      width: container.clientWidth || 300,
      height: container.clientHeight || 400,
      enableControls: props.interactive,
      zoom: props.zoom,
      fov: props.fov,
    })
    instance.pixelRatio = Math.min(props.pixelRatio, window.devicePixelRatio || 1)
    instance.autoRotate = props.autoRotate

    // 记录初始视角，供 idle 复位使用
    initialPoses.set(instance, {
      pos: instance.camera.position.clone(),
      target: instance.controls.target.clone(),
    })

    viewer.value = instance
    renderFailed.value = false

    if (props.interactive) {
      canvas.addEventListener('pointerdown', pauseAutoRotate)
      canvas.addEventListener('pointerup', scheduleResumeRotate)
      canvas.addEventListener('pointerleave', scheduleResumeRotate)
    }

    await applyTextures(instance)
  }
  catch {
    // WebGL 不可用或初始化失败 → 回退 PNG
    renderFailed.value = true
  }
}

async function applyTextures(instance: SkinViewer) {
  try {
    if (props.skinUrl)
      await instance.loadSkin(props.skinUrl, { model: props.model })
    if (props.capeUrl)
      await instance.loadCape(props.capeUrl)
    renderFailed.value = false
    fallbackVisible.value = false
  }
  catch {
    renderFailed.value = true
  }
}

function disposeViewer() {
  const canvas = canvasRef.value
  if (canvas) {
    canvas.removeEventListener('pointerdown', pauseAutoRotate)
    canvas.removeEventListener('pointerup', scheduleResumeRotate)
    canvas.removeEventListener('pointerleave', scheduleResumeRotate)
  }
  cancelViewReset()
  if (resumeRotateTimer) {
    clearTimeout(resumeRotateTimer)
    resumeRotateTimer = null
  }
  viewer.value?.dispose()
  viewer.value = null
}

function pauseAutoRotate() {
  cancelViewReset()
  if (viewer.value)
    viewer.value.autoRotate = false
}

function scheduleResumeRotate() {
  if (resumeRotateTimer)
    clearTimeout(resumeRotateTimer)
  resumeRotateTimer = setTimeout(() => {
    if (viewer.value)
      viewer.value.autoRotate = props.autoRotate
    resumeRotateTimer = null
  }, 400)
  // 停止操作后，几秒内无新操作则复位视角
  startIdleReset()
}

// ── 操作停止后自动复位视角 ──

function startIdleReset() {
  if (!props.idleReset || !viewer.value)
    return
  if (idleResetTimer)
    clearTimeout(idleResetTimer)
  idleResetTimer = setTimeout(() => {
    idleResetTimer = null
    resetView()
  }, props.idleReset * 1000)
}

function cancelViewReset() {
  if (idleResetTimer) {
    clearTimeout(idleResetTimer)
    idleResetTimer = null
  }
  if (resetAnimFrame !== null) {
    cancelAnimationFrame(resetAnimFrame)
    resetAnimFrame = null
  }
}

/** 平滑插值回到初始视角（400ms easeOutCubic） */
function resetView() {
  const instance = viewer.value
  const init = instance ? initialPoses.get(instance) : null
  if (!instance || !init)
    return
  const fromPos = instance.camera.position.clone()
  const fromTarget = instance.controls.target.clone()
  const duration = 400
  const start = performance.now()
  const easeOut = (t: number) => 1 - (1 - t) ** 3

  const step = (now: number) => {
    const t = Math.min((now - start) / duration, 1)
    const k = easeOut(t)
    instance.camera.position.lerpVectors(fromPos, init.pos, k)
    instance.controls.target.lerpVectors(fromTarget, init.target, k)
    instance.controls.update()
    if (t < 1)
      resetAnimFrame = requestAnimationFrame(step)
    else
      resetAnimFrame = null
  }
  resetAnimFrame = requestAnimationFrame(step)
}

// ── 视口懒渲染 ──
if (props.lazy) {
  stopObserver = useIntersectionObserver(containerRef, (entries) => {
    const entry = entries[0]
    if (!entry)
      return
    if (entry.isIntersecting)
      void mountViewer()
    else
      disposeViewer()
  }, { threshold: 0.05 }).stop
}

// ── 尺寸自适应 ──
onMounted(() => {
  if (!props.lazy)
    void mountViewer()

  if (containerRef.value) {
    resizeObserver = new ResizeObserver(() => {
      const instance = viewer.value
      const container = containerRef.value
      if (instance && container) {
        instance.width = container.clientWidth
        instance.height = container.clientHeight
      }
    })
    resizeObserver.observe(containerRef.value)
  }
})

// ── 纹理变化 ──
watch(() => props.skinUrl, (url) => {
  const instance = viewer.value
  if (!instance)
    return
  fallbackVisible.value = !!props.fallbackUrl
  if (url) {
    void instance.loadSkin(url, { model: props.model })
      .then(() => { fallbackVisible.value = false })
      .catch(() => { renderFailed.value = true })
  }
  else {
    instance.loadSkin(null)
  }
})

watch(() => props.capeUrl, (url) => {
  const instance = viewer.value
  if (!instance)
    return
  fallbackVisible.value = !!props.fallbackUrl
  if (url) {
    void instance.loadCape(url)
      .then(() => { fallbackVisible.value = false })
      .catch(() => { renderFailed.value = true })
  }
  else {
    instance.loadCape(null)
  }
})

watch(() => props.model, (model) => {
  const instance = viewer.value
  if (!instance || !props.skinUrl)
    return
  void instance.loadSkin(props.skinUrl, { model })
})

watch(() => props.autoRotate, (value) => {
  if (viewer.value)
    viewer.value.autoRotate = value
})

// ── 卸载 ──
onBeforeUnmount(() => {
  stopObserver?.()
  resizeObserver?.disconnect()
  disposeViewer()
})
</script>

<template>
  <div ref="containerRef" class="skin-stage relative h-full w-full overflow-hidden">
    <canvas ref="canvasRef" class="h-full w-full" />
    <img
      v-if="(fallbackVisible || renderFailed) && fallbackUrl"
      :src="fallbackUrl"
      :alt="fallbackUrl"
      class="img-pixel absolute inset-0 h-full w-full object-contain p-2"
      loading="lazy"
    >
  </div>
</template>
