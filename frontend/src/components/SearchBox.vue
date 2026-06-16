<template>
  <div class="search-box">
    <div class="search-input">
      <i class="icon-link"></i>
      <input 
        type="text" 
        v-model="modelUrl" 
        placeholder="粘贴视频/图文/主页链接" 
        autocomplete="off"
        @keyup.enter="$emit('parse')"
      >
    </div>
    <button class="btn-parse" @click="$emit('parse')" :disabled="loading">
      <div v-if="loading" class="spinner"></div>
      <i v-else class="icon-search"></i>
      <span>{{ loading ? '解析中' : '解析' }}</span>
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  url: String,
  loading: Boolean
})

const emit = defineEmits(['update:url', 'parse'])

const modelUrl = computed({
  get: () => props.url,
  set: (val) => emit('update:url', val)
})
</script>
