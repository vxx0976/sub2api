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
              <span v-if="t('docs.codex.step3.configHint')" class="text-xs text-amber-600 dark:text-amber-400">— {{ t('docs.codex.step3.configHint') }}</span>
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
        <section id="faq" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.faq.title') }}</h3>
          <div class="space-y-4">
            <div v-for="(item, i) in faqItems" :key="i">
              <p class="font-medium text-gray-800 dark:text-gray-200">{{ item.q }}</p>
              <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ item.a }}</p>
            </div>
          </div>
        </section>

        <!-- Divider -->
        <div class="mb-6 flex items-center gap-3">
          <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
          <span class="text-xs font-semibold uppercase tracking-widest text-gray-400 dark:text-gray-500">{{ t('docs.cheatSheetDivider') }}</span>
          <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
        </div>

        <!-- §5 Keyboard Shortcuts -->
        <section id="shortcuts" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.shortcuts.title') }}</h2>
          <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.codex.shortcuts.desc') }}</p>

          <!-- Basic -->
          <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.codex.shortcuts.basic') }}</h3>
          <div class="mb-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">按键</th>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">作用</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in shortcutsBasic" :key="row.key" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.key }}</code></td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Edit -->
          <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.codex.shortcuts.edit') }}</h3>
          <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">按键</th>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">作用</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in shortcutsEdit" :key="row.key" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.key }}</code></td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- §6 Slash Commands -->
        <section id="slash-commands" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.slashCmds.title') }}</h2>
          <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.codex.slashCmds.desc') }}</p>

          <div v-for="group in slashGroups" :key="group.key" class="mb-5 last:mb-0">
            <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ group.label }}</h3>
            <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
              <table class="w-full text-sm">
                <thead class="bg-gray-50 dark:bg-dark-700">
                  <tr>
                    <th class="w-44 px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">命令</th>
                    <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                  <tr v-for="row in group.rows" :key="row.cmd" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                    <td class="px-4 py-2.5"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-emerald-700 dark:bg-dark-600 dark:text-emerald-400 whitespace-nowrap">{{ row.cmd }}</code></td>
                    <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </section>

        <!-- §7 Launch Parameters -->
        <section id="launch-params" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.launchParams.title') }}</h2>
          <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.codex.launchParams.desc') }}</p>

          <!-- Basic -->
          <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.codex.launchParams.basic') }}</h3>
          <div class="mb-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">命令</th>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in launchBasic" :key="row.cmd" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.cmd }}</code></td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Full-Auto -->
          <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.codex.launchParams.autoMode') }}</h3>
          <div class="mb-3 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">参数</th>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in launchAuto" :key="row.cmd" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5 align-top">
                    <div class="flex items-center gap-2">
                      <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.cmd }}</code>
                      <span v-if="row.recommended" class="rounded-full bg-emerald-100 px-1.5 py-0.5 text-xs font-semibold text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400">推荐</span>
                    </div>
                  </td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="mb-5 rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-800 dark:bg-amber-900/20">
            <p class="text-sm text-amber-700 dark:text-amber-300">{{ t('docs.codex.launchParams.autoModeNote') }}</p>
          </div>

          <!-- Sandbox -->
          <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.codex.launchParams.sandbox') }}</h3>
          <div class="mb-3 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">模式</th>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in sandboxModes" :key="row.mode" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5 align-top">
                    <div class="flex items-center gap-2">
                      <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.mode }}</code>
                      <span :class="row.badgeClass" class="rounded-full px-1.5 py-0.5 text-xs font-semibold">{{ row.badge }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="rounded-lg border border-blue-200 bg-blue-50 p-3 dark:border-blue-800 dark:bg-blue-900/20">
            <p class="text-sm text-blue-700 dark:text-blue-300">{{ t('docs.codex.launchParams.sandboxNote') }}</p>
          </div>
        </section>

        <!-- §8 AGENTS.md -->
        <section id="agents-md" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.agentsMd.title') }}</h2>
          <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.codex.agentsMd.desc') }}</p>
          <div class="mb-5 rounded-lg border border-emerald-200 bg-emerald-50 p-3 dark:border-emerald-800 dark:bg-emerald-900/20">
            <p class="text-sm text-emerald-700 dark:text-emerald-300">💡 {{ t('docs.codex.agentsMd.init') }}</p>
          </div>

          <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.codex.agentsMd.priorityTitle') }}</h3>
          <div class="mb-4 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">级别</th>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">文件路径</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in agentsPriority" :key="row.level" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5 text-gray-500 dark:text-gray-400">{{ row.level }}</td>
                  <td class="px-4 py-2.5 font-mono text-xs text-gray-700 dark:text-gray-300">{{ row.path }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('docs.codex.agentsMd.mergeRule') }}</p>
        </section>

        <!-- §9 cx vs cc -->
        <section id="compare" class="scroll-mt-20 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
          <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.codex.compare.title') }}</h2>
          <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.codex.compare.desc') }}</p>
          <div class="overflow-x-auto overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">对比项</th>
                  <th class="px-4 py-2.5 text-left font-medium text-emerald-600 dark:text-emerald-400">Codex CLI</th>
                  <th class="px-4 py-2.5 text-left font-medium text-primary-600 dark:text-primary-400">Claude Code</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in compareRows" :key="row.item" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5 font-medium text-gray-700 dark:text-gray-300">{{ row.item }}</td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">
                    <span v-if="row.cxCode"><code class="rounded bg-gray-100 px-1 py-0.5 text-xs font-mono dark:bg-dark-600">{{ row.cx }}</code></span>
                    <span v-else>{{ row.cx }}</span>
                  </td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">
                    <span v-if="row.ccCode"><code class="rounded bg-gray-100 px-1 py-0.5 text-xs font-mono dark:bg-dark-600">{{ row.cc }}</code></span>
                    <span v-else>{{ row.cc }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
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
  { id: 'shortcuts', label: t('docs.codex.shortcuts.title') },
  { id: 'slash-commands', label: t('docs.codex.slashCmds.title') },
  { id: 'launch-params', label: t('docs.codex.launchParams.title') },
  { id: 'agents-md', label: t('docs.codex.agentsMd.title') },
  { id: 'compare', label: t('docs.codex.compare.title') },
])

