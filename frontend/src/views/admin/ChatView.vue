<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import {
  listConversations,
  getMessages,
  sendReply,
  closeConversation,
  markRead,
} from '@/api/admin/chat'
import { useChatWebSocket } from '@/composables/useChatWebSocket'
import { useAdminChatStore } from '@/stores/adminChat'
import Select from '@/components/common/Select.vue'
import type { ChatConversation, ChatMessage } from '@/types'

const { t } = useI18n()

const statusOptions = computed(() => [
  { value: '', label: t('admin.chat.allStatus') },
  { value: 'open', label: t('admin.chat.statusOpen') },
  { value: 'closed', label: t('admin.chat.statusClosed') },
])

// 会话展示名：登录用户显示用户名/邮箱，访客显示「访客」
function convName(conv: ChatConversation): string {
  return conv.display_name || conv.visitor_name || t('admin.chat.visitorLabel')
}

// 头像渐变：按名称稳定取色，让会话列表更有辨识度
const AVATAR_GRADIENTS = [
  'from-blue-500 to-indigo-500',
  'from-emerald-500 to-teal-500',
  'from-amber-500 to-orange-500',
  'from-pink-500 to-rose-500',
  'from-violet-500 to-purple-500',
  'from-cyan-500 to-sky-500',
]
function avatarGradient(name: string): string {
  const s = name || '?'
  let h = 0
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0
  return AVATAR_GRADIENTS[h % AVATAR_GRADIENTS.length]
}

const conversations = ref<ChatConversation[]>([])
const selectedConv = ref<ChatConversation | null>(null)
const messages = ref<ChatMessage[]>([])
const loading = ref(false)
const messagesLoading = ref(false)
const replyText = ref('')
const sending = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const messageContainer = ref<HTMLDivElement | null>(null)

// 未读数以全局 store 为单一来源，悬浮提醒与本页徽标保持一致
const adminChat = useAdminChatStore()
const totalUnread = computed(() => adminChat.unreadCount)

const { connect, disconnect, isConnected } = useChatWebSocket()

const filteredConversations = computed(() => {
  let list = conversations.value
  if (statusFilter.value) {
    list = list.filter((c) => c.status === statusFilter.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(
      (c) =>
        c.display_name?.toLowerCase().includes(q) ||
        c.visitor_name?.toLowerCase().includes(q) ||
        c.last_message_preview?.toLowerCase().includes(q)
    )
  }
  // 最新消息排最前:后端已按 last_message_at 排过序,但 WebSocket 推来的新消息
  // 只就地改字段不会重排,这里兜住实时更新后的顺序
  return [...list].sort((a, b) => convTime(b) - convTime(a))
})

function convTime(conv: ChatConversation): number {
  const t = Date.parse(conv.last_message_at || conv.created_at || '')
  return Number.isNaN(t) ? 0 : t
}

async function loadConversations() {
  loading.value = true
  try {
    const data = await listConversations(1, 100)
    conversations.value = data.items || []
  } catch (e) {
    console.error('Failed to load conversations:', e)
  } finally {
    loading.value = false
  }
}

async function selectConversation(conv: ChatConversation) {
  selectedConv.value = conv
  messages.value = []
  messagesLoading.value = true
  const targetId = conv.id
  try {
    const data = await getMessages(conv.id)
    if (selectedConv.value?.id !== targetId) return
    messages.value = data.messages || []
    scrollToBottom()

    if (conv.admin_unread_count > 0) {
      await markRead(conv.id)
      conv.admin_unread_count = 0
      adminChat.fetchUnreadCount()
    }
  } catch (e) {
    console.error('Failed to load messages:', e)
  } finally {
    if (selectedConv.value?.id === targetId) {
      messagesLoading.value = false
    }
  }
}

async function handleSendReply() {
  if (!selectedConv.value || !replyText.value.trim() || sending.value) return
  sending.value = true
  try {
    const msg = await sendReply(selectedConv.value.id, replyText.value.trim())
    messages.value.push(msg)
    replyText.value = ''
    scrollToBottom()

    selectedConv.value.last_message_preview = msg.content.substring(0, 100)
    selectedConv.value.last_message_at = msg.created_at
  } catch (e) {
    console.error('Failed to send reply:', e)
  } finally {
    sending.value = false
  }
}

async function handleClose() {
  if (!selectedConv.value) return
  if (!confirm(t('admin.chat.closeConfirm'))) return
  try {
    await closeConversation(selectedConv.value.id)
    selectedConv.value.status = 'closed'
  } catch (e) {
    console.error('Failed to close conversation:', e)
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight
    }
  })
}

