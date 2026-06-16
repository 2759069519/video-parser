import { ref } from 'vue'

export function useToast() {
  const errorMsg = ref('')
  const successMsg = ref('')
  const copied = ref(false)

  const showError = (msg) => {
    errorMsg.value = msg
    setTimeout(() => { errorMsg.value = '' }, 3000)
  }

  const showSuccess = (msg) => {
    successMsg.value = msg
    setTimeout(() => { successMsg.value = '' }, 2000)
  }

  const showCopied = () => {
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  }

  const copyLink = (link) => {
    navigator.clipboard.writeText(link).then(() => {
      showCopied()
    }).catch(() => {
      showError('复制失败')
    })
  }

  return { errorMsg, successMsg, copied, showError, showSuccess, copyLink }
}