const faqItems = computed(() => [
  { q: t('docs.codex.faq.q1'), a: t('docs.codex.faq.a1') },
  { q: t('docs.codex.faq.q2'), a: t('docs.codex.faq.a2') },
  { q: t('docs.codex.faq.q3'), a: t('docs.codex.faq.a3') },
])

// ── Shortcuts ──────────────────────────────────────────────
const shortcutsBasic = [
  { key: 'Enter', desc: '发送消息' },
  { key: 'Esc', desc: '取消当前请求' },
  { key: 'Esc Esc', desc: '编辑上一条消息' },
  { key: 'Ctrl+C', desc: '中断当前操作' },
  { key: 'Ctrl+C Ctrl+C', desc: '强制退出' },
  { key: 'Ctrl+D', desc: '退出 Codex' },
  { key: 'Ctrl+K', desc: '清屏' },
  { key: 'Ctrl+O', desc: '选择 Codex Cloud 环境' },
]

const shortcutsEdit = [
  { key: '↑ / ↓', desc: '翻历史输入' },
  { key: 'Tab', desc: '自动补全' },
  { key: '@', desc: '引用文件（模糊搜索）' },
  { key: '!command', desc: '内联执行 Shell 命令，如 !ls、!git status' },
]

// ── Slash Commands ─────────────────────────────────────────
const slashGroups = computed(() => [
  {
    key: 'session',
    label: t('docs.codex.slashCmds.session'),
    rows: [
      { cmd: '/compact', desc: '压缩上下文，续命神器' },
      { cmd: '/diff', desc: '查看当前 Git 差异' },
      { cmd: '/review', desc: '调用另一个 Codex 代理审查代码' },
      { cmd: '/plan', desc: '进入计划模式——只规划不执行' },
      { cmd: '/fork', desc: '克隆当前对话到新线程，用于尝试不同方案' },
      { cmd: '/resume', desc: '恢复之前的对话' },
      { cmd: '/quit / /exit', desc: '退出' },
    ],
  },
  {
    key: 'config',
    label: t('docs.codex.slashCmds.config'),
    rows: [
      { cmd: '/model', desc: '切换模型或调推理级别' },
      { cmd: '/personality', desc: '切性格：friendly / pragmatic / none' },
      { cmd: '/permissions', desc: '调整权限' },
      { cmd: '/status', desc: '查看工作目录、模型、token 用量' },
      { cmd: '/experimental', desc: '开实验性功能（如 Multi-agents），改后需重启' },
    ],
  },
  {
    key: 'dev',
    label: t('docs.codex.slashCmds.dev'),
    rows: [
      { cmd: '/init', desc: '在项目里创建 AGENTS.md 框架' },
      { cmd: '/skills', desc: '浏览和插入技能' },
      { cmd: '/mcp', desc: '列出已连接的 MCP 工具' },
      { cmd: '/debug-config', desc: '查看配置加载顺序，排查配置问题' },
    ],
  },
  {
    key: 'persist',
    label: t('docs.codex.slashCmds.persist'),
    rows: [
      { cmd: '/export session.json', desc: '导出当前会话' },
      { cmd: '/load session.json', desc: '加载之前的会话继续工作' },
    ],
  },
])

