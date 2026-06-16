<template>
  <div class="error-toast" :class="{ show: errorMsg }">
    {{ errorMsg }}
  </div>
  <div class="copy-toast" :class="{ show: copied }">已复制</div>
  <div class="success-toast" :class="{ show: successMsg }">
    <i class="icon-check-circle"></i>
    <span>{{ successMsg }}</span>
  </div>
</template>

<script setup>
defineProps({
  errorMsg: String,
  successMsg: String,
  copied: Boolean
})
</script>
