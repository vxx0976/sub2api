<template>
  <PublicLayout>
    <div class="mx-auto max-w-3xl px-4 py-8">
      <!-- Header -->
      <div class="mb-10 text-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.entry.title') }}</h1>
        <p class="mt-3 text-gray-600 dark:text-gray-400">{{ t('docs.entry.subtitle') }}</p>
      </div>

      <!-- Tool Cards -->
      <div class="grid gap-6 sm:grid-cols-2">
        <!-- Claude Code Card -->
        <router-link
          to="/docs/claude"
          class="group relative flex flex-col rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition hover:border-primary-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-700"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 shadow-lg shadow-primary-500/20">
              <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
              </svg>
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Claude Code</h2>
                <span class="rounded-full bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">{{ t('docs.entry.recommended') }}</span>
              </div>
              <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">claude-code CLI</p>
            </div>
          </div>
          <p class="flex-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.claudeDesc') }}</p>
          <div class="mt-4 flex items-center gap-1 text-sm font-medium text-primary-600 group-hover:text-primary-700 dark:text-primary-400 dark:group-hover:text-primary-300">
            {{ t('docs.entry.viewGuide') }}
            <svg class="h-4 w-4 transition-transform group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
            </svg>
          </div>
        </router-link>

        <!-- Codex CLI Card -->
        <router-link
          to="/docs/codex"
          class="group relative flex flex-col rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition hover:border-emerald-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-emerald-700"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-teal-600 shadow-lg shadow-emerald-500/20">
              <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75L22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3l-4.5 16.5" />
              </svg>
            </div>
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Codex CLI</h2>
              <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">@openai/codex</p>
            </div>
          </div>
          <p class="flex-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.codexDesc') }}</p>
          <div class="mt-4 flex items-center gap-1 text-sm font-medium text-emerald-600 group-hover:text-emerald-700 dark:text-emerald-400 dark:group-hover:text-emerald-300">
            {{ t('docs.entry.viewGuide') }}
            <svg class="h-4 w-4 transition-transform group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
            </svg>
          </div>
        </router-link>
      </div>

      <!-- Quick Info -->
      <div class="mt-10 rounded-2xl border border-gray-200 bg-gray-50 p-6 dark:border-dark-700 dark:bg-dark-800/50">
        <h3 class="mb-4 text-base font-semibold text-gray-900 dark:text-white">{{ t('docs.entry.quickInfo.title') }}</h3>
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="flex items-start gap-3">
            <div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300 text-sm font-bold">1</div>
            <div>
              <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('docs.entry.quickInfo.step1Title') }}</div>
              <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('docs.entry.quickInfo.step1Desc') }}</div>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300 text-sm font-bold">2</div>
            <div>
              <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('docs.entry.quickInfo.step2Title') }}</div>
              <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('docs.entry.quickInfo.step2Desc') }}</div>
            </div>
          </div>
          <div class="flex items-start gap-3">
            <div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300 text-sm font-bold">3</div>
            <div>
              <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('docs.entry.quickInfo.step3Title') }}</div>
              <div class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('docs.entry.quickInfo.step3Desc') }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </PublicLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import PublicLayout from '@/components/layout/PublicLayout.vue'

const { t } = useI18n()
</script>