function formatTime(dateStr: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' }) +
    ' ' +
    date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function connectAdminWS() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  const wsUrl = `${protocol}//${host}${baseUrl}/admin/chat/ws`

  connect(wsUrl, {
    // 函数形式:每次自动重连都现取 localStorage,避免用连接时冻结的过期 token
    protocols: () => {
      const token = localStorage.getItem('auth_token') || ''
      return ['sub2api-admin', `jwt.${token}`]
    },
    onMessage: (data: any) => {
      if (data.type === 'new_message' && data.data?.message) {
        const msg = data.data.message as ChatMessage
        if (selectedConv.value && msg.conversation_id === selectedConv.value.id) {
          const exists = messages.value.some((m) => m.id === msg.id)
          if (!exists) {
            messages.value.push(msg)
            scrollToBottom()
            markRead(selectedConv.value.id)
          }
        }
        const conv = conversations.value.find((c) => c.id === msg.conversation_id)
        if (conv) {
          conv.last_message_preview = msg.content.substring(0, 100)
          conv.last_message_at = msg.created_at
          if (!selectedConv.value || selectedConv.value.id !== msg.conversation_id) {
            conv.admin_unread_count = (conv.admin_unread_count || 0) + 1
          }
        } else {
          loadConversations()
        }
        adminChat.fetchUnreadCount()
      }
    },
  })
}

onMounted(() => {
  loadConversations()
  adminChat.fetchUnreadCount()
  connectAdminWS()
})

onUnmounted(() => {
  disconnect()
})
</script>

