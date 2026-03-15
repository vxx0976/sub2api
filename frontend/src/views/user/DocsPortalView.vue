<template>
  <PublicLayout>
    <div class="mx-auto max-w-5xl px-4 py-12">
      <!-- Header -->
      <div class="mb-12 text-center">
        <div class="mb-4 inline-flex items-center gap-2 rounded-full border border-primary-200 bg-primary-50 px-4 py-1.5 text-sm text-primary-600 dark:border-primary-800 dark:bg-primary-900/20 dark:text-primary-400">
          <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
          </svg>
          {{ t('docs.entry.badge') }}
        </div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.entry.title') }}</h1>
        <p class="mt-3 text-gray-500 dark:text-gray-400">{{ t('docs.entry.subtitle') }}</p>
      </div>

      <!-- Tool Cards -->
      <div class="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <!-- Claude Code Card -->
        <router-link
          to="/docs/claude"
          class="group flex flex-col rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-700"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 shadow-md shadow-primary-500/25">
              <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
              </svg>
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">Claude Code</h2>
              </div>
              <p class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">claude-code CLI</p>
            </div>
          </div>
          <p class="flex-1 text-sm leading-relaxed text-gray-500 dark:text-gray-400">{{ t('docs.entry.claudeDesc') }}</p>
          <div class="mt-5 flex items-center gap-1 text-sm font-medium text-primary-600 transition-colors group-hover:text-primary-700 dark:text-primary-400 dark:group-hover:text-primary-300">
            {{ t('docs.entry.viewGuide') }}
            <svg class="h-4 w-4 transition-transform duration-150 group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
            </svg>
          </div>
        </router-link>

        <!-- Codex CLI Card -->
        <router-link
          to="/docs/codex"
          class="group flex flex-col rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-emerald-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-emerald-700"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 shadow-md shadow-emerald-500/25">
              <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75L22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3l-4.5 16.5" />
              </svg>
            </div>
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">Codex CLI</h2>
              <p class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">@openai/codex</p>
            </div>
          </div>
          <p class="flex-1 text-sm leading-relaxed text-gray-500 dark:text-gray-400">{{ t('docs.entry.codexDesc') }}</p>
          <div class="mt-5 flex items-center gap-1 text-sm font-medium text-emerald-600 transition-colors group-hover:text-emerald-700 dark:text-emerald-400 dark:group-hover:text-emerald-300">
            {{ t('docs.entry.viewGuide') }}
            <svg class="h-4 w-4 transition-transform duration-150 group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
            </svg>
          </div>
        </router-link>

        <!-- Openclaw / Clawdbot Card -->
        <router-link
          to="/docs/openclaw"
          class="group flex flex-col rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-violet-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-violet-700"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 shadow-md shadow-violet-500/25">
              <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904L9 18.75l-.813-2.846a4.5 4.5 0 00-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 003.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 003.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 00-3.09 3.09zM18.259 8.715L18 9.75l-.259-1.035a3.375 3.375 0 00-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 002.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 002.455 2.456L21.75 6l-1.036.259a3.375 3.375 0 00-2.455 2.456zM16.894 20.567L16.5 21.75l-.394-1.183a2.25 2.25 0 00-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 001.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 001.423 1.423l1.183.394-1.183.394a2.25 2.25 0 00-1.423 1.423z" />
              </svg>
            </div>
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">Openclaw</h2>
              <p class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">Clawdbot / openclaw</p>
            </div>
          </div>
          <p class="flex-1 text-sm leading-relaxed text-gray-500 dark:text-gray-400">{{ t('docs.entry.openclawDesc') }}</p>
          <div class="mt-5 flex items-center gap-1 text-sm font-medium text-violet-600 transition-colors group-hover:text-violet-700 dark:text-violet-400 dark:group-hover:text-violet-300">
            {{ t('docs.entry.viewGuide') }}
            <svg class="h-4 w-4 transition-transform duration-150 group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
            </svg>
          </div>
        </router-link>
      </div>

      <!-- Quick Steps -->
      <div class="mt-10 rounded-2xl border border-gray-200 bg-gray-50 p-6 dark:border-dark-700 dark:bg-dark-800/50">
        <h3 class="mb-6 text-sm font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.entry.quickInfo.title') }}</h3>
        <div class="relative">
          <!-- Connecting line -->
          <div class="absolute left-4 top-5 hidden h-[calc(100%-2.5rem)] w-px bg-gray-200 dark:bg-dark-600 sm:block"></div>
          <div class="space-y-5">
            <div v-for="(step, i) in quickSteps" :key="i" class="flex items-start gap-4">
              <div class="relative flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full border-2 border-primary-200 bg-white text-xs font-bold text-primary-600 dark:border-primary-800 dark:bg-dark-900 dark:text-primary-400">
                {{ i + 1 }}
              </div>
              <div class="pt-1">
                <div class="text-sm font-medium text-gray-900 dark:text-white">{{ step.title }}</div>
                <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ step.desc }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </PublicLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicLayout from '@/components/layout/PublicLayout.vue'

const { t } = useI18n()

const quickSteps = computed(() => [
  { title: t('docs.entry.quickInfo.step1Title'), desc: t('docs.entry.quickInfo.step1Desc') },
  { title: t('docs.entry.quickInfo.step2Title'), desc: t('docs.entry.quickInfo.step2Desc') },
  { title: t('docs.entry.quickInfo.step3Title'), desc: t('docs.entry.quickInfo.step3Desc') },
])
</script>
