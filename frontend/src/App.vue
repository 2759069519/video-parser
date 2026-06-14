<template>
  <div class="container">
    <div class="top-bar">
      <a href="/" class="brand">
        <i class="icon-layers"></i>
        <span>聚合解析</span>
      </a>
    </div>

    <div class="header">
      <h1>聚合解析</h1>
      <p>支持视频、图文、主页解析</p>
    </div>

    <div class="search-box">
      <div class="search-input">
        <i class="icon-link"></i>
        <input 
          type="text" 
          v-model="url" 
          placeholder="粘贴视频/图文/主页链接" 
          autocomplete="off"
          @keyup.enter="parse"
        >
      </div>
      <button class="btn-parse" @click="parse" :disabled="loading">
        <div v-if="loading" class="spinner"></div>
        <i v-else class="icon-search"></i>
        <span>{{ loading ? '解析中' : '解析' }}</span>
      </button>
    </div>

    <div class="progress-bar" :class="{ active: loading }">
      <div class="progress-fill" :style="{ width: loading ? '100%' : '0%' }"></div>
    </div>

    <!-- 视频结果 -->
    <div class="result-card" :class="{ active: result && result.type === 'video' }">
      <div v-if="result && result.type === 'video'">
        <div class="video-container" ref="videoPlayerRef"></div>
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
            <button class="btn-action primary" @click="downloadVideo(result.video_url)">
              <i class="icon-download"></i> 下载视频
            </button>
            <button class="btn-action" @click="copyLink(result.video_url)">
              <i class="icon-copy"></i> 复制链接
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 图集结果 -->
    <div class="result-card" :class="{ active: result && result.type === 'atlas' }">
      <div v-if="result && result.type === 'atlas'">
        <div class="atlas-info">
          <div class="video-title">{{ result.title }}</div>
          <div class="video-meta">
            <div class="meta-author">
              <span class="meta-name">{{ result.author }}</span>
            </div>
            <div class="meta-stats">
              <span><i class="icon-heart"></i> {{ formatNumber(result.like_count) }}</span>
              <span><i class="icon-image"></i> {{ result.images ? result.images.length : 0 }}张</span>
            </div>
          </div>
        </div>
        <div class="atlas-grid">
          <div v-for="(img, index) in result.images" :key="index" class="atlas-item" @click="openImagePreview(index)">
            <img :src="getProxyImageUrl(img.url)" :alt="'图片 ' + (index + 1)" crossorigin="anonymous">
            <button class="atlas-download" @click.stop="downloadSingleImage(img.url, index)">
              <i class="icon-download"></i>
            </button>
          </div>
        </div>
        <div class="atlas-actions">
          <button class="btn-action primary" @click="downloadAllImages(result.images)">
            <i class="icon-download"></i> 下载全部图片
          </button>
        </div>
      </div>
    </div>

    <!-- 主页结果 -->
    <div class="result-card" :class="{ active: result && result.type === 'profile' }">
      <div v-if="result && result.type === 'profile'">
        <div class="profile-header">
          <img class="profile-avatar" :src="result.avatar" :alt="result.user_name">
          <div class="profile-info">
            <h3>{{ result.user_name }}</h3>
            <p>{{ result.description }}</p>
          </div>
        </div>
        <div class="profile-stats">
          <div class="stat-item">
            <div class="stat-value">{{ formatNumber(result.fan_count) }}</div>
            <div class="stat-label">粉丝</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ result.photos ? result.photos.length : 0 }}</div>
            <div class="stat-label">作品</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ formatNumber(result.follow_count) }}</div>
            <div class="stat-label">关注</div>
          </div>
        </div>
        <div class="video-list">
          <div v-for="photo in result.photos" :key="photo.photo_id" class="video-item" @click="openPreview(photo)">
            <img :src="photo.cover_url" :alt="photo.caption">
            <div class="overlay">
              <div class="title">{{ photo.caption }}</div>
              <div class="stats">
                <span><i class="icon-heart"></i> {{ formatNumber(photo.like_count) }}</span>
                <span><i class="icon-play"></i> {{ formatNumber(photo.view_count) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <div class="error-toast" :class="{ show: errorMsg }">
    {{ errorMsg }}
  </div>

  <div class="copy-toast" :class="{ show: copied }">已复制</div>

  <div class="success-toast" :class="{ show: successMsg }">
    <i class="icon-check-circle"></i>
    <span>{{ successMsg }}</span>
  </div>

  <!-- 视频预览弹窗 -->
  <div class="video-preview-modal" :class="{ active: previewVisible }" @click.self="closePreview">
    <div class="video-preview-content">
      <div class="video-preview-header">
        <div class="title">{{ previewTitle }}</div>
        <button class="video-preview-close" @click="closePreview">
          <i class="icon-x"></i>
        </button>
      </div>
      <div ref="previewPlayerRef"></div>
      <div class="video-preview-stats">{{ previewStats }}</div>
    </div>
  </div>

  <!-- 图片预览弹窗 -->
  <div class="image-preview-modal" :class="{ active: imagePreviewVisible }" @click.self="closeImagePreview">
    <div class="image-preview-content">
      <button class="image-preview-close" @click="closeImagePreview">
        <i class="icon-x"></i>
      </button>
      <div class="image-preview-counter">{{ imagePreviewIndex + 1 }} / {{ imagePreviewList.length }}</div>
      <div class="image-preview-wrapper">
        <button class="image-preview-prev" @click="prevImage" v-if="imagePreviewList.length > 1">
          <i class="icon-chevron-left"></i>
        </button>
        <img :src="getProxyImageUrl(imagePreviewList[imagePreviewIndex]?.url)" :alt="'图片 ' + (imagePreviewIndex + 1)" class="image-preview-img" crossorigin="anonymous">
        <button class="image-preview-next" @click="nextImage" v-if="imagePreviewList.length > 1">
          <i class="icon-chevron-right"></i>
        </button>
      </div>
      <div class="image-preview-info">
        <div class="title">{{ imagePreviewTitle }}</div>
        <div class="stats">{{ imagePreviewStats }}</div>
      </div>
      <div class="image-preview-actions">
        <button class="btn-action" @click="downloadSingleImage(imagePreviewList[imagePreviewIndex]?.url, imagePreviewIndex)">
          <i class="icon-download"></i> 下载当前图片
        </button>
        <button class="btn-action primary" @click="downloadAllImages(imagePreviewList)">
          <i class="icon-download"></i> 下载全部
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, onUnmounted } from 'vue'

const url = ref('')
const loading = ref(false)
const result = ref(null)
const errorMsg = ref('')
const copied = ref(false)
const successMsg = ref('')
const videoPlayerRef = ref(null)
const previewPlayerRef = ref(null)
const previewVisible = ref(false)
const previewLoading = ref(false)
const previewTitle = ref('')
const previewStats = ref('')
const imagePreviewVisible = ref(false)
const imagePreviewList = ref([])
const imagePreviewIndex = ref(0)
const imagePreviewTitle = ref('')
const imagePreviewStats = ref('')
let dp = null
let previewDp = null

const parse = async () => {
  if (!url.value.trim()) {
    showError('请输入链接')
    return
  }

  loading.value = true
  result.value = null

  // 销毁旧播放器
  if (dp) {
    dp.destroy()
    dp = null
  }

  try {
    const response = await fetch('/api/parse', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: url.value })
    })

    const data = await response.json()
    if (data.success) {
      result.value = { ...data.data, type: data.type }
      showSuccess('解析成功')
      
      // 如果是视频，初始化播放器
      if (data.type === 'video' && data.data.video_url) {
        await nextTick()
        initPlayer(data.data.video_url, data.data.cover_url)
      }
    } else {
      showError(data.error || '解析失败')
    }
  } catch (error) {
    showError('请求失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const initPlayer = (videoUrl, coverUrl) => {
  if (!videoPlayerRef.value) return
  
  // 动态加载 DPlayer
  if (!window.DPlayer) {
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/dplayer@1.27.1/dist/DPlayer.min.js'
    script.onload = () => {
      createPlayer(videoUrl, coverUrl)
    }
    document.head.appendChild(script)
  } else {
    createPlayer(videoUrl, coverUrl)
  }
}

const createPlayer = (videoUrl, coverUrl) => {
  if (!videoPlayerRef.value || !window.DPlayer) return
  
  dp = new window.DPlayer({
    container: videoPlayerRef.value,
    video: {
      url: videoUrl,
      pic: getProxyImageUrl(coverUrl)
    }
  })
}

onUnmounted(() => {
  if (dp) {
    dp.destroy()
  }
  if (previewDp) {
    previewDp.destroy()
  }
})

const openPreview = async (photo) => {
  if (!photo.photo_id) return
  
  // 如果是图文类型，获取图片列表并显示图片预览
  if (photo.type === 'HORIZONTAL_ATLAS') {
    openAtlasPreview(photo)
    return
  }
  
  previewTitle.value = photo.caption || ''
  previewStats.value = `❤ ${formatNumber(photo.like_count)}  ▶ ${formatNumber(photo.view_count)}`
  previewVisible.value = true
  previewLoading.value = true
  
  // 获取视频URL
  try {
    const response = await fetch('/api/fetch-video-url', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ photo_id: photo.photo_id, platform: 'kuaishou' })
    })
    
    const data = await response.json()
    if (data.success && data.data.video_url) {
      await nextTick()
      initPreviewPlayer(data.data.video_url, photo.cover_url)
    } else {
      showError(data.error || '获取视频失败')
      closePreview()
    }
  } catch (error) {
    showError('获取视频失败，请稍后重试')
    closePreview()
  } finally {
    previewLoading.value = false
  }
}

