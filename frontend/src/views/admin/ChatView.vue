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
  getUnreadCount,
} from '@/api/admin/chat'
import { useChatWebSocket } from '@/composables/useChatWebSocket'
import type { ChatConversation, ChatMessage } from '@/types'

const { t } = useI18n()

const conversations = ref<ChatConversation[]>([])
const selectedConv = ref<ChatConversation | null>(null)
const messages = ref<ChatMessage[]>([])
const loading = ref(false)
const messagesLoading = ref(false)
const replyText = ref('')
const sending = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const totalUnread = ref(0)
const messageContainer = ref<HTMLDivElement | null>(null)

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
  return list
})

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

async function loadUnreadCount() {
  try {
    const data = await getUnreadCount()
    totalUnread.value = data.count
  } catch {
    // ignore
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
      loadUnreadCount()
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
  const token = localStorage.getItem('auth_token') || ''
  const wsUrl = `${protocol}//${host}${baseUrl}/admin/chat/ws`

  connect(wsUrl, {
    protocols: ['sub2api-admin', `jwt.${token}`],
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
        loadUnreadCount()
      }
    },
  })
}

onMounted(() => {
  loadConversations()
  loadUnreadCount()
  connectAdminWS()
})

onUnmounted(() => {
  disconnect()
})
</script>

