<template>
  <div class="video-preview-modal" :class="{ active: visible }" @click.self="$emit('close')">
    <div class="video-preview-content">
      <div class="video-preview-header">
        <div class="title">{{ title }}</div>
        <button class="video-preview-close" @click="$emit('close')">
          <i class="icon-x"></i>
        </button>
      </div>
      <div ref="playerRef"></div>
      <div class="video-preview-stats">{{ stats }}</div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  visible: Boolean,
  title: String,
  stats: String,
  videoUrl: String,
  coverUrl: String
})

const emit = defineEmits(['close', 'init-player'])
const playerRef = ref(null)

watch(() => props.visible, (val) => {
  if (val && props.videoUrl) {
    setTimeout(() => emit('init-player', playerRef.value, props.videoUrl, props.coverUrl), 100)
  }
})
</script>