const openAtlasPreview = async (photo) => {
  imagePreviewTitle.value = photo.caption || ''
  imagePreviewStats.value = `❤ ${formatNumber(photo.like_count)}`
  imagePreviewIndex.value = 0
  
  // 获取图文作品的图片列表
  try {
    const response = await fetch('/api/fetch-atlas-images', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ photo_id: photo.photo_id, platform: 'kuaishou' })
    })
    
    const data = await response.json()
    if (data.success && data.data.images) {
      imagePreviewList.value = data.data.images
      imagePreviewVisible.value = true
    } else {
      showError(data.error || '获取图片失败')
    }
  } catch (error) {
    showError('获取图片失败，请稍后重试')
  }
}

const prevImage = () => {
  if (imagePreviewIndex.value > 0) {
    imagePreviewIndex.value--
  } else {
    imagePreviewIndex.value = imagePreviewList.value.length - 1
  }
}

const nextImage = () => {
  if (imagePreviewIndex.value < imagePreviewList.value.length - 1) {
    imagePreviewIndex.value++
  } else {
    imagePreviewIndex.value = 0
  }
}

const closeImagePreview = () => {
  imagePreviewVisible.value = false
  imagePreviewList.value = []
  imagePreviewIndex.value = 0
}

const initPreviewPlayer = (videoUrl, coverUrl) => {
  if (!previewPlayerRef.value) return
  
  // 动态加载 DPlayer
  if (!window.DPlayer) {
    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/dplayer@1.27.1/dist/DPlayer.min.js'
    script.onload = () => {
      createPreviewPlayer(videoUrl, coverUrl)
    }
    document.head.appendChild(script)
  } else {
    createPreviewPlayer(videoUrl, coverUrl)
  }
}

