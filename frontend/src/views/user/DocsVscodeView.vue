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
          <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 text-white">
            <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
              <path d="M17.583 2L6.25 11.042 2.833 8.375.5 9.75l4.917 4.25L.5 18.25l2.333 1.375 3.417-2.667L17.583 26 21.5 24.25V3.75L17.583 2zm-.833 3.958v16.084l-8.917-8.042 8.917-8.042z"/>
            </svg>
          </div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.vscode.title') }}</h1>
        </div>
        <p class="text-gray-600 dark:text-gray-400">{{ t('docs.vscode.subtitle') }}</p>
      </div>

      <!-- Plugin Options -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-4 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.vscode.plugins.title') }}</h2>
        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.vscode.plugins.desc') }}</p>

        <div class="grid gap-4 sm:grid-cols-2">
          <!-- Cline -->
          <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
            <div class="mb-2 flex items-center gap-2">
              <h3 class="font-semibold text-gray-900 dark:text-white">Cline</h3>
              <span class="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">{{ t('docs.entry.recommended') }}</span>
            </div>
            <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.vscode.plugins.cline') }}</p>
            <a
              href="https://marketplace.visualstudio.com/items?itemName=saoudrizwan.claude-dev"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1 text-sm text-primary-500 hover:underline"
            >
              {{ t('docs.vscode.plugins.install') }}
              <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
            </a>
          </div>

          <!-- Claude Code for VS Code -->
          <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
            <div class="mb-2 flex items-center gap-2">
              <h3 class="font-semibold text-gray-900 dark:text-white">Claude Code for VS Code</h3>
              <span class="rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{{ t('docs.vscode.plugins.official') }}</span>
            </div>
            <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.vscode.plugins.claudeCode') }}</p>
            <a
              href="https://marketplace.visualstudio.com/items?itemName=anthropic.claude-code"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-1 text-sm text-primary-500 hover:underline"
            >
              {{ t('docs.vscode.plugins.install') }}
              <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
            </a>
          </div>
        </div>
      </section>

      <!-- Step 1: Install Plugin -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">1</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.vscode.step1.title') }}</h2>
        </div>
        <ol class="space-y-2 text-gray-600 dark:text-gray-400">
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">1.</span>
            <span>{{ t('docs.vscode.step1.instruction1') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">2.</span>
            <span>{{ t('docs.vscode.step1.instruction2') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">3.</span>
            <span>{{ t('docs.vscode.step1.instruction3') }}</span>
          </li>
        </ol>
        <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">
          {{ t('docs.vscode.step1.tip') }}
        </p>
      </section>

      <!-- Step 2: Get API Key -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">2</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.vscode.step2.title') }}</h2>
        </div>
        <ol class="mb-4 space-y-2 text-gray-600 dark:text-gray-400">
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">1.</span>
            <span>{{ t('docs.vscode.step2.instruction1') }} <router-link to="/keys" class="text-primary-500 hover:underline">{{ t('docs.vscode.step2.instruction1Link') }}</router-link></span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">2.</span>
            <span>{{ t('docs.vscode.step2.instruction2') }}</span>
          </li>
        </ol>

        <!-- Key Info Box -->
        <div class="rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-900/20">
          <p class="mb-2 text-sm font-medium text-blue-700 dark:text-blue-300">{{ t('docs.vscode.step2.infoTitle') }}</p>
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
              <code class="rounded bg-white px-2 py-0.5 text-sm dark:bg-dark-800">{{ t('docs.vscode.step2.yourKey') }}</code>
            </div>
          </div>
        </div>
      </section>

      <!-- Step 3: Configure Plugin -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">3</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.vscode.step3.title') }}</h2>
        </div>

        <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.vscode.step3.selectPlugin') }}</p>

        <!-- Cline Config -->
        <div class="mb-6 rounded-lg border border-green-200 bg-green-50/50 p-4 dark:border-green-800 dark:bg-green-900/10">
          <div class="mb-3 flex items-center gap-2">
            <h4 class="font-semibold text-gray-800 dark:text-gray-200">Cline</h4>
            <span class="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">{{ t('docs.entry.recommended') }}</span>
          </div>
          <ol class="space-y-4 text-sm text-gray-600 dark:text-gray-400">
            <!-- Step 1 -->
            <li>
              <div class="mb-2 flex items-start gap-2">
                <span class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-green-500 text-xs font-medium text-white">1</span>
                <span>{{ t('docs.vscode.step3.cline1') }}</span>
              </div>
              <div class="ml-7 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
                <img src="/1.png" alt="Cline Step 1" class="w-full max-w-md" />
              </div>
            </li>
            <!-- Step 2 -->
            <li>
              <div class="mb-2 flex items-start gap-2">
                <span class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-green-500 text-xs font-medium text-white">2</span>
                <div>
                  <span>{{ t('docs.vscode.step3.cline2') }}</span>
                  <div class="mt-2 rounded-lg bg-gray-900 p-3">
                    <div class="space-y-1 text-sm">
                      <div><span class="text-gray-500">API Provider: </span><span class="text-yellow-300">Anthropic</span></div>
                      <div><span class="text-gray-500">Anthropic API Key: </span><span class="text-yellow-300">{{ t('docs.vscode.step2.yourKey') }}</span></div>
                      <div><span class="text-gray-500">Use custom base URL: </span><span class="text-green-400">✓</span></div>
                      <div><span class="text-gray-500">Base URL: </span><span class="text-yellow-300">{{ baseUrl }}</span></div>
                      <div><span class="text-gray-500">Model: </span><span class="text-yellow-300">claude-opus-4-5-20251101</span></div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="ml-7 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
                <img src="/2.png" alt="Cline Step 2" class="w-full max-w-md" />
              </div>
            </li>
            <!-- Step 3 -->
            <li>
              <div class="mb-2 flex items-start gap-2">
                <span class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-green-500 text-xs font-medium text-white">3</span>
                <span>{{ t('docs.vscode.step3.cline3') }}</span>
              </div>
              <div class="ml-7 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
                <img src="/3.png" alt="Cline Step 3" class="w-full max-w-md" />
              </div>
            </li>
          </ol>
        </div>

        <!-- Claude Code for VS Code Config -->
        <div class="rounded-lg border border-blue-200 bg-blue-50/50 p-4 dark:border-blue-800 dark:bg-blue-900/10">
          <div class="mb-3 flex items-center gap-2">
            <h4 class="font-semibold text-gray-800 dark:text-gray-200">Claude Code for VS Code</h4>
            <span class="rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{{ t('docs.vscode.plugins.official') }}</span>
          </div>
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.vscode.step3.claudeCodeDesc') }}</p>
          <ol class="space-y-3 text-sm text-gray-600 dark:text-gray-400">
            <li class="flex items-start gap-2">
              <span class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-blue-500 text-xs font-medium text-white">1</span>
              <span>{{ t('docs.vscode.step3.claudeCode1') }}</span>
            </li>
            <li class="flex items-start gap-2">
              <span class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-blue-500 text-xs font-medium text-white">2</span>
              <span>{{ t('docs.vscode.step3.claudeCode2') }}</span>
            </li>
            <li class="flex items-start gap-2">
              <span class="flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-blue-500 text-xs font-medium text-white">3</span>
              <span>{{ t('docs.vscode.step3.claudeCode3') }}</span>
            </li>
          </ol>
          <div class="mt-3 rounded-lg border border-blue-300 bg-blue-100 p-3 dark:border-blue-700 dark:bg-blue-900/30">
            <p class="text-sm text-blue-700 dark:text-blue-300">
              <span class="font-medium">{{ t('docs.vscode.step3.claudeCodeTipTitle') }}</span>
              {{ t('docs.vscode.step3.claudeCodeTip') }}
            </p>
          </div>
        </div>
      </section>

      <!-- Step 4: Verify -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">4</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.vscode.step4.title') }}</h2>
        </div>
        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.vscode.step4.desc') }}</p>
        <div class="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
          <p class="text-sm text-green-700 dark:text-green-300">{{ t('docs.vscode.step4.success') }}</p>
        </div>
      </section>

      <!-- FAQ -->
      <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.vscode.faq.title') }}</h3>
        <div class="space-y-4">
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.vscode.faq.q1') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.vscode.faq.a1') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.vscode.faq.q2') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.vscode.faq.a2') }}</p>
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
