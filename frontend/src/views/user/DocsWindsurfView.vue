<template>
  <PublicLayout>
    <div class="mx-auto max-w-4xl">
      <!-- Back Link -->
      <router-link to="/docs" class="mb-6 inline-flex items-center gap-2 text-sm text-gray-500 hover:text-primary-600 dark:text-gray-400 dark:hover:text-primary-400">
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
        </svg>
        {{ t('docs.backToList') }}
      </router-link>

      <!-- Header -->
      <div class="mb-8">
        <div class="mb-2 flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-teal-500 to-cyan-600 text-white">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14 10l-2 1m0 0l-2-1m2 1v2.5M20 7l-2 1m2-1l-2-1m2 1v2.5M14 4l-2-1-2 1M4 7l2-1M4 7l2 1M4 7v2.5M12 21l-2-1m2 1l2-1m-2 1v-2.5M6 18l-2-1v-2.5M18 18l2-1v-2.5" />
            </svg>
          </div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.windsurf.title') }}</h1>
        </div>
        <p class="text-gray-600 dark:text-gray-400">{{ t('docs.windsurf.subtitle') }}</p>
      </div>

      <!-- What is Windsurf -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-3 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.windsurf.whatIs.title') }}</h2>
        <p class="text-gray-600 dark:text-gray-400">{{ t('docs.windsurf.whatIs.desc') }}</p>
      </section>

      <!-- Step 1: Download -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">1</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.windsurf.step1.title') }}</h2>
        </div>
        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.windsurf.step1.desc') }}</p>
        <a
          href="https://codeium.com/windsurf"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-2 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-600"
        >
          {{ t('docs.windsurf.step1.download') }}
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
          </svg>
        </a>
        <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.windsurf.step1.tip') }}</p>
      </section>

      <!-- Step 2: Get API Key -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">2</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.windsurf.step2.title') }}</h2>
        </div>
        <ol class="mb-4 space-y-2 text-gray-600 dark:text-gray-400">
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">1.</span>
            <span>{{ t('docs.windsurf.step2.instruction1') }} <router-link to="/keys" class="text-primary-500 hover:underline">{{ t('docs.windsurf.step2.instruction1Link') }}</router-link></span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">2.</span>
            <span>{{ t('docs.windsurf.step2.instruction2') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">3.</span>
            <span>{{ t('docs.windsurf.step2.instruction3') }}</span>
          </li>
        </ol>

        <!-- Key Info Box -->
        <div class="rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-900/20">
          <p class="mb-2 text-sm font-medium text-blue-700 dark:text-blue-300">{{ t('docs.windsurf.step2.infoTitle') }}</p>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <span class="text-sm text-blue-600 dark:text-blue-400">Base URL:</span>
              <code class="rounded bg-white px-2 py-0.5 text-sm dark:bg-dark-800">{{ baseUrl }}</code>
              <button @click="copyToClipboard(baseUrl)" class="text-blue-500 hover:text-blue-600">
                <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-sm text-blue-600 dark:text-blue-400">API Key:</span>
              <code class="rounded bg-white px-2 py-0.5 text-sm dark:bg-dark-800">{{ t('docs.windsurf.step2.yourKey') }}</code>
            </div>
          </div>
        </div>
      </section>

      <!-- Step 3: Configure Windsurf -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">3</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.windsurf.step3.title') }}</h2>
        </div>

        <ol class="space-y-4 text-gray-600 dark:text-gray-400">
          <li>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.step3.instruction1') }}</p>
            <p class="mt-1 text-sm">{{ t('docs.windsurf.step3.instruction1Detail') }}</p>
          </li>
          <li>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.step3.instruction2') }}</p>
            <p class="mt-1 text-sm">{{ t('docs.windsurf.step3.instruction2Detail') }}</p>
          </li>
          <li>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.step3.instruction3') }}</p>
            <p class="mt-1 text-sm">{{ t('docs.windsurf.step3.instruction3Detail') }}</p>
          </li>
          <li>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.step3.instruction4') }}</p>
            <div class="mt-2 rounded-lg bg-gray-900 p-4">
              <div class="space-y-2 text-sm">
                <div>
                  <span class="text-gray-500">Base URL: </span>
                  <span class="text-yellow-300">{{ baseUrl }}</span>
                </div>
                <div>
                  <span class="text-gray-500">API Key: </span>
                  <span class="text-yellow-300">{{ t('docs.windsurf.step2.yourKey') }}</span>
                </div>
              </div>
            </div>
          </li>
          <li>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.step3.instruction5') }}</p>
          </li>
        </ol>
      </section>

      <!-- Step 4: Verify -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">4</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.windsurf.step4.title') }}</h2>
        </div>
        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.windsurf.step4.desc') }}</p>
        <div class="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
          <p class="text-sm text-green-700 dark:text-green-300">{{ t('docs.windsurf.step4.success') }}</p>
        </div>
      </section>

      <!-- FAQ -->
      <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.windsurf.faq.title') }}</h3>
        <div class="space-y-4">
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.faq.q1') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.windsurf.faq.a1') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.faq.q2') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.windsurf.faq.a2') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.windsurf.faq.q3') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.windsurf.faq.a3') }}</p>
          </div>
        </div>
      </section>
    </div>
  </PublicLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import PublicLayout from '@/components/layout/PublicLayout.vue'

const { t } = useI18n()
const appStore = useAppStore()

const baseUrl = computed(() => window.location.origin)

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
  appStore.showSuccess(t('docs.copied'))
}
</script>
