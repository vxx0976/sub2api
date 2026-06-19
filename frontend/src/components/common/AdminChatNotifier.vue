<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAdminChatStore } from '@/stores/adminChat'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const adminChat = useAdminChatStore()

const count = computed(() => adminChat.unreadCount)
const badge = computed(() => (count.value > 99 ? '99+' : String(count.value)))
// 已经在客服页时不再悬浮提醒，避免遮挡操作
const onChatPage = computed(() => route.path === '/admin/chat')

function goChat() {
  router.push('/admin/chat')
}

onMounted(() => adminChat.startPolling())
onUnmounted(() => adminChat.stopPolling())
</script>

<template>
  <Teleport to="body">
    <Transition name="notifier-pop">
      <button
        v-if="!onChatPage"
        type="button"
        @click="goChat"
        :title="t('admin.chat.notifierTitle')"
        :aria-label="t('admin.chat.notifierTitle')"
        class="group fixed bottom-6 right-6 z-[110] flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-indigo-600 text-white shadow-lg shadow-blue-500/30 ring-1 ring-white/10 transition-all duration-200 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/40 active:scale-95"
      >
        <!-- 客服气泡图标 -->
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-6 w-6 transition-transform duration-200 group-hover:-rotate-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.8"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
          />
        </svg>

        <!-- 未读数红点 -->
        <span
          v-if="count > 0"
          class="absolute -right-1 -top-1 flex h-5 min-w-[20px] items-center justify-center"
        >
          <span
            class="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-400 opacity-75"
          ></span>
          <span
            class="relative inline-flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-500 px-1.5 text-[11px] font-semibold leading-none text-white ring-2 ring-white dark:ring-dark-900"
          >
            {{ badge }}
          </span>
        </span>
      </button>
    </Transition>
  </Teleport>
</template>

<style scoped>
.notifier-pop-enter-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.notifier-pop-leave-active {
  transition: all 0.18s cubic-bezier(0.4, 0, 1, 1);
}
.notifier-pop-enter-from,
.notifier-pop-leave-to {
  opacity: 0;
  transform: scale(0.6) translateY(8px);
}
</style>
