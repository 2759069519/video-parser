<template>
  <div class="image-preview-modal" :class="{ active: visible }" @click.self="$emit('close')">
    <div class="image-preview-content">
      <button class="image-preview-close" @click="$emit('close')">
        <i class="icon-x"></i>
      </button>
      <div class="image-preview-counter">{{ index + 1 }} / {{ images.length }}</div>
      <div class="image-preview-wrapper">
        <button class="image-preview-prev" @click="$emit('prev')" v-if="images.length > 1">
          <i class="icon-chevron-left"></i>
        </button>
        <img :src="getProxyImageUrl(images[index]?.url)" :alt="'图片 ' + (index + 1)" class="image-preview-img" crossorigin="anonymous">
        <button class="image-preview-next" @click="$emit('next')" v-if="images.length > 1">
          <i class="icon-chevron-right"></i>
        </button>
      </div>
      <div class="image-preview-info">
        <div class="title">{{ title }}</div>
        <div class="stats">{{ stats }}</div>
      </div>
      <div class="image-preview-actions">
        <button class="btn-action" @click="$emit('download-single', images[index]?.url, index)">
          <i class="icon-download"></i> 下载当前图片
        </button>
        <button class="btn-action primary" @click="$emit('download-all', images)">
          <i class="icon-download"></i> 下载全部
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { getProxyImageUrl } from '../composables/url'

defineProps({
  visible: Boolean,
  images: Array,
  index: Number,
  title: String,
  stats: String
})

defineEmits(['close', 'prev', 'next', 'download-single', 'download-all'])
</script>
