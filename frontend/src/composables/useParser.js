import { ref } from 'vue'

export function useParser(showError, showSuccess) {
  const url = ref('')
  const loading = ref(false)
  const result = ref(null)

  const parse = async () => {
    if (!url.value.trim()) {
      showError('请输入链接')
      return
    }

    loading.value = true
    result.value = null

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
        return data
      } else {
        showError(data.error || '解析失败')
        return null
      }
    } catch (error) {
      showError('请求失败: ' + error.message)
      return null
    } finally {
      loading.value = false
    }
  }

  return { url, loading, result, parse }
}
