<template>
  <PublicLayout>
    <div class="mx-auto flex max-w-6xl gap-6 xl:px-4">
      <!-- Left sidebar navigation (xl screens only) -->
      <aside class="hidden xl:block w-48 flex-shrink-0">
        <nav class="sticky top-24 space-y-1 rounded-lg border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-800">
          <a
            v-for="item in navItems"
            :key="item.id"
            href="#"
            @click.prevent="scrollTo(item.id)"
            :class="[
              'block truncate rounded px-2.5 py-2 text-sm transition-colors',
              activeSection === item.id
                ? 'bg-primary-50 font-medium text-primary-600 dark:bg-primary-900/30 dark:text-primary-400'
                : 'text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-300'
            ]"
            :title="item.label"
          >
            {{ item.label }}
          </a>
        </nav>
      </aside>

      <!-- Main content -->
      <div class="min-w-0 flex-1 max-w-4xl">
        <!-- Back link -->
        <div class="mb-6">
          <router-link to="/docs" class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300">
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
            </svg>
            {{ t('docs.backToList') }}
          </router-link>
        </div>

        <!-- Header -->
        <div class="mb-8 text-center">
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.openclaw.title') }}</h1>
          <p class="mt-3 text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.subtitle') }}</p>
        </div>

        <!-- Prerequisites -->
        <section id="prereq" class="scroll-mt-20 mb-8 rounded-xl border border-amber-200 bg-amber-50 p-6 dark:border-amber-800 dark:bg-amber-900/20">
          <h3 class="mb-3 flex items-center gap-2 font-semibold text-amber-800 dark:text-amber-200">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            {{ t('docs.openclaw.prereq.title') }}
          </h3>
          <ul class="space-y-2 text-sm text-amber-700 dark:text-amber-300">
            <li class="flex items-start gap-2">
              <span class="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-amber-500"></span>
              {{ t('docs.openclaw.prereq.item1') }}
            </li>
            <li class="flex items-start gap-2">
              <span class="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-amber-500"></span>
              {{ t('docs.openclaw.prereq.item2') }}
            </li>
            <li class="flex items-start gap-2">
              <span class="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-amber-500"></span>
              {{ t('docs.openclaw.prereq.item3') }}
            </li>
          </ul>
        </section>

        <!-- Step 1: Install -->
        <section id="install" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">1</span>
            {{ t('docs.openclaw.step1.title') }}
          </h2>
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.step1.desc') }}</p>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <button @click="copyCode('install')" class="absolute right-2 top-2 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
              <svg v-if="copied !== 'install'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
              <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            </button>
            <code class="text-sm text-gray-100">npm install -g openclaw@latest</code>
          </div>
        </section>

        <!-- Step 2: Initialize -->
        <section id="init" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">2</span>
            {{ t('docs.openclaw.step2.title') }}
          </h2>
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.step2.desc') }}</p>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <button @click="copyCode('init')" class="absolute right-2 top-2 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
              <svg v-if="copied !== 'init'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
              <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            </button>
            <code class="text-sm text-gray-100">openclaw onboard --install-daemon</code>
          </div>
        </section>

        <!-- Step 3: Config file -->
        <section id="config" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">3</span>
            {{ t('docs.openclaw.step3.title') }}
          </h2>
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.step3.desc') }}</p>
          <div class="rounded-lg bg-gray-100 px-4 py-2 text-sm font-mono text-gray-700 dark:bg-dark-700 dark:text-gray-300">
            ~/.openclaw/openclaw.json
          </div>
          <p class="mt-4 mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.step3.editDesc') }}</p>

          <!-- JSON config code block -->
          <div class="relative rounded-lg bg-gray-900 p-4 overflow-x-auto">
            <div class="absolute right-2 top-2 flex items-center gap-2">
              <span class="rounded bg-gray-700 px-2 py-0.5 text-xs text-gray-400">JSON</span>
              <button @click="copyCode('config')" class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
                <svg v-if="copied !== 'config'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
                <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
              </button>
            </div>
            <pre class="text-sm leading-relaxed"><code class="text-gray-100" v-html="configJson"></code></pre>
          </div>
        </section>

        <!-- Step 4: Parameter description -->
        <section id="params" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">4</span>
            {{ t('docs.openclaw.step4.title') }}
          </h2>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 dark:border-dark-600">
                  <th class="pb-2 pr-4 text-left font-semibold text-gray-700 dark:text-gray-300">{{ t('docs.openclaw.step4.param') }}</th>
                  <th class="pb-2 text-left font-semibold text-gray-700 dark:text-gray-300">{{ t('docs.openclaw.step4.description') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="param in paramTable" :key="param.name">
                  <td class="py-2 pr-4">
                    <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-medium text-violet-600 dark:bg-dark-700 dark:text-violet-400">{{ param.name }}</code>
                  </td>
                  <td class="py-2 text-gray-600 dark:text-gray-400">{{ param.desc }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- Step 5: Switch model -->
        <section id="switch" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">5</span>
            {{ t('docs.openclaw.step5.title') }}
          </h2>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <button @click="copyCode('switch')" class="absolute right-2 top-2 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
              <svg v-if="copied !== 'switch'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
              <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            </button>
            <code class="text-sm text-gray-100">openclaw models set mayione/claude-opus-4-6</code>
          </div>
        </section>

        <!-- Step 6: Restart gateway -->
        <section id="restart" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">6</span>
            {{ t('docs.openclaw.step6.title') }}
          </h2>
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.step6.desc') }}</p>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <button @click="copyCode('restart')" class="absolute right-2 top-2 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
              <svg v-if="copied !== 'restart'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
              <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            </button>
            <code class="text-sm text-gray-100">openclaw gateway restart</code>
          </div>
        </section>

        <!-- Step 7: Verify -->
        <section id="verify" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">7</span>
            {{ t('docs.openclaw.step7.title') }}
          </h2>
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.step7.desc1') }}</p>
          <div class="relative mb-4 rounded-lg bg-gray-900 p-4">
            <button @click="copyCode('list')" class="absolute right-2 top-2 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
              <svg v-if="copied !== 'list'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
              <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            </button>
            <code class="text-sm text-gray-100">openclaw models list</code>
          </div>
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.openclaw.step7.desc2') }}</p>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <button @click="copyCode('status')" class="absolute right-2 top-2 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
              <svg v-if="copied !== 'status'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
              <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            </button>
            <code class="text-sm text-gray-100">openclaw models status</code>
          </div>
        </section>

        <!-- Step 8: Start using -->
        <section id="start" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 flex items-center gap-2 text-lg font-semibold text-gray-900 dark:text-white">
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-violet-100 text-sm font-bold text-violet-600 dark:bg-violet-900/30 dark:text-violet-400">8</span>
            {{ t('docs.openclaw.step8.title') }}
          </h2>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <button @click="copyCode('tui')" class="absolute right-2 top-2 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-700 hover:text-white">
              <svg v-if="copied !== 'tui'" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" /></svg>
              <svg v-else class="h-4 w-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
            </button>
            <pre class="text-sm text-gray-100"><span class="text-gray-500"># {{ t('docs.openclaw.step8.tui') }}</span>
openclaw tui

<span class="text-gray-500"># {{ t('docs.openclaw.step8.message') }}</span>
openclaw message "{{ t('docs.openclaw.step8.hello') }}"</pre>
          </div>
        </section>

        <!-- FAQ -->
        <section id="faq" class="scroll-mt-20 mb-12 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.openclaw.faq.title') }}</h2>
          <div class="space-y-4">
            <div v-for="(item, i) in faqItems" :key="i" class="rounded-lg bg-gray-50 p-4 dark:bg-dark-700/50">
              <h3 class="mb-2 font-medium text-gray-900 dark:text-white">{{ t(item.q) }}</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ t(item.a) }}</p>
            </div>
          </div>
        </section>
      </div>
    </div>
  </PublicLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicLayout from '@/components/layout/PublicLayout.vue'

const { t } = useI18n()
const copied = ref<string | null>(null)
const activeSection = ref('prereq')

const baseUrl = computed(() => window.location.origin)

const navItems = computed(() => [
  { id: 'prereq', label: t('docs.openclaw.prereq.title') },
  { id: 'install', label: t('docs.openclaw.step1.title') },
  { id: 'init', label: t('docs.openclaw.step2.title') },
  { id: 'config', label: t('docs.openclaw.step3.title') },
  { id: 'params', label: t('docs.openclaw.step4.title') },
  { id: 'switch', label: t('docs.openclaw.step5.title') },
  { id: 'restart', label: t('docs.openclaw.step6.title') },
  { id: 'verify', label: t('docs.openclaw.step7.title') },
  { id: 'start', label: t('docs.openclaw.step8.title') },
  { id: 'faq', label: t('docs.openclaw.faq.title') },
])

const paramTable = computed(() => [
  { name: 'baseUrl', desc: t('docs.openclaw.step4.baseUrl') },
  { name: 'apiKey', desc: t('docs.openclaw.step4.apiKey') },
  { name: 'api', desc: t('docs.openclaw.step4.api') },
  { name: 'id', desc: t('docs.openclaw.step4.id') },
  { name: 'name', desc: t('docs.openclaw.step4.name') },
  { name: 'contextWindow', desc: t('docs.openclaw.step4.contextWindow') },
  { name: 'maxTokens', desc: t('docs.openclaw.step4.maxTokens') },
  { name: 'headers', desc: t('docs.openclaw.step4.headers') },
])

const yourKey = computed(() => t('docs.openclaw.yourKey'))

const configJson = computed(() => {
  const url = baseUrl.value
  const key = yourKey.value
  return `{
  <span class="text-blue-300">"models"</span>: {
    <span class="text-blue-300">"mode"</span>: <span class="text-yellow-300">"merge"</span>,
    <span class="text-blue-300">"providers"</span>: {
      <span class="text-blue-300">"mayione"</span>: {
        <span class="text-blue-300">"baseUrl"</span>: <span class="text-yellow-300">"${url}"</span>,
        <span class="text-blue-300">"apiKey"</span>: <span class="text-yellow-300">"${key}"</span>,  <span class="text-gray-500">// &lt;--- ${t('docs.openclaw.step3.changeKey')}</span>
        <span class="text-blue-300">"api"</span>: <span class="text-yellow-300">"anthropic-messages"</span>,
        <span class="text-blue-300">"models"</span>: [
          {
            <span class="text-blue-300">"id"</span>: <span class="text-yellow-300">"claude-opus-4-5-20251101"</span>,
            <span class="text-blue-300">"name"</span>: <span class="text-yellow-300">"Claude Opus 4.5"</span>,
            <span class="text-blue-300">"reasoning"</span>: <span class="text-orange-300">false</span>,
            <span class="text-blue-300">"input"</span>: [<span class="text-yellow-300">"text"</span>],
            <span class="text-blue-300">"cost"</span>: { <span class="text-blue-300">"input"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"output"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheRead"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheWrite"</span>: <span class="text-orange-300">0</span> },
            <span class="text-blue-300">"contextWindow"</span>: <span class="text-orange-300">170000</span>,
            <span class="text-blue-300">"maxTokens"</span>: <span class="text-orange-300">64000</span>,
            <span class="text-blue-300">"headers"</span>: { <span class="text-blue-300">"User-Agent"</span>: <span class="text-yellow-300">"Mozilla/5.0"</span>, <span class="text-blue-300">"Content-Type"</span>: <span class="text-yellow-300">"application/json"</span> }
          },
          {
            <span class="text-blue-300">"id"</span>: <span class="text-yellow-300">"claude-sonnet-4-5-20250929"</span>,
            <span class="text-blue-300">"name"</span>: <span class="text-yellow-300">"Claude Sonnet 4.5"</span>,
            <span class="text-blue-300">"reasoning"</span>: <span class="text-orange-300">false</span>,
            <span class="text-blue-300">"input"</span>: [<span class="text-yellow-300">"text"</span>],
            <span class="text-blue-300">"cost"</span>: { <span class="text-blue-300">"input"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"output"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheRead"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheWrite"</span>: <span class="text-orange-300">0</span> },
            <span class="text-blue-300">"contextWindow"</span>: <span class="text-orange-300">170000</span>,
            <span class="text-blue-300">"maxTokens"</span>: <span class="text-orange-300">64000</span>,
            <span class="text-blue-300">"headers"</span>: { <span class="text-blue-300">"User-Agent"</span>: <span class="text-yellow-300">"Mozilla/5.0"</span>, <span class="text-blue-300">"Content-Type"</span>: <span class="text-yellow-300">"application/json"</span> }
          },
          {
            <span class="text-blue-300">"id"</span>: <span class="text-yellow-300">"claude-opus-4-6"</span>,
            <span class="text-blue-300">"name"</span>: <span class="text-yellow-300">"Claude Opus 4.6"</span>,
            <span class="text-blue-300">"reasoning"</span>: <span class="text-orange-300">false</span>,
            <span class="text-blue-300">"input"</span>: [<span class="text-yellow-300">"text"</span>],
            <span class="text-blue-300">"cost"</span>: { <span class="text-blue-300">"input"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"output"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheRead"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheWrite"</span>: <span class="text-orange-300">0</span> },
            <span class="text-blue-300">"contextWindow"</span>: <span class="text-orange-300">170000</span>,
            <span class="text-blue-300">"maxTokens"</span>: <span class="text-orange-300">64000</span>,
            <span class="text-blue-300">"headers"</span>: { <span class="text-blue-300">"User-Agent"</span>: <span class="text-yellow-300">"Mozilla/5.0"</span>, <span class="text-blue-300">"Content-Type"</span>: <span class="text-yellow-300">"application/json"</span> }
          },
          {
            <span class="text-blue-300">"id"</span>: <span class="text-yellow-300">"claude-haiku-4-5-20251001"</span>,
            <span class="text-blue-300">"name"</span>: <span class="text-yellow-300">"Claude Haiku 4.5"</span>,
            <span class="text-blue-300">"reasoning"</span>: <span class="text-orange-300">false</span>,
            <span class="text-blue-300">"input"</span>: [<span class="text-yellow-300">"text"</span>],
            <span class="text-blue-300">"cost"</span>: { <span class="text-blue-300">"input"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"output"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheRead"</span>: <span class="text-orange-300">0</span>, <span class="text-blue-300">"cacheWrite"</span>: <span class="text-orange-300">0</span> },
            <span class="text-blue-300">"contextWindow"</span>: <span class="text-orange-300">170000</span>,
            <span class="text-blue-300">"maxTokens"</span>: <span class="text-orange-300">64000</span>,
            <span class="text-blue-300">"headers"</span>: { <span class="text-blue-300">"User-Agent"</span>: <span class="text-yellow-300">"Mozilla/5.0"</span>, <span class="text-blue-300">"Content-Type"</span>: <span class="text-yellow-300">"application/json"</span> }
          }
        ]
      }
    }
  },
  <span class="text-blue-300">"agents"</span>: {
    <span class="text-blue-300">"defaults"</span>: {
      <span class="text-blue-300">"model"</span>: { <span class="text-blue-300">"primary"</span>: <span class="text-yellow-300">"claude-sonnet-4-5-20250929"</span> },
      <span class="text-blue-300">"models"</span>: {
        <span class="text-blue-300">"mayione/claude-opus-4-6"</span>: {},
        <span class="text-blue-300">"mayione/claude-opus-4-5-20251101"</span>: {},
        <span class="text-blue-300">"mayione/claude-sonnet-4-5-20250929"</span>: {},
        <span class="text-blue-300">"mayione/claude-haiku-4-5-20251001"</span>: {}
      },
      <span class="text-blue-300">"workspace"</span>: <span class="text-yellow-300">"~/.openclaw/workspace"</span>,  <span class="text-gray-500">// &lt;--- ${t('docs.openclaw.step3.changeWorkspace')}</span>
      <span class="text-blue-300">"compaction"</span>: { <span class="text-blue-300">"mode"</span>: <span class="text-yellow-300">"safeguard"</span> },
      <span class="text-blue-300">"maxConcurrent"</span>: <span class="text-orange-300">4</span>,
      <span class="text-blue-300">"subagents"</span>: { <span class="text-blue-300">"maxConcurrent"</span>: <span class="text-orange-300">8</span> }
    }
  }
}`
})

const faqItems = [
  { q: 'docs.openclaw.faq.q1', a: 'docs.openclaw.faq.a1' },
  { q: 'docs.openclaw.faq.q2', a: 'docs.openclaw.faq.a2' },
  { q: 'docs.openclaw.faq.q3', a: 'docs.openclaw.faq.a3' },
  { q: 'docs.openclaw.faq.q4', a: 'docs.openclaw.faq.a4' },
]

const getConfigJsonRaw = () => {
  const url = baseUrl.value
  return `{
  "models": {
    "mode": "merge",
    "providers": {
      "mayione": {
        "baseUrl": "${url}",
        "apiKey": "sk-xxx",
        "api": "anthropic-messages",
        "models": [
          {
            "id": "claude-opus-4-5-20251101",
            "name": "Claude Opus 4.5",
            "reasoning": false,
            "input": ["text"],
            "cost": { "input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0 },
            "contextWindow": 170000,
            "maxTokens": 64000,
            "headers": { "User-Agent": "Mozilla/5.0", "Content-Type": "application/json" }
          },
          {
            "id": "claude-sonnet-4-5-20250929",
            "name": "Claude Sonnet 4.5",
            "reasoning": false,
            "input": ["text"],
            "cost": { "input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0 },
            "contextWindow": 170000,
            "maxTokens": 64000,
            "headers": { "User-Agent": "Mozilla/5.0", "Content-Type": "application/json" }
          },
          {
            "id": "claude-opus-4-6",
            "name": "Claude Opus 4.6",
            "reasoning": false,
            "input": ["text"],
            "cost": { "input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0 },
            "contextWindow": 170000,
            "maxTokens": 64000,
            "headers": { "User-Agent": "Mozilla/5.0", "Content-Type": "application/json" }
          },
          {
            "id": "claude-haiku-4-5-20251001",
            "name": "Claude Haiku 4.5",
            "reasoning": false,
            "input": ["text"],
            "cost": { "input": 0, "output": 0, "cacheRead": 0, "cacheWrite": 0 },
            "contextWindow": 170000,
            "maxTokens": 64000,
            "headers": { "User-Agent": "Mozilla/5.0", "Content-Type": "application/json" }
          }
        ]
      }
    }
  },
  "agents": {
    "defaults": {
      "model": { "primary": "claude-sonnet-4-5-20250929" },
      "models": {
        "mayione/claude-opus-4-6": {},
        "mayione/claude-opus-4-5-20251101": {},
        "mayione/claude-sonnet-4-5-20250929": {},
        "mayione/claude-haiku-4-5-20251001": {}
      },
      "workspace": "~/.openclaw/workspace",
      "compaction": { "mode": "safeguard" },
      "maxConcurrent": 4,
      "subagents": { "maxConcurrent": 8 }
    }
  }
}`
}

const copyCode = (type: string) => {
  let code = ''
  switch (type) {
    case 'install': code = 'npm install -g openclaw@latest'; break
    case 'init': code = 'openclaw onboard --install-daemon'; break
    case 'config': code = getConfigJsonRaw(); break
    case 'switch': code = 'openclaw models set mayione/claude-opus-4-6'; break
    case 'restart': code = 'openclaw gateway restart'; break
    case 'list': code = 'openclaw models list'; break
    case 'status': code = 'openclaw models status'; break
    case 'tui': code = 'openclaw tui'; break
    default: return
  }
  navigator.clipboard.writeText(code)
  copied.value = type
  setTimeout(() => { copied.value = null }, 2000)
}

const scrollTo = (id: string) => {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' })
}

// Intersection observer for active section tracking
let observer: IntersectionObserver | null = null

onMounted(() => {
  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          activeSection.value = entry.target.id
        }
      }
    },
    { rootMargin: '-20% 0px -70% 0px' }
  )
  const ids = ['prereq', 'install', 'init', 'config', 'params', 'switch', 'restart', 'verify', 'start', 'faq']
  ids.forEach((id) => {
    const el = document.getElementById(id)
    if (el) observer!.observe(el)
  })
})

onBeforeUnmount(() => {
  observer?.disconnect()
})
</script>