<template>
  <AppLayout>
    <div class="flex h-[calc(100vh-64px)] overflow-hidden">
      <!-- Left: Conversation list -->
      <div class="flex w-80 flex-shrink-0 flex-col border-r border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
        <!-- Header -->
        <div class="border-b border-gray-200 px-4 py-3 dark:border-dark-600">
          <div class="flex items-center justify-between">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.chat.title') }}
            </h2>
            <div class="flex items-center gap-2">
              <span v-if="isConnected" class="h-2 w-2 rounded-full bg-green-500"></span>
              <span v-if="totalUnread > 0" class="rounded-full bg-red-500 px-2 py-0.5 text-xs text-white">
                {{ totalUnread }}
              </span>
            </div>
          </div>
          <!-- Search + Filter -->
          <div class="mt-2 flex gap-2">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.chat.searchPlaceholder')"
              class="input flex-1 text-sm"
            />
            <select v-model="statusFilter" class="input w-24 text-sm">
              <option value="">{{ t('admin.chat.allStatus') }}</option>
              <option value="open">{{ t('admin.chat.statusOpen') }}</option>
              <option value="closed">{{ t('admin.chat.statusClosed') }}</option>
            </select>
          </div>
        </div>

        <!-- Conversation list -->
        <div class="flex-1 overflow-y-auto">
          <div v-if="loading" class="flex justify-center py-8">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent"></div>
          </div>
          <div v-else-if="filteredConversations.length === 0" class="px-4 py-8 text-center text-sm text-gray-400">
            {{ t('admin.chat.noConversations') }}
          </div>
          <div
            v-for="conv in filteredConversations"
            :key="conv.id"
            @click="selectConversation(conv)"
            :class="[
              'cursor-pointer border-b border-gray-100 px-3 py-2 transition hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-700',
              selectedConv?.id === conv.id ? 'bg-blue-50 dark:bg-blue-900/20' : '',
            ]"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-gray-200 text-xs font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300">
                  {{ (conv.display_name || conv.visitor_name || '?')[0] }}
                </div>
                <div class="min-w-0">
                  <div class="flex items-center gap-1.5">
                    <span class="truncate text-sm font-medium text-gray-900 dark:text-white">
                      {{ conv.display_name || conv.visitor_name || t('admin.chat.visitorLabel') }}
                    </span>
                    <span
                      :class="[
                        'inline-flex h-1.5 w-1.5 rounded-full',
                        conv.status === 'open' ? 'bg-green-500' : 'bg-gray-400',
                      ]"
                    ></span>
                  </div>
                  <p class="truncate text-xs text-gray-500 dark:text-gray-400">
                    {{ conv.last_message_preview || '...' }}
                  </p>
                </div>
              </div>
              <div class="flex flex-col items-end gap-1">
                <span class="text-[10px] text-gray-400">
                  {{ conv.last_message_at ? formatTime(conv.last_message_at) : '' }}
                </span>
                <span
                  v-if="conv.admin_unread_count > 0"
                  class="flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-500 px-1.5 text-[10px] text-white"
                >
                  {{ conv.admin_unread_count }}
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
          <div class="flex items-center justify-between border-b border-gray-200 bg-white px-4 py-3 dark:border-dark-600 dark:bg-dark-800">
            <div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ selectedConv.display_name || selectedConv.visitor_name || t('admin.chat.visitorLabel') }}
              </h3>
              <span
                :class="[
                  'text-xs',
                  selectedConv.status === 'open' ? 'text-green-600' : 'text-gray-400',
                ]"
              >
                {{ selectedConv.status === 'open' ? t('admin.chat.statusOpen') : t('admin.chat.statusClosed') }}
              </span>
            </div>
            <div class="flex gap-2">
              <button
                v-if="selectedConv.status === 'open'"
                @click="handleClose"
                class="btn btn-secondary text-xs"
              >
                {{ t('admin.chat.close') }}
              </button>
              <button
                @click="loadConversations"
                class="btn btn-secondary text-xs"
                :title="t('common.refresh')"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Messages -->
          <div ref="messageContainer" class="flex-1 space-y-3 overflow-y-auto px-6 py-4">
            <div v-if="messagesLoading" class="flex justify-center py-8">
              <div class="h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent"></div>
            </div>
            <div v-else-if="messages.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
              <svg xmlns="http://www.w3.org/2000/svg" class="mb-2 h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
              <span class="text-sm">{{ t('admin.chat.noMessages') }}</span>
            </div>
            <template v-for="msg in messages" :key="msg.id">
              <div :class="msg.sender_type === 'admin' ? 'flex justify-end' : 'flex justify-start'">
                <div class="max-w-[65%]">
                  <div class="mb-0.5 text-[10px] text-gray-400" :class="msg.sender_type === 'admin' ? 'text-right' : 'text-left'">
                    {{ msg.sender_type === 'admin' ? t('admin.chat.adminLabel') : t('admin.chat.visitorLabel') }}
                  </div>
                  <div
                    :class="[
                      'rounded-2xl px-4 py-2 text-sm leading-relaxed',
                      msg.sender_type === 'admin'
                        ? 'rounded-br-md bg-blue-600 text-white'
                        : 'rounded-bl-md bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-gray-100',
                    ]"
                  >
                    {{ msg.content }}
                  </div>
                  <div class="mt-0.5 text-[10px] text-gray-400" :class="msg.sender_type === 'admin' ? 'text-right' : 'text-left'">
                    {{ formatTime(msg.created_at) }}
                  </div>
                </div>
              </div>
            </template>
          </div>

          <!-- Reply input -->
          <div v-if="selectedConv.status === 'open'" class="border-t border-gray-200 bg-white px-4 py-3 dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-end gap-3">
              <textarea
                v-model="replyText"
                @keydown.enter.exact.prevent="handleSendReply"
                :placeholder="t('admin.chat.replyPlaceholder')"
                :disabled="sending"
                rows="2"
                class="input flex-1 resize-none text-sm"
                style="max-height: 120px"
              ></textarea>
              <button
                @click="handleSendReply"
                :disabled="!replyText.trim() || sending"
                class="btn btn-primary flex-shrink-0"
              >
                {{ t('admin.chat.send') }}
              </button>
            </div>
          </div>
          <div v-else class="border-t border-gray-200 bg-gray-50 px-4 py-3 text-center text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-900">
            {{ t('admin.chat.closedNotice') }}
          </div>
        </template>

        <!-- No conversation selected -->
        <div v-else class="flex flex-1 flex-col items-center justify-center text-gray-400">
          <svg xmlns="http://www.w3.org/2000/svg" class="mb-3 h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
          <span class="text-sm">{{ t('admin.chat.selectConversation') }}</span>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
