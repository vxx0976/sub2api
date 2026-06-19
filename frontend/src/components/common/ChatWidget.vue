<script setup lang="ts">
import { ref, nextTick, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { useChatWebSocket } from '@/composables/useChatWebSocket'

const { t } = useI18n()
const chatStore = useChatStore()
const authStore = useAuthStore()
const messageContainer = ref<HTMLDivElement | null>(null)
const inputText = ref('')
const { connect, disconnect, isConnected } = useChatWebSocket()

function toggleChat() {
  if (chatStore.isOpen) {
    chatStore.closeChat()
    disconnect()
  } else {
    chatStore.openChat()
  }
}

async function handleSend() {
  const content = inputText.value.trim()
  if (!content || chatStore.sending) return
  inputText.value = ''
  await chatStore.send(content)
  scrollToBottom()
}

function scrollToBottom() {
  nextTick(() => {
    if (messageContainer.value) {
      messageContainer.value.scrollTop = messageContainer.value.scrollHeight
    }
  })
}

function connectWS() {
  if (!chatStore.conversation) return

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  const convId = chatStore.conversation.id

  let wsUrl: string
  if (authStore.isAuthenticated) {
    wsUrl = `${protocol}//${host}${baseUrl}/chat/ws?conversation_id=${convId}`
  } else {
    wsUrl = `${protocol}//${host}${baseUrl}/chat/ws?conversation_id=${convId}&guest_token=${chatStore.guestToken}`
  }

  connect(wsUrl, {
    onMessage: (data: any) => {
      if (data.type === 'new_message' && data.data?.message) {
        chatStore.receiveMessage(data.data.message)
        scrollToBottom()
      } else if (data.type === 'conversation_closed') {
        if (chatStore.conversation) {
          chatStore.conversation.status = 'closed'
        }
      }
    },
  })
}

watch(
  () => chatStore.conversation,
  (conv) => {
    if (conv && chatStore.isOpen) {
      connectWS()
      scrollToBottom()
    }
  }
)

watch(
  () => chatStore.messages.length,
  () => scrollToBottom()
)

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
  // 启动轮询兜底：面板关闭(WS 断开)时也能发现管理员的新回复并点亮红点
  chatStore.startPolling()
})

onUnmounted(() => {
  disconnect()
  chatStore.stopPolling()
})
</script>

<template>
  <Teleport to="body">
    <!-- Chat bubble button -->
    <button
      @click="toggleChat"
      class="fixed bottom-6 right-6 z-[110] flex h-14 w-14 items-center justify-center rounded-full bg-blue-600 text-white shadow-lg transition-all hover:bg-blue-700 hover:shadow-xl active:scale-95"
      :title="t('chat.title')"
    >
      <!-- Chat icon -->
      <svg v-if="!chatStore.isOpen" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
      </svg>
      <!-- Close icon -->
      <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
      </svg>
      <!-- Unread dot -->
      <span
        v-if="chatStore.hasNewMessage && !chatStore.isOpen"
        class="absolute -right-0.5 -top-0.5 flex h-4 w-4"
      >
        <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-400 opacity-75"></span>
        <span class="relative inline-flex h-4 w-4 rounded-full bg-red-500"></span>
      </span>
    </button>

    <!-- Chat panel -->
    <Transition name="chat-slide">
      <div
        v-if="chatStore.isOpen"
        class="fixed bottom-24 right-6 z-[110] flex w-[360px] flex-col overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-2xl dark:border-dark-600 dark:bg-dark-800 sm:w-[380px]"
        style="max-height: min(520px, calc(100vh - 140px))"
      >
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-gray-200 bg-blue-600 px-4 py-3 dark:border-dark-600">
          <div class="flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
            </svg>
            <span class="text-sm font-semibold text-white">{{ t('chat.title') }}</span>
          </div>
          <span v-if="isConnected" class="h-2 w-2 rounded-full bg-green-400" :title="t('chat.connected')"></span>
        </div>

        <!-- Messages area -->
        <div ref="messageContainer" class="flex-1 space-y-3 overflow-y-auto px-4 py-3" style="min-height: 200px">
          <!-- Welcome message -->
          <div v-if="!chatStore.welcomeShown && chatStore.messages.length === 0" class="rounded-xl bg-blue-50 p-3 text-sm text-blue-800 dark:bg-blue-900/20 dark:text-blue-300">
            <p class="font-medium">{{ t('chat.welcomeTitle') }}</p>
            <p class="mt-1 text-xs opacity-80">{{ t('chat.welcomeMessage') }}</p>
            <button
              @click="chatStore.markWelcomeShown()"
              class="mt-2 rounded-lg bg-blue-600 px-3 py-1 text-xs text-white hover:bg-blue-700"
            >
              {{ t('chat.gotIt') }}
            </button>
          </div>

          <!-- Loading -->
          <div v-if="chatStore.loading" class="flex justify-center py-8">
            <div class="h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent"></div>
          </div>

          <!-- Messages -->
          <template v-for="msg in chatStore.messages" :key="msg.id">
            <div :class="msg.sender_type === 'visitor' ? 'flex justify-end' : 'flex justify-start'">
              <div class="max-w-[80%]">
                <div
                  :class="[
                    'rounded-2xl px-3.5 py-2 text-sm leading-relaxed',
                    msg.sender_type === 'visitor'
                      ? 'rounded-br-md bg-blue-600 text-white'
                      : 'rounded-bl-md bg-gray-100 text-gray-900 dark:bg-dark-700 dark:text-gray-100',
                  ]"
                >
                  {{ msg.content }}
                </div>
                <div
                  :class="[
                    'mt-0.5 text-[10px] text-gray-400',
                    msg.sender_type === 'visitor' ? 'text-right' : 'text-left',
                  ]"
                >
                  {{ formatTime(msg.created_at) }}
                </div>
              </div>
            </div>
          </template>

          <!-- Empty state -->
          <div
            v-if="!chatStore.loading && chatStore.messages.length === 0 && chatStore.welcomeShown"
            class="flex flex-col items-center justify-center py-8 text-gray-400"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="mb-2 h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <span class="text-xs">{{ t('chat.emptyState') }}</span>
          </div>
        </div>

        <!-- Closed notice -->
        <div
          v-if="chatStore.conversation?.status === 'closed'"
          class="border-t border-gray-200 bg-gray-50 px-4 py-3 text-center text-xs text-gray-500 dark:border-dark-600 dark:bg-dark-900"
        >
          {{ t('chat.conversationClosed') }}
        </div>

        <!-- Input area -->
        <div
          v-else
          class="border-t border-gray-200 px-3 py-2.5 dark:border-dark-600"
        >
          <div class="flex items-end gap-2">
            <textarea
              v-model="inputText"
              @keydown.enter.exact.prevent="handleSend"
              :placeholder="t('chat.inputPlaceholder')"
              :disabled="chatStore.sending || chatStore.loading"
              rows="1"
              class="flex-1 resize-none rounded-xl border border-gray-300 bg-white px-3 py-2 text-sm outline-none transition focus:border-blue-500 focus:ring-1 focus:ring-blue-500 dark:border-dark-500 dark:bg-dark-700 dark:text-white"
              style="max-height: 80px"
            ></textarea>
            <button
              @click="handleSend"
              :disabled="!inputText.trim() || chatStore.sending"
              class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl bg-blue-600 text-white transition hover:bg-blue-700 disabled:opacity-50"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.chat-slide-enter-active {
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.chat-slide-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 1, 1);
}
.chat-slide-enter-from {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}
.chat-slide-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.98);
}
</style>
