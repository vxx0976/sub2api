<template>
  <PublicLayout>
    <div class="mx-auto max-w-4xl">
      <!-- Header -->
      <div class="mb-8 text-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.guide.title') }}</h1>
        <p class="mt-3 text-gray-600 dark:text-gray-400">{{ t('docs.guide.subtitle') }}</p>
        <p class="mt-2 text-sm text-gray-400 dark:text-gray-500">{{ t('docs.guide.note') }}</p>
      </div>

      <!-- Warning Box -->
      <div class="mb-8 rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-900/20">
        <div class="flex items-start gap-3">
          <svg class="h-6 w-6 flex-shrink-0 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div>
            <h3 class="font-semibold text-amber-800 dark:text-amber-200">{{ t('docs.guide.warning.title') }}</h3>
            <div class="mt-2 text-sm text-amber-700 dark:text-amber-300">
              <p>{{ t('docs.guide.warning.intro') }}</p>
              <ul class="mt-1 ml-4 list-disc space-y-0.5">
                <li>{{ t('docs.guide.warning.item1') }}</li>
                <li>{{ t('docs.guide.warning.item2') }}</li>
                <li>{{ t('docs.guide.warning.item3') }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <!-- Step 1: Install Claude CLI -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">1</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.step1.title') }}</h2>
        </div>

        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.guide.step1.description') }}</p>

        <!-- OS Tabs -->
        <div class="mb-4 border-b border-gray-200 dark:border-dark-600">
          <div class="flex gap-0">
            <button
              v-for="tab in installTabs"
              :key="tab.id"
              @click="activeInstallTab = tab.id"
              :class="[
                'px-4 py-2 text-sm font-medium border-b-2 transition-colors',
                activeInstallTab === tab.id
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
              ]"
            >
              {{ tab.label }}
            </button>
          </div>
        </div>

        <!-- Install Code Block - macOS/Linux -->
        <div v-if="activeInstallTab === 'mac'" class="relative rounded-lg bg-gray-900 p-4">
          <div class="mb-2 text-xs text-gray-500">bash</div>
          <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step1.commentBash') }}</span>
<span class="text-green-400">curl</span> <span class="text-white">-fsSL</span> <span class="text-yellow-300">https://claude.ai/install.sh</span> <span class="text-white">|</span> <span class="text-green-400">bash</span></code></pre>
          <button
            @click="copyCode('install')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'install' ? t('docs.copied') : t('docs.copy') }}
          </button>
        </div>

        <!-- Install Code Block - Windows PowerShell -->
        <div v-else-if="activeInstallTab === 'powershell'" class="relative rounded-lg bg-gray-900 p-4">
          <div class="mb-2 text-xs text-gray-500">powershell</div>
          <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step1.commentPS') }}</span>
<span class="text-green-400">irm</span> <span class="text-yellow-300">https://claude.ai/install.ps1</span> <span class="text-white">|</span> <span class="text-green-400">iex</span></code></pre>
          <button
            @click="copyCode('install')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'install' ? t('docs.copied') : t('docs.copy') }}
          </button>
        </div>

        <!-- Install Code Block - Windows CMD -->
        <div v-else class="relative rounded-lg bg-gray-900 p-4">
          <div class="mb-2 text-xs text-gray-500">cmd</div>
          <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step1.commentCMD') }}</span>