const createPreviewPlayer = (videoUrl, coverUrl) => {
  if (!previewPlayerRef.value || !window.DPlayer) return
  
  if (previewDp) {
    previewDp.destroy()
  }
  
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

const closePreview = () => {
  previewVisible.value = false
  if (previewDp) {
    previewDp.destroy()
    previewDp = null
  }
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

const showError = (msg) => {
  errorMsg.value = msg
  setTimeout(() => { errorMsg.value = '' }, 3000)
}

const showSuccess = (msg) => {
  successMsg.value = msg
  setTimeout(() => { successMsg.value = '' }, 2000)
}

const copyLink = (link) => {
  navigator.clipboard.writeText(link).then(() => {
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }).catch(() => {
    showError('复制失败')
  })
}

const getProxyImageUrl = (url) => {
  if (!url) return ''
  if (url.includes('xhscdn.com') || url.includes('xiaohongshu.com') || url.includes('sns-') ||
      url.includes('douyin.com') || url.includes('kuaishou.com') || url.includes('gifshow.com')) {
    return `/api/proxy-image?url=${encodeURIComponent(url)}`
  }
  return url
}

const downloadSingleImage = async (url, index) => {
  const ext = (url.match(/\.(jpg|jpeg|png|webp|heic|avif)(\?|$)/i) || [null, 'jpg'])[1] || 'jpg'
  const filename = `image_${index + 1}.${ext}`
  const downloadUrl = `/api/download?url=${encodeURIComponent(url)}&filename=${filename}`
  window.location.href = downloadUrl
}

const downloadVideo = async (videoUrl) => {
  const downloadUrl = `/api/download?url=${encodeURIComponent(videoUrl)}&filename=video.mp4`
  window.location.href = downloadUrl
}

const downloadAllImages = async (images) => {
  // 支持传入图片列表或使用默认的图片列表
  const imageList = images || (result.value && result.value.images) || imagePreviewList.value
  
  if (!imageList || imageList.length === 0) {
    showError('没有可下载的图片')
    return
  }
  
  showSuccess('正在打包图片...')
  
  try {
    const JSZip = (await import('jszip')).default
    const zip = new JSZip()
    const folder = zip.folder('images')
    
    // 并行下载所有图片
    const promises = imageList.map(async (image, index) => {
      const ext = (image.url.match(/\.(jpg|jpeg|png|webp|heic|avif)(\?|$)/i) || [null, 'jpg'])[1] || 'jpg'
      const downloadUrl = `/api/download?url=${encodeURIComponent(image.url)}&filename=image_${index + 1}.${ext}`
      const response = await fetch(downloadUrl)
      const blob = await response.blob()
      folder.file(`image_${index + 1}.${ext}`, blob)
    })
    
    await Promise.all(promises)
    
    // 生成 zip 文件
    const content = await zip.generateAsync({ type: 'blob' })
    
    // 下载 zip 文件
    const { saveAs } = await import('file-saver')
    saveAs(content, 'images.zip')
  } catch (error) {
    showError('打包下载失败: ' + error.message)
  }
}
</script>

<style>
@import url('https://cdn.jsdelivr.net/npm/lucide-static@latest/font/lucide.min.css');

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f5f5f5;
  min-height: 100vh;
}

.container {
  max-width: 800px;
  margin: 0 auto;
  padding: 40px 20px;
}

.header {
  text-align: center;
  margin-bottom: 40px;
}

.header h1 {
  font-size: 28px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 8px;
}

.header p {
  font-size: 14px;
  color: #666;
}

.top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding: 0 4px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
}

