import { getProxyImageUrl, getImageExtension } from './url'

export function useDownloader(showError, showSuccess) {
  const downloadVideo = (videoUrl) => {
    const downloadUrl = `/api/download?url=${encodeURIComponent(videoUrl)}&filename=video.mp4`
    window.location.href = downloadUrl
  }

  const downloadSingleImage = (url, index) => {
    const ext = getImageExtension(url)
    const filename = `image_${index + 1}.${ext}`
    const downloadUrl = `/api/download?url=${encodeURIComponent(url)}&filename=${filename}`
    window.location.href = downloadUrl
  }

  const downloadAllImages = async (images) => {
    if (!images || images.length === 0) {
      showError('没有可下载的图片')
      return
    }
    
    showSuccess('正在打包图片...')
    
    try {
      const JSZip = (await import('jszip')).default
      const zip = new JSZip()
      const folder = zip.folder('images')
      
      const promises = images.map(async (image, index) => {
        const ext = getImageExtension(image.url)
        const downloadUrl = `/api/download?url=${encodeURIComponent(image.url)}&filename=image_${index + 1}.${ext}`
        const response = await fetch(downloadUrl)
        const blob = await response.blob()
        folder.file(`image_${index + 1}.${ext}`, blob)
      })
      
      await Promise.all(promises)
      
      const content = await zip.generateAsync({ type: 'blob' })
      
      const { saveAs } = await import('file-saver')
      saveAs(content, 'images.zip')
    } catch (error) {
      showError('打包下载失败: ' + error.message)
    }
  }

  return { downloadVideo, downloadSingleImage, downloadAllImages }
}