<span class="text-green-400">npm</span> <span class="text-white">install -g</span> <span class="text-yellow-300">@anthropic-ai/claude-code</span></code></pre>
          <button
            @click="copyCode('install')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'install' ? t('docs.copied') : t('docs.copy') }}
          </button>
        </div>

        <!-- Install Tip -->
        <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-yellow-500">💡</span>
          <span v-if="activeInstallTab === 'mac'">
            {{ t('docs.guide.step1.tipMac') }}
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">source ~/.bashrc</code>
          </span>
          <span v-else-if="activeInstallTab === 'powershell'">
            {{ t('docs.guide.step1.tipPS') }}
          </span>
          <span v-else>
            {{ t('docs.guide.step1.tipCMD') }}
          </span>
        </p>

        <!-- Verify Installation -->
        <div class="mt-4 rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
          <div class="mb-2 text-sm font-medium text-green-700 dark:text-green-300">{{ t('docs.guide.step1.verify') }}</div>
          <div class="rounded bg-gray-900 p-3">
            <code class="text-sm text-green-400">claude --version</code>
          </div>
          <p class="mt-2 text-xs text-green-600 dark:text-green-400">{{ t('docs.guide.step1.verifySuccess') }}</p>
        </div>

        <!-- Network Tip -->
        <p class="mt-3 text-xs text-gray-400 dark:text-gray-500">
          {{ t('docs.guide.step1.networkTip') }}
        </p>
      </section>

      <!-- Step 2: Create API Key and Configure -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">2</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.step2.title') }}</h2>
        </div>

        <ol class="mb-6 space-y-3 text-gray-700 dark:text-gray-300">
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">1.</span>
            <span>{{ t('docs.guide.step2.instruction1') }} <router-link to="/keys" class="text-primary-500 hover:underline">{{ t('docs.guide.step2.instruction1Link') }}</router-link></span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">2.</span>
            <span>{{ t('docs.guide.step2.instruction2') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">3.</span>
            <span>{{ t('docs.guide.step2.instruction3Pre') }} <span class="inline-block rounded bg-primary-500 px-2 py-0.5 text-xs text-white">{{ t('docs.guide.step2.instruction3Button') }}</span> {{ t('docs.guide.step2.instruction3Post') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">4.</span>
            <span>{{ t('docs.guide.step2.instruction4') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">5.</span>
            <span>{{ t('docs.guide.step2.instruction5') }}</span>
          </li>
        </ol>

        <!-- Example Code Section -->
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-700/50">
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.guide.step2.exampleTitle') }}</p>

          <!-- OS Tabs for Config -->
          <div class="mb-3 flex flex-wrap gap-2">
            <button
              v-for="tab in configTabs"
              :key="tab.id"
              @click="activeConfigTab = tab.id"
              :class="[
                'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                activeConfigTab === tab.id
                  ? 'bg-primary-500 text-white'
                  : 'bg-gray-200 text-gray-700 hover:bg-gray-300 dark:bg-dark-600 dark:text-gray-300 dark:hover:bg-dark-500'
              ]"
            >
              {{ tab.label }}
            </button>
          </div>

          <!-- Config Code Block - macOS/Linux -->
          <div v-if="activeConfigTab === 'mac'" class="relative rounded-lg bg-gray-900 p-4">
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step2.commentBash') }}</span>
<span class="text-blue-400">export</span> <span class="text-white">ANTHROPIC_BASE_URL=</span><span class="text-yellow-300">"{{ baseUrl }}"</span>
<span class="text-blue-400">export</span> <span class="text-white">ANTHROPIC_API_KEY=</span><span class="text-yellow-300">"{{ t('docs.guide.step2.yourKey') }}"</span></code></pre>
            <button
              @click="copyCode('config')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'config' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>

          <!-- Config Code Block - Windows PowerShell -->
          <div v-else-if="activeConfigTab === 'powershell'" class="relative rounded-lg bg-gray-900 p-4">
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step2.commentPS') }}</span>
<span class="text-blue-400">$env:</span><span class="text-white">ANTHROPIC_BASE_URL</span> <span class="text-white">=</span> <span class="text-yellow-300">"{{ baseUrl }}"</span>
<span class="text-blue-400">$env:</span><span class="text-white">ANTHROPIC_API_KEY</span> <span class="text-white">=</span> <span class="text-yellow-300">"{{ t('docs.guide.step2.yourKey') }}"</span></code></pre>
            <button
              @click="copyCode('config')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'config' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>

          <!-- Config Code Block - Windows CMD -->
          <div v-else class="relative rounded-lg bg-gray-900 p-4">
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step2.commentCMD') }}</span>
<span class="text-blue-400">set</span> <span class="text-white">ANTHROPIC_BASE_URL=</span><span class="text-yellow-300">{{ baseUrl }}</span>
<span class="text-blue-400">set</span> <span class="text-white">ANTHROPIC_API_KEY=</span><span class="text-yellow-300">{{ t('docs.guide.step2.yourKey') }}</span></code></pre>
            <button
              @click="copyCode('config')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'config' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>
        </div>

        <!-- Config Tip -->
        <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-yellow-500">💡</span>
          <span v-if="activeConfigTab === 'mac'">
            {{ t('docs.guide.step2.tipMac') }}
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">~/.zshrc</code>
            {{ t('docs.guide.step2.tipMacOr') }}
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">~/.bashrc</code>
            {{ t('docs.guide.step2.tipMacSuffix') }}
          </span>
          <span v-else-if="activeConfigTab === 'powershell'">
            {{ t('docs.guide.step2.tipPS') }}
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">$PROFILE</code>
            {{ t('docs.guide.step2.tipPSSuffix') }}
          </span>
          <span v-else>
            {{ t('docs.guide.step2.tipCMD') }}
          </span>
        </p>

        <!-- Verify Config -->
        <div class="mt-4 rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
          <div class="mb-2 text-sm font-medium text-green-700 dark:text-green-300">{{ t('docs.guide.step2.verify') }}</div>

          <!-- macOS/Linux -->
          <div v-if="activeConfigTab === 'mac'" class="rounded bg-gray-900 p-3">
            <pre class="text-sm"><code><span class="text-gray-500">{{ t('docs.guide.step2.verifyComment') }}</span>
<span class="text-green-400">echo</span> <span class="text-white">$ANTHROPIC_BASE_URL</span>
<span class="text-green-400">echo</span> <span class="text-white">$ANTHROPIC_API_KEY</span></code></pre>
          </div>

          <!-- PowerShell -->
          <div v-else-if="activeConfigTab === 'powershell'" class="rounded bg-gray-900 p-3">
            <pre class="text-sm"><code><span class="text-gray-500">{{ t('docs.guide.step2.verifyComment') }}</span>
<span class="text-green-400">echo</span> <span class="text-white">$env:ANTHROPIC_BASE_URL</span>
<span class="text-green-400">echo</span> <span class="text-white">$env:ANTHROPIC_API_KEY</span></code></pre>
          </div>

          <!-- CMD -->
          <div v-else class="rounded bg-gray-900 p-3">
            <pre class="text-sm"><code><span class="text-gray-500">{{ t('docs.guide.step2.verifyCMDComment') }}</span>
<span class="text-green-400">echo</span> <span class="text-white">%ANTHROPIC_BASE_URL%</span>
<span class="text-green-400">echo</span> <span class="text-white">%ANTHROPIC_API_KEY%</span></code></pre>
          </div>

          <p class="mt-2 text-xs text-green-600 dark:text-green-400">{{ t('docs.guide.step2.verifyFail') }}</p>
        </div>
      </section>

      <!-- Step 3: Start Claude -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">3</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.step3.title') }}</h2>
        </div>

        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.guide.step3.description') }}</p>

        <div class="relative rounded-lg border border-gray-200 bg-gray-900 p-4 dark:border-dark-600">
          <code class="text-lg text-green-400">claude</code>
          <button
            @click="copyCode('start')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'start' ? t('docs.copied') : t('docs.copy') }}
          </button>
        </div>

        <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-yellow-500">💡</span>
          {{ t('docs.guide.step3.tip') }}
        </p>

        <p class="mt-2 text-xs text-gray-400 dark:text-gray-500">
          {{ t('docs.guide.step3.debugTip') }} <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">claude --debug</code> {{ t('docs.guide.step3.debugTipSuffix') }}
        </p>
      </section>

      <!-- Tips Box -->
      <section class="rounded-xl border border-blue-200 bg-blue-50 p-6 dark:border-blue-800 dark:bg-blue-900/20">
        <h3 class="mb-4 flex items-center gap-2 font-semibold text-blue-800 dark:text-blue-200">
          <span class="text-yellow-500">💡</span>
          {{ t('docs.guide.tips.title') }}
        </h3>
        <ul class="space-y-2 text-sm text-blue-700 dark:text-blue-300">
          <li class="flex items-start gap-2">
            <span>•</span>
            <span>{{ t('docs.guide.tips.tip1') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span>•</span>
            <span>{{ t('docs.guide.tips.tip2') }}</span>
          </li>
          <li class="flex items-start gap-2">
            <span>•</span>
            <span>{{ t('docs.guide.tips.tip3Pre') }} <router-link to="/dashboard" class="font-medium underline">{{ t('docs.guide.tips.tip3Link') }}</router-link> {{ t('docs.guide.tips.tip3Post') }}</span>
          </li>
        </ul>
      </section>

      <!-- FAQ Section -->
      <section class="mt-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.faq.title') }}</h3>
        <div class="space-y-4">
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.guide.faq.q1') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.guide.faq.a1') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.guide.faq.q2') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.guide.faq.a2') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.guide.faq.q3') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.guide.faq.a3') }}</p>
          </div>
        </div>
      </section>
    </div>
  </PublicLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicLayout from '@/components/layout/PublicLayout.vue'

const { t } = useI18n()

// Detect user's operating system
function detectOS(): 'mac' | 'powershell' {
  const platform = navigator.platform.toLowerCase()
  const userAgent = navigator.userAgent.toLowerCase()

  if (platform.includes('win') || userAgent.includes('windows')) {
    return 'powershell'
  }
  // macOS, Linux, and others default to mac/bash
  return 'mac'
}

const copied = ref<string | null>(null)
const activeInstallTab = ref('mac')
const activeConfigTab = ref('mac')

onMounted(() => {
  const detectedOS = detectOS()
  activeInstallTab.value = detectedOS
  activeConfigTab.value = detectedOS
})

const installTabs = [
  { id: 'mac', label: 'macOS / Linux / WSL' },
  { id: 'powershell', label: 'Windows PowerShell' },
  { id: 'cmd', label: 'Windows CMD' }
]

const configTabs = [
  { id: 'mac', label: 'macOS / Linux' },
  { id: 'powershell', label: 'Windows PowerShell' },
  { id: 'cmd', label: 'Windows CMD' }
]

const baseUrl = computed(() => {
  return window.location.origin
})

const getInstallCode = () => {
  switch (activeInstallTab.value) {
    case 'mac':
      return 'curl -fsSL https://claude.ai/install.sh | bash'
    case 'powershell':
      return 'irm https://claude.ai/install.ps1 | iex'
    case 'cmd':
      return 'npm install -g @anthropic-ai/claude-code'
    default:
      return ''
  }
}

const getConfigCode = () => {
  const yourKey = t('docs.guide.step2.yourKey')
  switch (activeConfigTab.value) {
    case 'mac':
      return `export ANTHROPIC_BASE_URL="${baseUrl.value}"
export ANTHROPIC_API_KEY="${yourKey}"`
    case 'powershell':
      return `$env:ANTHROPIC_BASE_URL = "${baseUrl.value}"
$env:ANTHROPIC_API_KEY = "${yourKey}"`
    case 'cmd':
      return `set ANTHROPIC_BASE_URL=${baseUrl.value}
set ANTHROPIC_API_KEY=${yourKey}`
    default:
      return ''
  }
}

const copyCode = (type: string) => {
  let code = ''
  if (type === 'install') {
    code = getInstallCode()
  } else if (type === 'config') {
    code = getConfigCode()
  } else if (type === 'start') {
    code = 'claude'
  }

  navigator.clipboard.writeText(code)
  copied.value = type
  setTimeout(() => {
    copied.value = null
  }, 2000)
}
</script>