.brand i {
  font-size: 22px;
  color: #ff2442;
}

.brand span {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
}

.search-box {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
  padding: 6px;
  display: flex;
  gap: 6px;
  margin-bottom: 24px;
}

.search-input {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
}

.search-input i {
  color: #999;
  font-size: 20px;
}

.search-input input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 15px;
  padding: 14px 0;
  background: transparent;
}

.search-input input::placeholder {
  color: #bbb;
}

.btn-parse {
  padding: 14px 32px;
  background: #ff2442;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}

.btn-parse:hover {
  background: #e6203c;
}

.btn-parse:active {
  transform: scale(0.98);
}

.btn-parse:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.btn-parse .spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255,255,255,0.3);
  border-radius: 50%;
  border-top-color: white;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.progress-bar {
  height: 3px;
  background: #eee;
  border-radius: 2px;
  margin-bottom: 24px;
  overflow: hidden;
  display: none;
}

.progress-bar.active {
  display: block;
}

.progress-fill {
  height: 100%;
  background: #ff2442;
  border-radius: 2px;
  transition: width 0.3s;
  width: 0%;
  animation: progress 2s ease-in-out infinite;
}

@keyframes progress {
  0% { width: 0%; margin-left: 0; }
  50% { width: 60%; margin-left: 20%; }
  100% { width: 0%; margin-left: 100%; }
}

.result-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.08);
  overflow: hidden;
  display: none;
  margin-bottom: 20px;
}

.result-card.active {
  display: block;
}

.video-container {
  position: relative;
  background: #000;
  border-radius: 12px 12px 0 0;
  overflow: hidden;
}

.video-container .dplayer {
  border-radius: 0;
}

.video-info {
  padding: 20px;
}

.video-title {
  font-size: 16px;
  font-weight: 500;
  color: #1a1a1a;
  margin-bottom: 12px;
  line-height: 1.5;
}

.video-meta {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.meta-author {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-name {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.meta-stats {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #999;
}

.meta-stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.meta-stats i {
  font-size: 16px;
}

.video-actions {
  display: flex;
  gap: 10px;
}

.btn-action {
  flex: 1;
  padding: 12px;
  border: 1px solid #eee;
  border-radius: 8px;
  background: white;
  font-size: 14px;
  color: #333;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  text-decoration: none;
}

.btn-action:hover {
  background: #f5f5f5;
  border-color: #ddd;
}

.btn-action.primary {
  background: #ff2442;
  color: white;
  border-color: #ff2442;
}

.btn-action.primary:hover {
  background: #e6203c;
}

.atlas-info {
  padding: 20px;
  border-bottom: 1px solid #f0f0f0;
}

.atlas-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2px;
  padding: 2px;
}

.atlas-item {
  position: relative;
  aspect-ratio: 1;
  overflow: hidden;
  cursor: pointer;
}

.atlas-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
  display: block;
}

.atlas-item:hover img {
  transform: scale(1.05);
}

.atlas-item .atlas-download {
  position: absolute;
  bottom: 8px;
  right: 8px;
  width: 32px;
  height: 32px;
  background: rgba(0,0,0,0.6);
  border: none;
  border-radius: 6px;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
  z-index: 10;
  text-decoration: none;
}

.atlas-item:hover .atlas-download {
  opacity: 1;
}

.atlas-actions {
  padding: 16px 20px;
  display: flex;
  gap: 10px;
}

.profile-header {
  padding: 24px 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.profile-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  object-fit: cover;
}

.profile-info h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 4px;
}

