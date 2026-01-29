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
          <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-amber-500 to-orange-500 text-white">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
          </div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.tips.title') }}</h1>
        </div>
        <p class="text-gray-600 dark:text-gray-400">{{ t('docs.tips.subtitle') }}</p>
      </div>

      <!-- Common Commands -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-4 flex items-center gap-2 text-xl font-semibold text-gray-900 dark:text-white">
          <svg class="h-5 w-5 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          {{ t('docs.tips.commands.title') }}
        </h2>

        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-200 dark:border-dark-600">
                <th class="py-3 pr-4 text-left text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('docs.tips.commands.command') }}</th>
                <th class="py-3 pr-4 text-left text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('docs.tips.commands.function') }}</th>
                <th class="py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('docs.tips.commands.usage') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="cmd in commands" :key="cmd.command">
                <td class="py-3 pr-4">
                  <code class="rounded bg-gray-100 px-2 py-1 text-sm font-medium text-primary-600 dark:bg-dark-700 dark:text-primary-400">{{ cmd.command }}</code>
                </td>
                <td class="py-3 pr-4 text-sm text-gray-700 dark:text-gray-300">{{ t(cmd.functionKey) }}</td>
                <td class="py-3 text-sm text-gray-500 dark:text-gray-400">{{ t(cmd.usageKey) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- Keyboard Shortcuts -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-4 flex items-center gap-2 text-xl font-semibold text-gray-900 dark:text-white">
          <svg class="h-5 w-5 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
          </svg>
          {{ t('docs.tips.shortcuts.title') }}
        </h2>

        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-200 dark:border-dark-600">
                <th class="py-3 pr-4 text-left text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('docs.tips.shortcuts.shortcut') }}</th>
                <th class="py-3 text-left text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('docs.tips.shortcuts.function') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="shortcut in shortcuts" :key="shortcut.keys">
                <td class="py-3 pr-4">
                  <kbd class="rounded border border-gray-300 bg-gray-50 px-2 py-1 text-sm font-medium text-gray-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300">{{ shortcut.keys }}</kbd>
                </td>
                <td class="py-3 text-sm text-gray-700 dark:text-gray-300">{{ t(shortcut.functionKey) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- Model Selection -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-4 flex items-center gap-2 text-xl font-semibold text-gray-900 dark:text-white">
          <svg class="h-5 w-5 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          {{ t('docs.tips.models.title') }}
        </h2>

        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.tips.models.desc') }}</p>

        <div class="grid gap-4 sm:grid-cols-2">
          <!-- Sonnet -->
          <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
            <div class="mb-2 flex items-center gap-2">
              <span class="rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">Sonnet</span>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('docs.tips.models.sonnetLabel') }}</span>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('docs.tips.models.sonnetDesc') }}</p>
            <ul class="mt-2 space-y-1 text-xs text-gray-500 dark:text-gray-400">
              <li>• {{ t('docs.tips.models.sonnetUse1') }}</li>
              <li>• {{ t('docs.tips.models.sonnetUse2') }}</li>
              <li>• {{ t('docs.tips.models.sonnetUse3') }}</li>
            </ul>
          </div>

          <!-- Opus -->
          <div class="rounded-lg border border-purple-200 bg-purple-50/50 p-4 dark:border-purple-800 dark:bg-purple-900/10">
            <div class="mb-2 flex items-center gap-2">
              <span class="rounded bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900/30 dark:text-purple-400">Opus</span>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ t('docs.tips.models.opusLabel') }}</span>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('docs.tips.models.opusDesc') }}</p>
            <ul class="mt-2 space-y-1 text-xs text-gray-500 dark:text-gray-400">
              <li>• {{ t('docs.tips.models.opusUse1') }}</li>
              <li>• {{ t('docs.tips.models.opusUse2') }}</li>
              <li>• {{ t('docs.tips.models.opusUse3') }}</li>
            </ul>
          </div>
        </div>

        <div class="mt-4 rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-900/20">
          <p class="text-sm text-blue-700 dark:text-blue-300">
            <span class="font-medium">{{ t('docs.tips.models.switchTip') }}</span>
            {{ t('docs.tips.models.switchDesc') }}
            <code class="rounded bg-white px-1.5 py-0.5 text-xs dark:bg-dark-800">/model</code>
          </p>
        </div>
      </section>

      <!-- Best Practices -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-4 flex items-center gap-2 text-xl font-semibold text-gray-900 dark:text-white">
          <svg class="h-5 w-5 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z" />
          </svg>
          {{ t('docs.tips.practices.title') }}
        </h2>

        <div class="space-y-4">
          <div class="flex items-start gap-3">
            <div class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.tips.practices.tip1Title') }}</p>
              <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.tips.practices.tip1Desc') }}</p>
            </div>
          </div>

          <div class="flex items-start gap-3">
            <div class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.tips.practices.tip2Title') }}</p>
              <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.tips.practices.tip2Desc') }}</p>
            </div>
          </div>

          <div class="flex items-start gap-3">
            <div class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.tips.practices.tip3Title') }}</p>
              <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.tips.practices.tip3Desc') }}</p>
            </div>
          </div>

          <div class="flex items-start gap-3">
            <div class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.tips.practices.tip4Title') }}</p>
              <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.tips.practices.tip4Desc') }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- Quick Links -->
      <section class="rounded-xl border border-blue-200 bg-blue-50 p-6 dark:border-blue-800 dark:bg-blue-900/20">
        <h3 class="mb-4 font-semibold text-blue-800 dark:text-blue-200">{{ t('docs.tips.links.title') }}</h3>
        <div class="flex flex-wrap gap-3">
          <router-link to="/keys" class="inline-flex items-center gap-2 rounded-lg bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
            {{ t('docs.tips.links.keys') }}
          </router-link>
          <router-link to="/dashboard" class="inline-flex items-center gap-2 rounded-lg bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
            {{ t('docs.tips.links.usage') }}
          </router-link>
          <router-link to="/status" class="inline-flex items-center gap-2 rounded-lg bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ t('docs.tips.links.status') }}
          </router-link>
        </div>
      </section>
    </div>
  </PublicLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import PublicLayout from '@/components/layout/PublicLayout.vue'

const { t } = useI18n()

const commands = [
  { command: '/model', functionKey: 'docs.tips.commands.modelFunc', usageKey: 'docs.tips.commands.modelUsage' },
  { command: '/compact', functionKey: 'docs.tips.commands.compactFunc', usageKey: 'docs.tips.commands.compactUsage' },
  { command: '/clear', functionKey: 'docs.tips.commands.clearFunc', usageKey: 'docs.tips.commands.clearUsage' },
  { command: '/cost', functionKey: 'docs.tips.commands.costFunc', usageKey: 'docs.tips.commands.costUsage' },
  { command: '/help', functionKey: 'docs.tips.commands.helpFunc', usageKey: 'docs.tips.commands.helpUsage' }
]

const shortcuts = [
  { keys: 'Esc', functionKey: 'docs.tips.shortcuts.escFunc' },
  { keys: 'Shift+Tab', functionKey: 'docs.tips.shortcuts.shiftTabFunc' },
  { keys: 'Ctrl+C', functionKey: 'docs.tips.shortcuts.ctrlCFunc' }
]
</script>
