import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUnreadCount } from '@/api/admin/chat'

/**
 * 管理员客服未读消息全局状态。
 *
 * 通过轮询 /admin/chat/unread-count 驱动「任意页面可见」的全局红点提醒，
 * 与 ChatView 自身的 admin WebSocket 解耦——避免同一管理员开两条 WS 连接
 * 互相顶号。页面隐藏(切到后台标签)时跳过轮询以省资源，重新可见时立即刷新。
 */
export const useAdminChatStore = defineStore('adminChat', () => {
  const unreadCount = ref(0)
  let timer: ReturnType<typeof setInterval> | null = null

  async function fetchUnreadCount() {
    try {
      const data = await getUnreadCount()
      unreadCount.value = data?.count ?? 0
    } catch {
      // 静默失败，下次轮询自动重试
    }
  }

  function onVisible() {
    if (document.visibilityState === 'visible') fetchUnreadCount()
  }

  function startPolling(intervalMs = 10000) {
    if (timer) return
    fetchUnreadCount()
    timer = setInterval(() => {
      if (document.visibilityState === 'visible') fetchUnreadCount()
    }, intervalMs)
    document.addEventListener('visibilitychange', onVisible)
  }

  function stopPolling() {
    if (timer) {
      clearInterval(timer)
      timer = null
    }
    document.removeEventListener('visibilitychange', onVisible)
  }

  function reset() {
    stopPolling()
    unreadCount.value = 0
  }

  return { unreadCount, fetchUnreadCount, startPolling, stopPolling, reset }
})
