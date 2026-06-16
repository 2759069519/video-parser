<template>
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
        <div v-for="photo in result.photos" :key="photo.photo_id" class="video-item" @click="$emit('preview', photo)">
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
</template>

<script setup>
defineProps({ result: Object })
defineEmits(['preview'])

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) return (num / 10000).toFixed(1) + '万'
  return num.toString()
}
</script>