.profile-info p {
  font-size: 13px;
  color: #999;
}

.profile-stats {
  display: flex;
  gap: 24px;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  justify-content: center;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
}

.stat-label {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.video-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 2px;
  padding: 2px;
}

.video-item {
  position: relative;
  aspect-ratio: 9/16;
  overflow: hidden;
  cursor: pointer;
}

.video-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-item .overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 40px 10px 10px;
  background: linear-gradient(transparent, rgba(0,0,0,0.7));
  color: white;
}

.video-item .overlay .title {
  font-size: 12px;
  margin-bottom: 4px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.video-item .overlay .stats {
  font-size: 11px;
  opacity: 0.8;
  display: flex;
  gap: 8px;
}

.error-toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%) translateY(-100px);
  background: #ff4d4f;
  color: white;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  transition: transform 0.3s;
  z-index: 100;
}

.error-toast.show {
  transform: translateX(-50%) translateY(0);
}

.copy-toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%) translateY(-100px);
  background: #52c41a;
  color: white;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  transition: transform 0.3s;
  z-index: 100;
}

.copy-toast.show {
  transform: translateX(-50%) translateY(0);
}

.success-toast {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%) translateY(-100px);
  background: #52c41a;
  color: white;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  transition: transform 0.3s;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 8px;
}

.success-toast.show {
  transform: translateX(-50%) translateY(0);
}

.success-toast i {
  font-size: 18px;
}

.video-preview-modal {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.9);
  z-index: 1000;
  justify-content: center;
  align-items: center;
}

.video-preview-modal.active {
  display: flex;
}

.video-preview-content {
  width: 90%;
  max-width: 400px;
  position: relative;
}

.video-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.video-preview-header .title {
  color: white;
  font-size: 14px;
  flex: 1;
  margin-right: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-preview-close {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.video-preview-stats {
  color: #999;
  font-size: 12px;
  margin-top: 12px;
  text-align: center;
}

.image-preview-modal {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.95);
  z-index: 1000;
  justify-content: center;
  align-items: center;
}

.image-preview-modal.active {
  display: flex;
}

.image-preview-content {
  width: 90%;
  max-width: 500px;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.image-preview-close {
  position: absolute;
  top: -40px;
  right: 0;
  background: none;
  border: none;
  color: white;
  font-size: 24px;
  cursor: pointer;
  z-index: 1001;
}

.image-preview-counter {
  color: white;
  font-size: 14px;
  margin-bottom: 12px;
}

.image-preview-wrapper {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-preview-img {
  max-width: 100%;
  max-height: 60vh;
  object-fit: contain;
  border-radius: 8px;
}

.image-preview-prev,
.image-preview-next {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0,0,0,0.5);
  border: none;
  color: white;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.image-preview-prev {
  left: -20px;
}

.image-preview-next {
  right: -20px;
}

.image-preview-info {
  color: white;
  margin-top: 12px;
  font-size: 14px;
  text-align: center;
}

.image-preview-info .title {
  margin-bottom: 8px;
}

.image-preview-info .stats {
  color: #999;
  font-size: 12px;
}

.image-preview-actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
  width: 100%;
}

.image-preview-actions .btn-action {
  flex: 1;
}

@media (max-width: 600px) {
  .container {
    padding: 20px 16px;
  }

  .search-box {
    flex-direction: column;
  }

  .btn-parse {
    justify-content: center;
  }

  .atlas-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .video-list {
    grid-template-columns: repeat(2, 1fr);
  }

  .video-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
  }

  .video-actions {
    flex-direction: column;
  }

  .profile-header {
    flex-direction: column;
    text-align: center;
  }
}
</style>
