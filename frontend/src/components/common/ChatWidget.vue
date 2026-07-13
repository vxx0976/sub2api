<script setup lang="ts">
import { ref, computed, nextTick, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useChatStore } from '@/stores/chat'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores'
import { useChatWebSocket } from '@/composables/useChatWebSocket'

const { t } = useI18n()
const chatStore = useChatStore()
const authStore = useAuthStore()
const appStore = useAppStore()

// 联系方式：与 /console-home 同源(公共设置),客服离线时在对话框内提示
const contactWechat = computed(() => appStore.cachedPublicSettings?.contact_wechat || '')
const contactTelegram = computed(() => appStore.cachedPublicSettings?.contact_telegram || '')
const contactQQ = computed(() => appStore.cachedPublicSettings?.contact_qq || '')
const contactEmail = computed(() =>
  appStore.isResellerDomain || authStore.isResellerUser ? '' : 'vanxuehan@gmail.com'
)
const hasContactInfo = computed(
  () => !!contactWechat.value || !!contactTelegram.value || !!contactQQ.value || !!contactEmail.value
)
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
  let protocols: (() => string[] | undefined) | undefined
  if (authStore.isAuthenticated) {
    wsUrl = `${protocol}//${host}${baseUrl}/chat/ws?conversation_id=${convId}`
    protocols = () => {
      const token = localStorage.getItem('auth_token')
      return token ? ['sub2api-chat', `jwt.${token}`] : undefined
    }
  } else {
    wsUrl = `${protocol}//${host}${baseUrl}/chat/ws?conversation_id=${convId}&guest_token=${chatStore.guestToken}`
  }

  connect(wsUrl, {
    protocols,
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
  () => [chatStore.isOpen, chatStore.conversation?.id] as const,
  ([isOpen, conversationId]) => {
    if (isOpen && conversationId) {
      connectWS()
      scrollToBottom()
    }
  }
)

watch(
  () => [authStore.isAuthenticated, authStore.user?.id ?? null] as const,
  ([isAuthenticated, userId], [wasAuthenticated, previousUserId]) => {
    if (isAuthenticated === wasAuthenticated && userId === previousUserId) return
    disconnect()
    chatStore.reset()
    // SPA 内登录/登出不刷新页面,重置后立即探查新身份的既有会话,恢复未读红点
    void chatStore.restoreConversation()
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

onMounted(async () => {
  // 先探查既有会话(刷新后内存为空时也能恢复 convId 并据此点亮红点),再启动轮询兜底
  await chatStore.restoreConversation()
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

          <!-- Offline notice: 客服不在线时提示联系方式（与 /console-home 同源） -->
          <div
            v-if="!chatStore.loading && !chatStore.adminOnline"
            class="rounded-xl bg-amber-50 p-3 dark:bg-amber-900/15"
          >
            <div class="flex items-center gap-1.5 text-sm font-medium text-amber-800 dark:text-amber-300">
              <svg class="h-4 w-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              {{ t('chat.offlineTitle') }}
            </div>
            <template v-if="hasContactInfo">
              <p class="mt-1 text-xs text-amber-700/80 dark:text-amber-300/70">
                {{ t('chat.offlineContactHint') }}
              </p>
              <div class="mt-2 space-y-1.5">
                <!-- WeChat -->
                <div v-if="contactWechat" class="flex items-center gap-2">
                  <svg class="h-4 w-4 flex-shrink-0 text-green-500" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 01.213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 00.167-.054l1.903-1.114a.864.864 0 01.717-.098 10.16 10.16 0 002.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178A1.17 1.17 0 014.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178 1.17 1.17 0 01-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 01.598.082l1.584.926a.272.272 0 00.14.045c.134 0 .24-.109.24-.245 0-.06-.024-.12-.04-.177l-.325-1.233a.493.493 0 01.177-.554C23.04 18.423 24 16.837 24 15.069c0-3.07-3.022-5.997-7.062-6.21zM13.544 12.5c.535 0 .969.44.969.982a.976.976 0 01-.969.983.976.976 0 01-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 01-.969.983.976.976 0 01-.969-.983c0-.542.434-.982.969-.982z"/>
                  </svg>
                  <span class="text-xs font-medium text-gray-700 dark:text-gray-200">{{ t('consoleHome.contact.wechat') }}: {{ contactWechat }}</span>
                </div>
                <!-- Telegram -->
                <a
                  v-if="contactTelegram"
                  :href="`https://t.me/${contactTelegram}`"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="flex items-center gap-2 hover:underline"
                >
                  <svg class="h-4 w-4 flex-shrink-0 text-blue-500" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M11.944 0A12 12 0 000 12a12 12 0 0012 12 12 12 0 0012-12A12 12 0 0012 0a12 12 0 00-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 01.171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.479.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
                  </svg>
                  <span class="text-xs font-medium text-gray-700 dark:text-gray-200">{{ t('consoleHome.contact.telegram') }}: @{{ contactTelegram }}</span>
                </a>
                <!-- QQ -->
                <div v-if="contactQQ" class="flex items-center gap-2">
                  <svg class="h-4 w-4 flex-shrink-0 text-sky-500" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12.003 2C6.004 2 3 6.086 3 9.166c0 3.313 1.727 6.286 2.907 7.594-.09.86-.455 2.11-.809 3.063-.18.484.235.696.56.512 1.108-.628 2.613-1.62 3.31-2.12.98.254 1.965.384 3.035.384 6 0 9.003-4.086 9.003-7.166C21.006 6.086 18.003 2 12.003 2z"/>
                  </svg>
                  <span class="text-xs font-medium text-gray-700 dark:text-gray-200">{{ t('consoleHome.contact.qq') }}: {{ contactQQ }}</span>
                </div>
                <!-- Email -->
                <a
                  v-if="contactEmail"
                  :href="`mailto:${contactEmail}`"
                  class="flex items-center gap-2 hover:underline"
                >
                  <svg class="h-4 w-4 flex-shrink-0 text-red-500" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M1.5 8.67v8.58a3 3 0 003 3h15a3 3 0 003-3V8.67l-8.928 5.493a3 3 0 01-3.144 0L1.5 8.67z" />
                    <path d="M22.5 6.908V6.75a3 3 0 00-3-3h-15a3 3 0 00-3 3v.158l9.714 5.978a1.5 1.5 0 001.572 0L22.5 6.908z" />
                  </svg>
                  <span class="text-xs font-medium text-gray-700 dark:text-gray-200">{{ t('consoleHome.contact.email') }}: {{ contactEmail }}</span>
                </a>
              </div>
            </template>
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
