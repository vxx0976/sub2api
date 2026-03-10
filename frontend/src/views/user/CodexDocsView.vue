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
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.codex.title') }}</h1>
          <p class="mt-3 text-gray-600 dark:text-gray-400">{{ t('docs.codex.subtitle') }}</p>
        </div>

        <!-- Step 1: Install Codex CLI -->
        <section id="step-1" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4 flex items-center gap-3">
            <span class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500 text-sm font-bold text-white">1</span>
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.step1.title') }}</h2>
          </div>
          <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.codex.step1.description') }}</p>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <div class="mb-2 text-xs text-gray-500">bash</div>
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-green-400">npm</span> <span class="text-white">install -g</span> <span class="text-yellow-300">@openai/codex</span></code></pre>
            <button
              @click="copyText('npm install -g @openai/codex', 'install')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'install' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>
          <div class="mt-4 rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
            <div class="mb-2 text-sm font-medium text-green-700 dark:text-green-300">{{ t('docs.codex.step1.verify') }}</div>
            <div class="rounded bg-gray-900 p-3">
              <code class="text-sm text-green-400">codex --version</code>
            </div>
            <p class="mt-2 text-xs text-green-600 dark:text-green-400">{{ t('docs.codex.step1.verifySuccess') }}</p>
          </div>
        </section>

        <!-- Step 2: Create API Key -->
        <section id="step-2" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4 flex items-center gap-3">
            <span class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500 text-sm font-bold text-white">2</span>
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.step2.title') }}</h2>
          </div>
          <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.codex.step2.description') }}</p>
          <ol class="space-y-3 text-gray-700 dark:text-gray-300">
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
          </ol>
        </section>

        <!-- Step 3: Configure Config Files -->
        <section id="step-3" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4 flex items-center gap-3">
            <span class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500 text-sm font-bold text-white">3</span>
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.step3.title') }}</h2>
          </div>
          <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.codex.step3.description') }}</p>

          <!-- Connection Mode Tabs -->
          <div class="mb-4 border-b border-gray-200 dark:border-dark-700">
            <nav class="-mb-px flex space-x-6">
              <button
                v-for="mode in modes"
                :key="mode.id"
                @click="activeMode = mode.id"
                :class="[
                  'whitespace-nowrap pb-2.5 px-1 border-b-2 font-medium text-sm transition-colors',
                  activeMode === mode.id
                    ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
                ]"
              >
                {{ mode.label }}
              </button>
            </nav>
          </div>

          <!-- WebSocket mode note -->
          <div v-if="activeMode === 'ws'" class="mb-4 rounded-lg border border-blue-200 bg-blue-50 p-3 dark:border-blue-800 dark:bg-blue-900/20">
            <p class="text-sm text-blue-700 dark:text-blue-300">{{ t('docs.codex.step3.wsNote') }}</p>
          </div>

          <!-- OS Tabs -->
          <div class="mb-4 border-b border-gray-200 dark:border-dark-700">
            <nav class="-mb-px flex space-x-4">
              <button
                v-for="os in osTabs"
                :key="os.id"
                @click="activeOS = os.id"
                :class="[
                  'whitespace-nowrap pb-2.5 px-1 border-b-2 font-medium text-sm transition-colors',
                  activeOS === os.id
                    ? 'border-emerald-500 text-emerald-600 dark:text-emerald-400'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
                ]"
              >
                {{ os.label }}
              </button>
            </nav>
          </div>

          <!-- Config dir hint -->
          <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">
            {{ t('docs.codex.step3.configDirLabel') }}: <code class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-700 dark:bg-dark-700 dark:text-gray-300">{{ configDir }}</code>
          </p>

          <!-- File 1: config.toml -->
          <div class="mb-4">
            <div class="mb-2 flex items-center gap-2">
              <code class="rounded bg-gray-100 px-2 py-0.5 text-sm font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-300">{{ configDir }}/config.toml</code>
              <span v-if="configHint" class="text-xs text-amber-600 dark:text-amber-400">— {{ configHint }}</span>
            </div>
            <div class="relative rounded-lg bg-gray-900">
              <div class="flex items-center justify-between px-4 py-2 bg-gray-800 rounded-t-lg border-b border-gray-700">
                <span class="text-xs text-gray-400 font-mono">config.toml</span>
                <button
                  @click="copyText(configToml, 'config')"
                  class="flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-lg transition-colors"
                  :class="copied === 'config' ? 'bg-green-500/20 text-green-400' : 'bg-gray-700 hover:bg-gray-600 text-gray-300'"
                >
                  {{ copied === 'config' ? t('docs.copied') : t('docs.copy') }}
                </button>
              </div>
              <pre class="p-4 text-sm font-mono text-gray-100 overflow-x-auto leading-relaxed">{{ configToml }}</pre>
            </div>
          </div>

          <!-- File 2: auth.json -->
          <div class="mb-4">
            <div class="mb-2 flex items-center gap-2">
              <code class="rounded bg-gray-100 px-2 py-0.5 text-sm font-medium text-gray-700 dark:bg-dark-700 dark:text-gray-300">{{ configDir }}/auth.json</code>
            </div>
            <div class="relative rounded-lg bg-gray-900">
              <div class="flex items-center justify-between px-4 py-2 bg-gray-800 rounded-t-lg border-b border-gray-700">
                <span class="text-xs text-gray-400 font-mono">auth.json</span>
                <button
                  @click="copyText(authJson, 'auth')"
                  class="flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-lg transition-colors"
                  :class="copied === 'auth' ? 'bg-green-500/20 text-green-400' : 'bg-gray-700 hover:bg-gray-600 text-gray-300'"
                >
                  {{ copied === 'auth' ? t('docs.copied') : t('docs.copy') }}
                </button>
              </div>
              <pre class="p-4 text-sm font-mono text-gray-100 overflow-x-auto leading-relaxed">{{ authJson }}</pre>
            </div>
          </div>

          <!-- Dir note -->
          <div class="rounded-lg border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-900/20">
            <p class="text-sm text-blue-700 dark:text-blue-300">{{ activeOS === 'windows' ? t('docs.codex.step3.noteWindows') : t('docs.codex.step3.note') }}</p>
          </div>
        </section>

        <!-- Step 4: Start Codex -->
        <section id="step-4" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <div class="mb-4 flex items-center gap-3">
            <span class="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500 text-sm font-bold text-white">4</span>
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.step4.title') }}</h2>
          </div>
          <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.codex.step4.description') }}</p>
          <div class="relative rounded-lg bg-gray-900 p-4">
            <code class="text-lg text-green-400">codex</code>
            <button
              @click="copyText('codex', 'start')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'start' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>
          <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
            <span class="text-yellow-500">*</span>
            {{ t('docs.codex.step4.tip') }}
          </p>
          <p class="mt-2 text-xs text-gray-400 dark:text-gray-500">{{ t('docs.codex.step4.debugTip') }}</p>
        </section>

        <!-- FAQ Section -->
        <section id="faq" class="scroll-mt-20 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.faq.title') }}</h3>
          <div class="space-y-4">
            <div v-for="(item, i) in faqItems" :key="i">
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ item.q }}</p>
              <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ item.a }}</p>
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
const baseUrl = computed(() => window.location.origin)