// ── Launch Parameters ──────────────────────────────────────
const launchBasic = [
  { cmd: 'codex', desc: '启动交互式 TUI' },
  { cmd: 'codex "任务描述"', desc: '带任务直接启动' },
  { cmd: 'codex exec "任务"', desc: '非交互式执行，适合 CI / 脚本' },
  { cmd: 'codex resume --last', desc: '恢复最近一次对话（同 cc 的 claude -c）' },
  { cmd: 'codex fork --last', desc: '分叉最近一次对话' },
]

const launchAuto = [
  { cmd: '--full-auto', desc: '全自动执行 + 沙盒保护，日常首选', recommended: true },
  { cmd: '-a never', desc: '从不暂停等待审批，配合 -s 使用' },
  { cmd: '-a on-request', desc: '需要时才请求确认（交互式推荐）' },
  { cmd: '--yolo', desc: '禁用所有审批 + 禁用沙盒，仅限容器 / VM', recommended: false },
]

const sandboxModes = [
  {
    mode: '-s read-only',
    badge: '安全',
    badgeClass: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
    desc: '只读，不能写任何文件（含 /tmp），用于纯查看代码',
  },
  {
    mode: '-s workspace-write',
    badge: '推荐',
    badgeClass: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    desc: '可在项目目录和临时目录写入，默认禁网络，日常开发首选',
  },
  {
    mode: '-s danger-full-access',
    badge: '谨慎',
    badgeClass: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    desc: '完全放开写入和网络，系统级操作才用',
  },
]

// ── AGENTS.md ──────────────────────────────────────────────
const agentsPriority = [
  { level: '全局', path: '~/.codex/AGENTS.md' },
  { level: '全局覆盖', path: '~/.codex/AGENTS.override.md（存在时替换全局）' },
  { level: '项目根', path: '<project>/AGENTS.md' },
  { level: '项目覆盖', path: '<project>/AGENTS.override.md' },
  { level: '子目录', path: '<project>/<subdir>/AGENTS.md（逐层叠加）' },
]

// ── cx vs cc ───────────────────────────────────────────────
const compareRows = [
  { item: '开发商', cx: 'OpenAI', cc: 'Anthropic', cxCode: false, ccCode: false },
  { item: '默认模型', cx: 'gpt-5.4-codex', cc: 'Claude Sonnet / Opus', cxCode: false, ccCode: false },
  { item: '项目指令文件', cx: 'AGENTS.md', cc: 'CLAUDE.md', cxCode: true, ccCode: true },
  { item: '配置格式', cx: 'TOML', cc: 'JSON', cxCode: false, ccCode: false },
  { item: '全自动参数', cx: '--full-auto / --yolo', cc: '--dangerously-skip-permissions', cxCode: true, ccCode: true },
  { item: '内置沙盒', cx: '✓ 有（macOS: Seatbelt / Linux: Landlock）', cc: '✗ 无', cxCode: false, ccCode: false },
  { item: '恢复上次对话', cx: 'codex resume --last', cc: 'claude -c', cxCode: true, ccCode: true },
  { item: '上下文压缩', cx: '/compact', cc: '/compact', cxCode: true, ccCode: true },
  { item: 'MCP 支持', cx: '✓', cc: '✓', cxCode: false, ccCode: false },
  { item: '开源', cx: '✓ Rust', cc: '✗', cxCode: false, ccCode: false },
]

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
