<template>
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
        <div v-for="(img, index) in result.images" :key="index" class="atlas-item" @click="$emit('preview-image', index)">
          <img :src="getProxyImageUrl(img.url)" :alt="'图片 ' + (index + 1)" crossorigin="anonymous">
          <button class="atlas-download" @click.stop="$emit('download-single', img.url, index)">
            <i class="icon-download"></i>
          </button>
        </div>
      </div>
      <div class="atlas-actions">
        <button class="btn-action primary" @click="$emit('download-all', result.images)">
          <i class="icon-download"></i> 下载全部图片
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { getProxyImageUrl } from '../composables/url'

defineProps({ result: Object })
defineEmits(['preview-image', 'download-single', 'download-all'])

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) return (num / 10000).toFixed(1) + '万'
  return num.toString()
}
</script>