type ModeId = 'standard' | 'ws'
type OSId = 'unix' | 'windows'

const activeMode = ref<ModeId>('standard')
const activeOS = ref<OSId>('unix')

const modes = computed(() => [
  { id: 'standard' as ModeId, label: t('docs.codex.step3.modeStandard') },
  { id: 'ws' as ModeId, label: t('docs.codex.step3.modeWs') },
])

const osTabs = computed(() => [
  { id: 'unix' as OSId, label: t('docs.codex.step3.osMacLinux') },
  { id: 'windows' as OSId, label: t('docs.codex.step3.osWindows') },
])

const configDir = computed(() =>
  activeOS.value === 'windows' ? '%userprofile%\\.codex' : '~/.codex'
)

const configHint = computed(() => t('docs.codex.step3.configHint'))

const configToml = computed(() => {
  const base = baseUrl.value
  const common = `model_provider = "OpenAI"
model = "gpt-5.4"
review_model = "gpt-5.4"
model_reasoning_effort = "xhigh"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true
model_context_window = 1000000
model_auto_compact_token_limit = 900000`

  if (activeMode.value === 'ws') {
    return `${common}

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${base}"
wire_api = "responses"
supports_websockets = true
requires_openai_auth = true

[features]
responses_websockets_v2 = true`
  }

  return `${common}

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${base}"
wire_api = "responses"
requires_openai_auth = true`
})

const authJson = computed(() => {
  const placeholder = t('docs.guide.step2.yourKey')
  return `{
  "OPENAI_API_KEY": "${placeholder}"
}`
})

const navItems = computed(() => [
  { id: 'step-1', label: '1. ' + t('docs.codex.step1.title') },
  { id: 'step-2', label: '2. ' + t('docs.codex.step2.title') },
  { id: 'step-3', label: '3. ' + t('docs.codex.step3.title') },
  { id: 'step-4', label: '4. ' + t('docs.codex.step4.title') },
  { id: 'faq', label: t('docs.codex.faq.title') },
])

const faqItems = computed(() => [
  { q: t('docs.codex.faq.q1'), a: t('docs.codex.faq.a1') },
  { q: t('docs.codex.faq.q2'), a: t('docs.codex.faq.a2') },
  { q: t('docs.codex.faq.q3'), a: t('docs.codex.faq.a3') },
])

const activeSection = ref('')
let observer: IntersectionObserver | null = null

const scrollTo = (id: string) => {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

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
  for (const item of navItems.value) {
    const el = document.getElementById(item.id)
    if (el) observer.observe(el)
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
})

const copyText = async (text: string, key: string) => {
  await navigator.clipboard.writeText(text)
  copied.value = key
  setTimeout(() => { copied.value = null }, 2000)
}
</script>
