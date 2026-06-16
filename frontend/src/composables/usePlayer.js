import { ref, onUnmounted } from 'vue'
import { getProxyImageUrl } from './url'

export function usePlayer() {
  const videoPlayerRef = ref(null)
  const previewPlayerRef = ref(null)
  let dp = null
  let previewDp = null

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

  const createPlayer = async (videoUrl, coverUrl) => {
    if (!videoPlayerRef.value) return
    await loadDPlayer()
    if (dp) dp.destroy()
    dp = new window.DPlayer({
      container: videoPlayerRef.value,
      video: {
        url: videoUrl,
        pic: getProxyImageUrl(coverUrl)
      }
    })
  }

  const createPreviewPlayer = async (videoUrl, coverUrl) => {
    if (!previewPlayerRef.value) return
    await loadDPlayer()
    if (previewDp) previewDp.destroy()
    previewDp = new window.DPlayer({
      container: previewPlayerRef.value,
      autoplay: true,
      theme: '#ff2442',
      loop: true,
      screenshot: true,
      hotkey: true,
      preload: 'auto',
      volume: 0.7,
      mutex: true,
      video: {
        url: videoUrl,
        pic: coverUrl,
        type: 'auto'
      }
    })
  }

  const destroyPlayer = () => {
    if (dp) { dp.destroy(); dp = null }
  }

  const destroyPreviewPlayer = () => {
    if (previewDp) { previewDp.destroy(); previewDp = null }
  }

  onUnmounted(() => {
    destroyPlayer()
    destroyPreviewPlayer()
  })

  return {
    videoPlayerRef,
    previewPlayerRef,
    createPlayer,
    createPreviewPlayer,
    destroyPlayer,
    destroyPreviewPlayer
  }
}
