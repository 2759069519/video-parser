<template>
  <div class="result-card" :class="{ active: result && result.type === 'video' }">
    <div v-if="result && result.type === 'video'">
      <div class="video-container" ref="containerRef"></div>
      <div class="video-info">
        <div class="video-title">{{ result.title }}</div>
        <div class="video-meta">
          <div class="meta-author">
            <span class="meta-name">{{ result.author }}</span>
          </div>
          <div class="meta-stats">
            <span><i class="icon-heart"></i> {{ formatNumber(result.like_count) }}</span>
            <span><i class="icon-message-circle"></i> {{ formatNumber(result.comment_count) }}</span>
            <span><i class="icon-play"></i> {{ formatNumber(result.view_count) }}</span>
          </div>
        </div>
        <div class="video-actions">
          <button class="btn-action primary" @click="$emit('download-video', result.video_url)">
            <i class="icon-download"></i> 下载视频
          </button>
          <button class="btn-action" @click="$emit('copy-link', result.video_url)">
            <i class="icon-copy"></i> 复制链接
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onUnmounted } from 'vue'
import { getProxyImageUrl } from '../composables/url'

const props = defineProps({ 
  result: Object,
  videoUrl: String,
  coverUrl: String
})
defineEmits(['download-video', 'copy-link'])

const containerRef = ref(null)
let dp = null

const loadDPlayer = () => {
  return new Promise((resolve) => {
    if (window.DPlayer) {
      resolve()
      return
    }
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/dplayer@1.27.1/dist/DPlayer.min.js'
    script.onload = resolve
    document.head.appendChild(script)
  })
}

const createPlayer = async (url, cover) => {
  if (!containerRef.value) return
  await loadDPlayer()
  if (dp) dp.destroy()
  dp = new window.DPlayer({
    container: containerRef.value,
    video: {
      url: url,
      pic: getProxyImageUrl(cover)
    }
  })
}

const destroyPlayer = () => {
  if (dp) { dp.destroy(); dp = null }
}

watch(() => props.videoUrl, (newUrl) => {
  if (newUrl && props.result && props.result.type === 'video') {
    createPlayer(newUrl, props.coverUrl)
  }
}, { immediate: true })

onUnmounted(() => {
  destroyPlayer()
})

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) return (num / 10000).toFixed(1) + '万'
  return num.toString()
}

defineExpose({ destroyPlayer })
</script>