<template>
  <AppLayout>
    <div class="-m-4 flex h-[calc(100dvh-4rem)] overflow-hidden rounded-none bg-white md:-m-6 lg:-m-8 dark:bg-dark-900">
      <!-- Left: Conversation list -->
      <div class="flex w-80 flex-shrink-0 flex-col border-r border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <!-- Header -->
        <div class="flex-shrink-0 border-b border-gray-200 px-4 py-3 dark:border-dark-700">
          <div class="flex items-center justify-between">
            <h2 class="flex items-center gap-2 text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.chat.title') }}
              <span
                v-if="totalUnread > 0"
                class="inline-flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-500 px-1.5 text-[11px] font-semibold text-white"
              >
                {{ totalUnread > 99 ? '99+' : totalUnread }}
              </span>
            </h2>
            <span
              class="flex items-center gap-1.5 text-[11px] font-medium"
              :class="isConnected ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'"
              :title="isConnected ? t('admin.chat.online') : t('admin.chat.offline')"
            >
              <span class="relative flex h-2 w-2">
                <span
                  v-if="isConnected"
                  class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
                ></span>
                <span
                  class="relative inline-flex h-2 w-2 rounded-full"
                  :class="isConnected ? 'bg-emerald-500' : 'bg-gray-400'"
                ></span>
              </span>
              {{ isConnected ? t('admin.chat.online') : t('admin.chat.offline') }}
            </span>
          </div>
          <!-- Search + Filter -->
          <div class="mt-3 flex gap-2">
            <div class="relative flex-1">
              <svg
                class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35M17 11a6 6 0 11-12 0 6 6 0 0112 0z" />
              </svg>
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.chat.searchPlaceholder')"
                class="input w-full pl-8 text-sm"
              />
            </div>
            <Select v-model="statusFilter" :options="statusOptions" class="w-28 flex-shrink-0" />
          </div>
        </div>

        <!-- Conversation list -->
        <div class="flex-1 overflow-y-auto">
          <div v-if="loading" class="flex justify-center py-10">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent"></div>
          </div>
          <div v-else-if="filteredConversations.length === 0" class="flex flex-col items-center justify-center gap-2 px-4 py-12 text-center text-gray-400">
            <svg class="h-9 w-9 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <span class="text-sm">{{ t('admin.chat.noConversations') }}</span>
          </div>
          <div
            v-for="conv in filteredConversations"
            :key="conv.id"
            @click="selectConversation(conv)"
            :class="[
              'relative flex w-full cursor-pointer items-center gap-2.5 border-b border-gray-100 px-3 py-2.5 text-left transition-colors hover:bg-gray-50 dark:border-dark-700/60 dark:hover:bg-dark-700/50',
              selectedConv?.id === conv.id ? 'bg-blue-50/70 dark:bg-blue-900/20' : '',
            ]"
          >
            <span
              v-if="selectedConv?.id === conv.id"
              class="absolute inset-y-0 left-0 w-1 rounded-r bg-gradient-to-b from-blue-500 to-indigo-500"
            ></span>
            <!-- Avatar -->
            <div class="relative flex-shrink-0">
              <div
                class="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br text-sm font-semibold uppercase text-white shadow-sm"
                :class="avatarGradient(convName(conv))"
              >
                {{ (convName(conv) || '?')[0] }}
              </div>
              <span
                class="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-white dark:border-dark-800"
                :class="conv.status === 'open' ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-dark-500'"
              ></span>
            </div>
            <!-- Content -->
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <span
                  class="truncate text-sm font-medium text-gray-900 dark:text-white"
                  :class="{ 'font-semibold': conv.admin_unread_count > 0 }"
                >
                  {{ convName(conv) }}
                </span>
                <span class="flex-shrink-0 text-[10px] text-gray-400">
                  {{ conv.last_message_at ? formatTime(conv.last_message_at) : '' }}
                </span>
              </div>
              <div class="mt-0.5 flex items-center justify-between gap-2">
                <p
                  class="truncate text-xs"
                  :class="conv.admin_unread_count > 0 ? 'font-medium text-gray-700 dark:text-gray-200' : 'text-gray-500 dark:text-gray-400'"
                >
                  {{ conv.last_message_preview || '...' }}
                </p>
                <span
                  v-if="conv.admin_unread_count > 0"
                  class="flex h-[18px] min-w-[18px] flex-shrink-0 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-semibold text-white"
                >
                  {{ conv.admin_unread_count > 99 ? '99+' : conv.admin_unread_count }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Message area -->
      <div class="flex flex-1 flex-col bg-gray-50 dark:bg-dark-900">
        <template v-if="selectedConv">
          <!-- Chat header -->
          <div class="flex flex-shrink-0 items-center justify-between border-b border-gray-200 bg-white px-5 py-3 dark:border-dark-700 dark:bg-dark-800">
            <div class="flex min-w-0 items-center gap-3">
              <div
                class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-gradient-to-br text-sm font-semibold uppercase text-white shadow-sm"
                :class="avatarGradient(convName(selectedConv))"
              >
                {{ (convName(selectedConv) || '?')[0] }}
              </div>
              <div class="min-w-0">
                <h3 class="truncate text-sm font-semibold text-gray-900 dark:text-white">
                  {{ convName(selectedConv) }}
                </h3>
                <span
                  class="inline-flex items-center gap-1 text-xs"
                  :class="selectedConv.status === 'open' ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'"
                >
                  <span
                    class="h-1.5 w-1.5 rounded-full"
                    :class="selectedConv.status === 'open' ? 'bg-emerald-500' : 'bg-gray-400'"
                  ></span>
                  {{ selectedConv.status === 'open' ? t('admin.chat.statusOpen') : t('admin.chat.statusClosed') }}
                </span>
              </div>
            </div>
            <div class="flex flex-shrink-0 gap-2">
              <button
                v-if="selectedConv.status === 'open'"
                @click="handleClose"
                class="btn btn-secondary text-xs"
              >
                {{ t('admin.chat.close') }}
              </button>
              <button
                @click="loadConversations"
                class="btn btn-secondary !px-2 text-xs"
                :title="t('common.refresh')"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Messages -->
          <div ref="messageContainer" class="flex-1 space-y-4 overflow-y-auto px-6 py-5">
            <div v-if="messagesLoading" class="flex justify-center py-8">
              <div class="h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent"></div>
            </div>
            <div v-else-if="messages.length === 0" class="flex flex-col items-center justify-center py-16 text-gray-400">
              <svg xmlns="http://www.w3.org/2000/svg" class="mb-2 h-10 w-10 opacity-60" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <span class="text-sm">{{ t('admin.chat.noMessages') }}</span>
            </div>
            <div
              v-for="msg in messages"
              :key="msg.id"
              :class="msg.sender_type === 'admin' ? 'flex justify-end' : 'flex justify-start'"
            >
              <div class="flex max-w-[72%] flex-col" :class="msg.sender_type === 'admin' ? 'items-end' : 'items-start'">
                <span class="mb-1 px-1 text-[10px] text-gray-400">
                  {{ msg.sender_type === 'admin' ? t('admin.chat.adminLabel') : t('admin.chat.visitorLabel') }}
                </span>
                <div
                  :class="[
                    'whitespace-pre-wrap break-words px-4 py-2.5 text-sm leading-relaxed shadow-sm',
                    msg.sender_type === 'admin'
                      ? 'rounded-2xl rounded-br-md bg-gradient-to-br from-blue-600 to-indigo-600 text-white'
                      : 'rounded-2xl rounded-bl-md bg-white text-gray-900 ring-1 ring-gray-100 dark:bg-dark-700 dark:text-gray-100 dark:ring-dark-600',
                  ]"
                >
                  {{ msg.content }}
                </div>
                <span class="mt-1 px-1 text-[10px] text-gray-400">
                  {{ formatTime(msg.created_at) }}
                </span>
              </div>
            </div>
          </div>

          <!-- Reply input -->
          <div v-if="selectedConv.status === 'open'" class="flex-shrink-0 border-t border-gray-200 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-800">
            <div class="flex items-end gap-3">
              <textarea
                v-model="replyText"
                @keydown.enter.exact.prevent="handleSendReply"
                :placeholder="t('admin.chat.replyPlaceholder')"
                :disabled="sending"
                rows="1"
                class="input min-h-[42px] flex-1 resize-none text-sm"
                style="max-height: 120px"
              ></textarea>
              <button
                @click="handleSendReply"
                :disabled="!replyText.trim() || sending"
                class="btn btn-primary flex flex-shrink-0 items-center gap-1.5"
              >
                <svg v-if="!sending" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                </svg>
                <span v-if="sending" class="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"></span>
                {{ t('admin.chat.send') }}
              </button>
            </div>
            <p class="mt-1.5 px-1 text-[10px] text-gray-400">{{ t('admin.chat.enterToSend') }}</p>
          </div>
          <div v-else class="flex-shrink-0 border-t border-gray-200 bg-gray-50 px-4 py-3 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900">
            {{ t('admin.chat.closedNotice') }}
          </div>
        </template>

        <!-- No conversation selected -->
        <div v-else class="flex flex-1 flex-col items-center justify-center px-6 text-center text-gray-400">
          <div class="mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-dark-800 dark:to-dark-700">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-blue-400/70" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.3">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <span class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('admin.chat.selectConversation') }}</span>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
