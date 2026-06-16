export const getProxyImageUrl = (url) => {
  if (!url) return ''
  if (url.includes('xhscdn.com') || url.includes('xiaohongshu.com') || url.includes('sns-') ||
      url.includes('douyin.com') || url.includes('kuaishou.com') || url.includes('gifshow.com')) {
    return `/api/proxy-image?url=${encodeURIComponent(url)}`
  }
  return url
}

export const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

export const getImageExtension = (url) => {
  return (url.match(/\.(jpg|jpeg|png|webp|heic|avif)(\?|$)/i) || [null, 'jpg'])[1] || 'jpg'
}
