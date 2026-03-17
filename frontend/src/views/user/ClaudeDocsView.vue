<template>
  <!-- Reseller domain with doc_url: embed in iframe -->
  <div v-if="isResellerDomain && resellerDocUrl" class="flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <PublicHeader />
    <main class="flex-1">
      <iframe
        :src="resellerDocUrl"
        class="h-[calc(100vh-57px)] w-full border-0"
        allowfullscreen
        referrerpolicy="no-referrer-when-downgrade"
      ></iframe>
    </main>
  </div>

  <!-- Normal docs page -->
  <PublicLayout v-else>
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

      <!-- Prerequisites -->
      <section id="prereq" class="scroll-mt-20 mb-8 rounded-xl border border-amber-200 bg-amber-50 p-6 dark:border-amber-800 dark:bg-amber-900/20">
        <h3 class="mb-3 flex items-center gap-2 font-semibold text-amber-800 dark:text-amber-200">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          {{ t('docs.cli.prereq.title') }}
        </h3>
        <div class="space-y-3 text-sm text-amber-700 dark:text-amber-300">
          <div>
            <p class="font-medium">{{ t('docs.cli.prereq.windows') }}</p>
            <ul class="mt-1 ml-4 list-disc space-y-1">
              <li>{{ t('docs.cli.prereq.windowsNode') }} <a href="https://nodejs.org/" target="_blank" rel="noopener noreferrer" class="underline">nodejs.org</a></li>
              <li>{{ t('docs.cli.prereq.windowsTerminal') }}</li>
              <li>
                {{ t('docs.cli.prereq.windowsPsPolicy') }}
                <code class="ml-1 rounded bg-amber-100 px-1.5 py-0.5 font-mono text-xs text-amber-900 dark:bg-amber-800/40 dark:text-amber-200">{{ t('docs.cli.prereq.windowsPsPolicyCmd') }}</code>
              </li>
            </ul>
          </div>
          <div>
            <p class="font-medium">{{ t('docs.cli.prereq.mac') }}</p>
            <ul class="mt-1 ml-4 list-disc space-y-1">
              <li>{{ t('docs.cli.prereq.macTerminal') }}</li>
            </ul>
          </div>
        </div>
      </section>

      <!-- Step 1: Install Claude CLI -->
      <section id="step-1" class="scroll-mt-20 mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
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
          <span class="text-yellow-500">*</span>
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

      <!-- Step 2: Create API Key -->
      <section id="step-2" class="scroll-mt-20 mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">2</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.createKey.title') }}</h2>
        </div>

        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.guide.createKey.description') }}</p>

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

      <!-- Step 3: Configure Environment Variables -->
      <section id="step-3" class="scroll-mt-20 mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">3</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.configEnv.title') }}</h2>
        </div>

        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.guide.configEnv.description') }}</p>

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
<span class="text-blue-400">export</span> <span class="text-white">ANTHROPIC_AUTH_TOKEN=</span><span class="text-yellow-300">"{{ t('docs.guide.step2.yourKey') }}"</span></code></pre>
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
<span class="text-blue-400">$env:</span><span class="text-white">ANTHROPIC_AUTH_TOKEN</span> <span class="text-white">=</span> <span class="text-yellow-300">"{{ t('docs.guide.step2.yourKey') }}"</span></code></pre>
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
<span class="text-blue-400">set</span> <span class="text-white">ANTHROPIC_AUTH_TOKEN=</span><span class="text-yellow-300">{{ t('docs.guide.step2.yourKey') }}</span></code></pre>
          <button
            @click="copyCode('config')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'config' ? t('docs.copied') : t('docs.copy') }}
          </button>
        </div>

        <!-- Config Tip -->
        <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-yellow-500">*</span>
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
<span class="text-green-400">echo</span> <span class="text-white">$ANTHROPIC_AUTH_TOKEN</span></code></pre>
          </div>

          <!-- PowerShell -->
          <div v-else-if="activeConfigTab === 'powershell'" class="rounded bg-gray-900 p-3">
            <pre class="text-sm"><code><span class="text-gray-500">{{ t('docs.guide.step2.verifyComment') }}</span>
<span class="text-green-400">echo</span> <span class="text-white">$env:ANTHROPIC_BASE_URL</span>
<span class="text-green-400">echo</span> <span class="text-white">$env:ANTHROPIC_AUTH_TOKEN</span></code></pre>
          </div>

          <!-- CMD -->
          <div v-else class="rounded bg-gray-900 p-3">
            <pre class="text-sm"><code><span class="text-gray-500">{{ t('docs.guide.step2.verifyCMDComment') }}</span>
<span class="text-green-400">echo</span> <span class="text-white">%ANTHROPIC_BASE_URL%</span>
<span class="text-green-400">echo</span> <span class="text-white">%ANTHROPIC_AUTH_TOKEN%</span></code></pre>
          </div>

          <p class="mt-2 text-xs text-green-600 dark:text-green-400">{{ t('docs.guide.step2.verifyFail') }}</p>
        </div>

        <!-- Permanent Configuration -->
        <div class="mt-6 rounded-lg border border-purple-200 bg-purple-50 p-4 dark:border-purple-800 dark:bg-purple-900/20">
          <h3 class="mb-3 flex items-center gap-2 font-semibold text-purple-800 dark:text-purple-200">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
            </svg>
            {{ t('docs.guide.step2.permanent.title') }}
          </h3>
          <p class="mb-3 text-sm text-purple-700 dark:text-purple-300">{{ t('docs.guide.step2.permanent.desc') }}</p>

          <!-- Permanent Config Tabs -->
          <div class="mb-3 flex flex-wrap gap-2">
            <button
              v-for="tab in permanentTabs"
              :key="tab.id"
              @click="activePermanentTab = tab.id"
              :class="[
                'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                activePermanentTab === tab.id
                  ? 'bg-purple-600 text-white'
                  : 'bg-purple-200 text-purple-700 hover:bg-purple-300 dark:bg-purple-900/40 dark:text-purple-300 dark:hover:bg-purple-900/60'
              ]"
            >
              {{ tab.label }}
            </button>
          </div>

          <!-- macOS Permanent Config -->
          <div v-if="activePermanentTab === 'macos'" class="relative rounded-lg bg-gray-900 p-4">
            <div class="mb-2 text-xs text-gray-500">bash (zsh)</div>
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step2.permanent.macComment') }}</span>
<span class="text-green-400">echo</span> <span class="text-yellow-300">'export ANTHROPIC_BASE_URL="{{ baseUrl }}"'</span> <span class="text-white">>></span> <span class="text-white">~/.zshrc</span>
<span class="text-green-400">echo</span> <span class="text-yellow-300">'export ANTHROPIC_AUTH_TOKEN="{{ t('docs.guide.step2.yourKey') }}"'</span> <span class="text-white">>></span> <span class="text-white">~/.zshrc</span>
<span class="text-green-400">source</span> <span class="text-white">~/.zshrc</span></code></pre>
            <button
              @click="copyCode('permanent')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'permanent' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>

          <!-- Linux Permanent Config -->
          <div v-else-if="activePermanentTab === 'linux'" class="relative rounded-lg bg-gray-900 p-4">
            <div class="mb-2 text-xs text-gray-500">bash</div>
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step2.permanent.linuxComment') }}</span>
<span class="text-green-400">echo</span> <span class="text-yellow-300">'export ANTHROPIC_BASE_URL="{{ baseUrl }}"'</span> <span class="text-white">>></span> <span class="text-white">~/.bashrc</span>
<span class="text-green-400">echo</span> <span class="text-yellow-300">'export ANTHROPIC_AUTH_TOKEN="{{ t('docs.guide.step2.yourKey') }}"'</span> <span class="text-white">>></span> <span class="text-white">~/.bashrc</span>
<span class="text-green-400">source</span> <span class="text-white">~/.bashrc</span></code></pre>
            <button
              @click="copyCode('permanent')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'permanent' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>

          <!-- Windows PowerShell Permanent Config -->
          <div v-else class="relative rounded-lg bg-gray-900 p-4">
            <div class="mb-2 text-xs text-gray-500">powershell</div>
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">{{ t('docs.guide.step2.permanent.psComment') }}</span>
<span class="text-white">[System.Environment]::</span><span class="text-green-400">SetEnvironmentVariable</span><span class="text-white">(</span><span class="text-yellow-300">"ANTHROPIC_BASE_URL"</span><span class="text-white">,</span> <span class="text-yellow-300">"{{ baseUrl }}"</span><span class="text-white">,</span> <span class="text-white">[System.EnvironmentVariableTarget]::</span><span class="text-blue-400">User</span><span class="text-white">)</span>
<span class="text-white">[System.Environment]::</span><span class="text-green-400">SetEnvironmentVariable</span><span class="text-white">(</span><span class="text-yellow-300">"ANTHROPIC_AUTH_TOKEN"</span><span class="text-white">,</span> <span class="text-yellow-300">"{{ t('docs.guide.step2.yourKey') }}"</span><span class="text-white">,</span> <span class="text-white">[System.EnvironmentVariableTarget]::</span><span class="text-blue-400">User</span><span class="text-white">)</span></code></pre>
            <button
              @click="copyCode('permanent')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'permanent' ? t('docs.copied') : t('docs.copy') }}
            </button>
          </div>

          <p class="mt-3 text-xs text-purple-600 dark:text-purple-400">{{ t('docs.guide.step2.permanent.note') }}</p>
        </div>
      </section>

      <!-- Step 4: Start Claude -->
      <section id="step-4" class="scroll-mt-20 mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">4</span>
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
          <span class="text-yellow-500">*</span>
          {{ t('docs.guide.step3.tip') }}
        </p>

        <p class="mt-2 text-xs text-gray-400 dark:text-gray-500">
          {{ t('docs.guide.step3.debugTip') }} <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">claude --debug</code> {{ t('docs.guide.step3.debugTipSuffix') }}
        </p>
      </section>

      <!-- Step 5: VSCode Plugin -->
      <section id="step-5" class="scroll-mt-20 mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">5</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.vscodeStep.title') }}</h2>
        </div>

        <p class="mb-4 text-gray-600 dark:text-gray-400">{{ t('docs.guide.vscodeStep.description') }}</p>

        <!-- Plugin Info Card -->
        <div class="mb-6 rounded-lg border border-blue-200 bg-blue-50/50 p-4 dark:border-blue-800 dark:bg-blue-900/10">
          <div class="mb-2 flex items-center gap-2">
            <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 text-white">
              <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                <path d="M17.583 2L6.25 11.042 2.833 8.375.5 9.75l4.917 4.25L.5 18.25l2.333 1.375 3.417-2.667L17.583 26 21.5 24.25V3.75L17.583 2zm-.833 3.958v16.084l-8.917-8.042 8.917-8.042z"/>
              </svg>
            </div>
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

        <!-- Install Steps -->
        <div class="mb-6">
          <h3 class="mb-3 text-base font-medium text-gray-900 dark:text-white">{{ t('docs.vscode.step1.title') }}</h3>
          <ol class="space-y-2 text-sm text-gray-600 dark:text-gray-400">
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
        </div>

        <!-- Configure -->
        <div class="mb-6">
          <h3 class="mb-3 text-base font-medium text-gray-900 dark:text-white">{{ t('docs.vscode.step3.title') }}</h3>
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
          <div class="mt-4 rounded-lg border border-blue-300 bg-blue-100 p-3 dark:border-blue-700 dark:bg-blue-900/30">
            <p class="text-sm text-blue-700 dark:text-blue-300">
              <span class="font-medium">{{ t('docs.vscode.step3.claudeCodeTipTitle') }}</span>
              {{ t('docs.vscode.step3.claudeCodeTip') }}
            </p>
          </div>
        </div>

        <!-- Verify -->
        <div class="rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
          <p class="mb-2 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.vscode.step4.desc') }}</p>
          <p class="text-sm text-green-700 dark:text-green-300">{{ t('docs.vscode.step4.success') }}</p>
        </div>
      </section>

      <!-- Tips Section -->
      <section id="tips" class="scroll-mt-20 mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-4 flex items-center gap-2 text-xl font-semibold text-gray-900 dark:text-white">
          <svg class="h-5 w-5 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          {{ t('docs.tips.title') }}
        </h2>

        <!-- Commands Table -->
        <h3 class="mb-3 text-base font-medium text-gray-800 dark:text-gray-200">{{ t('docs.tips.commands.title') }}</h3>
        <div class="mb-6 overflow-x-auto">
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

        <!-- Shortcuts Table -->
        <h3 class="mb-3 text-base font-medium text-gray-800 dark:text-gray-200">{{ t('docs.tips.shortcuts.title') }}</h3>
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

      <!-- FAQ Section -->
      <section id="faq" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.faq.title') }}</h3>
        <div class="divide-y divide-gray-100 dark:divide-dark-700">
          <div v-for="(faq, i) in faqItems" :key="i">
            <button
              class="flex w-full items-center justify-between gap-4 py-4 text-left"
              @click="expandedFaq = expandedFaq === i ? null : i"
            >
              <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ t(faq.q) }}</span>
              <svg
                class="h-4 w-4 flex-shrink-0 text-gray-400 transition-transform duration-200"
                :class="{ 'rotate-180': expandedFaq === i }"
                fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            <div v-show="expandedFaq === i" class="pb-4 text-sm text-gray-600 dark:text-gray-400">
              {{ t(faq.a) }}
            </div>
          </div>
        </div>
      </section>

      <!-- Divider -->
      <div class="mb-6 flex items-center gap-3">
        <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
        <span class="text-xs font-semibold uppercase tracking-widest text-gray-400 dark:text-gray-500">{{ t('docs.cheatSheetDivider') }}</span>
        <div class="h-px flex-1 bg-gray-200 dark:bg-dark-700"></div>
      </div>

      <!-- §A 快捷键 -->
      <section id="cc-shortcuts" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.ccShortcuts.title') }}</h2>
        <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.guide.ccShortcuts.desc') }}</p>

        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.guide.ccShortcuts.general') }}</h3>
        <div class="mb-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">按键</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">作用</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccShortcutsGeneral" :key="row.key" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 align-top"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.key }}</code></td>
                <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.guide.ccShortcuts.textEdit') }}</h3>
        <div class="mb-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">按键</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">作用</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccShortcutsEdit" :key="row.key" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 align-top"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.key }}</code></td>
                <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.guide.ccShortcuts.multiline') }}</h3>
        <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">方式</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccMultiline" :key="row.key" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 align-top"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.key }}</code></td>
                <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- §B 斜杠命令 -->
      <section id="cc-slash" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.ccSlashCmds.title') }}</h2>
        <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.guide.ccSlashCmds.desc') }}</p>
        <div v-for="group in ccSlashGroups" :key="group.key" class="mb-5 last:mb-0">
          <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ group.label }}</h3>
          <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
            <table class="w-full text-sm">
              <thead class="bg-gray-50 dark:bg-dark-700">
                <tr>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">命令</th>
                  <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="row in group.rows" :key="row.cmd" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                  <td class="px-4 py-2.5 align-top"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-primary-600 dark:bg-dark-600 dark:text-primary-400 whitespace-nowrap">{{ row.cmd }}</code></td>
                  <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      <!-- §C 启动参数 -->
      <section id="cc-flags" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.ccCliFlags.title') }}</h2>
        <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.guide.ccCliFlags.desc') }}</p>

        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.guide.ccCliFlags.basic') }}</h3>
        <div class="mb-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">命令</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccFlagsBasic" :key="row.cmd" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 align-top"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.cmd }}</code></td>
                <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.guide.ccCliFlags.sessionMgmt') }}</h3>
        <div class="mb-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">参数</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccFlagsSession" :key="row.cmd" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 align-top"><code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.cmd }}</code></td>
                <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.guide.ccCliFlags.permission') }}</h3>
        <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">参数</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccFlagsPermission" :key="row.cmd" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 align-top">
                  <div class="flex items-center gap-2">
                    <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs font-mono text-gray-700 dark:bg-dark-600 dark:text-gray-300 whitespace-nowrap">{{ row.cmd }}</code>
                    <span v-if="row.badge" :class="row.badgeClass" class="rounded-full px-1.5 py-0.5 text-xs font-semibold whitespace-nowrap">{{ row.badge }}</span>
                  </div>
                </td>
                <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- §D CLAUDE.md -->
      <section id="cc-claude-md" class="scroll-mt-20 mb-6 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.ccClaudeMd.title') }}</h2>
        <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.guide.ccClaudeMd.desc') }}</p>
        <div class="mb-5 rounded-lg border border-primary-200 bg-primary-50 p-3 dark:border-primary-800 dark:bg-primary-900/20">
          <p class="text-sm text-primary-700 dark:text-primary-300">💡 {{ t('docs.guide.ccClaudeMd.init') }}</p>
        </div>
        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">{{ t('docs.guide.ccClaudeMd.priorityTitle') }}</h3>
        <div class="mb-4 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">级别</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">文件路径</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccClaudeMdPriority" :key="row.level" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">{{ row.level }}</td>
                <td class="px-4 py-2.5 font-mono text-xs text-gray-700 dark:text-gray-300">{{ row.path }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('docs.guide.ccClaudeMd.mergeRule') }}</p>
      </section>

      <!-- §E 权限模式 -->
      <section id="cc-permissions" class="scroll-mt-20 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h2 class="mb-1 text-xl font-semibold text-gray-900 dark:text-white">{{ t('docs.guide.ccPermissions.title') }}</h2>
        <p class="mb-5 text-sm text-gray-500 dark:text-gray-400">{{ t('docs.guide.ccPermissions.desc') }}</p>
        <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700">
              <tr>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">模式</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">启动方式</th>
                <th class="px-4 py-2.5 text-left font-medium text-gray-600 dark:text-gray-300">说明</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="row in ccPermissionModes" :key="row.mode" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="px-4 py-2.5 align-top">
                  <span :class="row.badgeClass" class="rounded-full px-2 py-0.5 text-xs font-semibold whitespace-nowrap">{{ row.mode }}</span>
                </td>
                <td class="px-4 py-2.5 align-top font-mono text-xs text-gray-600 dark:text-gray-400">{{ row.how }}</td>
                <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400">{{ row.desc }}</td>
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
import { useAppStore } from '@/stores'
import PublicLayout from '@/components/layout/PublicLayout.vue'
import PublicHeader from '@/components/layout/PublicHeader.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Reseller domain detection
const isResellerDomain = computed(() => !!appStore.cachedPublicSettings?.reseller_id)
const resellerDocUrl = computed(() => appStore.cachedPublicSettings?.doc_url || '')

// Detect user's operating system
function detectOS(): 'mac' | 'powershell' {
  const platform = navigator.platform.toLowerCase()
  const userAgent = navigator.userAgent.toLowerCase()

  if (platform.includes('win') || userAgent.includes('windows')) {
    return 'powershell'
  }
  return 'mac'
}

const copied = ref<string | null>(null)
const expandedFaq = ref<number | null>(null)
const activeInstallTab = ref('mac')
const activeConfigTab = ref('mac')
const activePermanentTab = ref('macos')

// Side navigation
const navItems = [
  { id: 'prereq', label: t('docs.cli.prereq.title') },
  { id: 'step-1', label: '1. ' + t('docs.guide.step1.title') },
  { id: 'step-2', label: '2. ' + t('docs.guide.createKey.title') },
  { id: 'step-3', label: '3. ' + t('docs.guide.configEnv.title') },
  { id: 'step-4', label: '4. ' + t('docs.guide.step3.title') },
  { id: 'step-5', label: '5. ' + t('docs.guide.vscodeStep.title') },
  { id: 'tips', label: t('docs.tips.title') },
  { id: 'faq', label: t('docs.guide.faq.title') },
  { id: 'cc-shortcuts', label: t('docs.guide.ccShortcuts.title') },
  { id: 'cc-slash', label: t('docs.guide.ccSlashCmds.title') },
  { id: 'cc-flags', label: t('docs.guide.ccCliFlags.title') },
  { id: 'cc-claude-md', label: t('docs.guide.ccClaudeMd.title') },
  { id: 'cc-permissions', label: t('docs.guide.ccPermissions.title') },
]

const activeSection = ref('')
let observer: IntersectionObserver | null = null

const scrollTo = (id: string) => {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

onMounted(() => {
  const detectedOS = detectOS()
  activeInstallTab.value = detectedOS
  activeConfigTab.value = detectedOS
  if (detectedOS === 'powershell') {
    activePermanentTab.value = 'powershell'
  } else {
    const platform = navigator.platform.toLowerCase()
    activePermanentTab.value = platform.includes('mac') ? 'macos' : 'linux'
  }

  // Scroll-spy with IntersectionObserver
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
  for (const item of navItems) {
    const el = document.getElementById(item.id)
    if (el) observer.observe(el)
  }
})

onBeforeUnmount(() => {
  observer?.disconnect()
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

const permanentTabs = [
  { id: 'macos', label: 'macOS' },
  { id: 'linux', label: 'Linux' },
  { id: 'powershell', label: 'Windows PowerShell' }
]

const baseUrl = computed(() => window.location.origin)

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
export ANTHROPIC_AUTH_TOKEN="${yourKey}"`
    case 'powershell':
      return `$env:ANTHROPIC_BASE_URL = "${baseUrl.value}"
$env:ANTHROPIC_AUTH_TOKEN = "${yourKey}"`
    case 'cmd':
      return `set ANTHROPIC_BASE_URL=${baseUrl.value}
set ANTHROPIC_AUTH_TOKEN=${yourKey}`
    default:
      return ''
  }
}

const getPermanentCode = () => {
  const yourKey = t('docs.guide.step2.yourKey')
  switch (activePermanentTab.value) {
    case 'macos':
      return `echo 'export ANTHROPIC_BASE_URL="${baseUrl.value}"' >> ~/.zshrc
echo 'export ANTHROPIC_AUTH_TOKEN="${yourKey}"' >> ~/.zshrc
source ~/.zshrc`
    case 'linux':
      return `echo 'export ANTHROPIC_BASE_URL="${baseUrl.value}"' >> ~/.bashrc
echo 'export ANTHROPIC_AUTH_TOKEN="${yourKey}"' >> ~/.bashrc
source ~/.bashrc`
    case 'powershell':
      return `[System.Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL", "${baseUrl.value}", [System.EnvironmentVariableTarget]::User)
[System.Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN", "${yourKey}", [System.EnvironmentVariableTarget]::User)`
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
  } else if (type === 'permanent') {
    code = getPermanentCode()
  } else if (type === 'start') {
    code = 'claude'
  }

  navigator.clipboard.writeText(code)
  copied.value = type
  setTimeout(() => {
    copied.value = null
  }, 2000)
}

// FAQ data
const faqItems = [
  { q: 'docs.guide.faq.q1', a: 'docs.guide.faq.a1' },
  { q: 'docs.guide.faq.q2', a: 'docs.guide.faq.a2' },
  { q: 'docs.guide.faq.q3', a: 'docs.guide.faq.a3' },
  { q: 'docs.vscode.faq.q1', a: 'docs.vscode.faq.a1' },
  { q: 'docs.vscode.faq.q2', a: 'docs.vscode.faq.a2' },
]

// Tips data
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

// ── CC Shortcuts ───────────────────────────────────────────
const ccShortcutsGeneral = [
  { key: 'Ctrl+C', desc: '取消当前输入或生成' },
  { key: 'Ctrl+D', desc: '退出 Claude Code 会话' },
  { key: 'Ctrl+L', desc: '清屏（保留对话历史）' },
  { key: 'Ctrl+G', desc: '在默认文本编辑器中打开当前输入' },
  { key: 'Ctrl+O', desc: '切换详细输出（显示工具调用详情）' },
  { key: 'Ctrl+R', desc: '反向搜索命令历史' },
  { key: 'Ctrl+V / Cmd+V', desc: '从剪贴板粘贴图片' },
  { key: 'Ctrl+B', desc: '将当前 Bash 命令移到后台运行' },
  { key: 'Ctrl+T', desc: '切换任务列表显示' },
  { key: 'Ctrl+F', desc: '终止所有后台 Agent（3 秒内按两次确认）' },
  { key: 'Esc Esc', desc: '回退对话 / 恢复到之前某个节点' },
  { key: 'Shift+Tab', desc: '循环切换权限模式（普通 → 自动接受 → 计划模式）' },
  { key: 'Option+P / Alt+P', desc: '切换模型（不清空输入框）' },
  { key: 'Option+T / Alt+T', desc: '切换扩展思考模式（需先运行 /terminal-setup）' },
  { key: '↑ / ↓', desc: '浏览命令历史' },
]

const ccShortcutsEdit = [
  { key: 'Ctrl+K', desc: '删除到行尾（保存至粘贴历史）' },
  { key: 'Ctrl+U', desc: '删除整行（保存至粘贴历史）' },
  { key: 'Ctrl+Y', desc: '粘贴已删除文本' },
  { key: 'Alt+Y', desc: '循环粘贴历史（Ctrl+Y 后使用）' },
  { key: 'Alt+B', desc: '向后移动一个单词' },
  { key: 'Alt+F', desc: '向前移动一个单词' },
]

const ccMultiline = [
  { key: '\\ + Enter', desc: '所有终端通用' },
  { key: 'Option+Enter', desc: 'macOS 默认' },
  { key: 'Shift+Enter', desc: 'iTerm2 / WezTerm / Ghostty / Kitty 开箱即用；其他终端运行 /terminal-setup' },
  { key: 'Ctrl+J', desc: '换行符方式' },
]

// ── CC Slash Commands ──────────────────────────────────────
const ccSlashGroups = computed(() => [
  {
    key: 'session',
    label: t('docs.guide.ccSlashCmds.session'),
    rows: [
      { cmd: '/compact', desc: '压缩上下文，聊太久 token 爆了用这个续命' },
      { cmd: '/clear', desc: '清空对话历史，重新开始（别名 /reset、/new）' },
      { cmd: '/fork [name]', desc: '在当前节点分叉出新对话，用于尝试不同方案' },
      { cmd: '/resume [session]', desc: '恢复指定会话或打开会话选择器（别名 /continue）' },
      { cmd: '/export [filename]', desc: '导出当前对话为纯文本' },
      { cmd: '/rewind', desc: '回退代码和对话到之前某个节点（别名 /checkpoint）' },
      { cmd: '/rename [name]', desc: '重命名当前会话，不指定名称则自动生成' },
    ],
  },
  {
    key: 'workflow',
    label: t('docs.guide.ccSlashCmds.workflow'),
    rows: [
      { cmd: '/plan', desc: '进入计划模式——只规划不执行，先看方案再动手' },
      { cmd: '/diff', desc: '查看未提交变更及每轮操作的 diff' },
      { cmd: '/cost', desc: '查看当前会话 token 用量和费用统计' },
      { cmd: '/context', desc: '以可视化方式显示当前上下文用量' },
      { cmd: '/btw <question>', desc: '快速提问，不加入对话历史（即问即答，不影响主对话）' },
      { cmd: '/tasks', desc: '列出和管理后台任务' },
    ],
  },
  {
    key: 'config',
    label: t('docs.guide.ccSlashCmds.config'),
    rows: [
      { cmd: '/model [model]', desc: '切换模型（sonnet / opus），立即生效' },
      { cmd: '/permissions', desc: '查看或更新工具权限（别名 /allowed-tools）' },
      { cmd: '/config', desc: '打开设置界面（别名 /settings）' },
      { cmd: '/theme', desc: '更换配色主题（含深色、浅色、色盲友好模式）' },
      { cmd: '/vim', desc: '切换 Vim / 普通编辑模式' },
      { cmd: '/statusline', desc: '配置底部状态栏显示内容' },
      { cmd: '/output-style', desc: '切换输出风格（默认 / 解释型 / 学习型）' },
    ],
  },
  {
    key: 'dev',
    label: t('docs.guide.ccSlashCmds.dev'),
    rows: [
      { cmd: '/init', desc: '初始化项目，生成 CLAUDE.md 框架' },
      { cmd: '/memory', desc: '编辑 CLAUDE.md 文件，管理跨会话记忆' },
      { cmd: '/mcp', desc: '管理 MCP 服务器连接' },
      { cmd: '/hooks', desc: '管理工具事件 Hook' },
      { cmd: '/skills', desc: '查看可用的 Skill 列表' },
      { cmd: '/doctor', desc: '诊断并验证 Claude Code 安装和配置' },
      { cmd: '/security-review', desc: '分析当前分支未提交变更中的安全风险' },
    ],
  },
])

// ── CC CLI Flags ───────────────────────────────────────────
const ccFlagsBasic = [
  { cmd: 'claude', desc: '启动交互式会话' },
  { cmd: 'claude "任务描述"', desc: '带初始提示启动' },
  { cmd: 'claude -p "query"', desc: '非交互式执行后退出，适合脚本 / CI' },
  { cmd: 'cat file | claude -p "query"', desc: '通过管道传入内容处理' },
  { cmd: 'claude update', desc: '更新到最新版本' },
  { cmd: 'claude --model sonnet', desc: '指定模型（sonnet / opus 或完整模型名）' },
  { cmd: 'claude -w feature-auth', desc: '在隔离的 git worktree 中启动' },
]

const ccFlagsSession = [
  { cmd: 'claude -c', desc: '恢复当前目录最近一次对话' },
  { cmd: 'claude -r "session-name"', desc: '按 ID 或名称恢复指定会话' },
  { cmd: 'claude --from-pr 123', desc: '恢复与指定 PR 关联的会话' },
  { cmd: 'claude --resume --fork-session', desc: '恢复会话并生成新的会话 ID（分叉）' },
]

const ccFlagsPermission = [
  {
    cmd: '--permission-mode plan',
    badge: '计划模式',
    badgeClass: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    desc: '只读分析，不执行任何文件操作，先看方案再动手',
  },
  {
    cmd: '--permission-mode auto-accept',
    badge: '自动接受',
    badgeClass: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    desc: '自动接受所有操作请求，无需手动确认',
  },
  {
    cmd: '--dangerously-skip-permissions',
    badge: '跳过全部',
    badgeClass: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    desc: '跳过所有权限提示，仅限容器 / CI 环境使用',
  },
  {
    cmd: '--add-dir ../lib',
    badge: null,
    badgeClass: '',
    desc: '为当前会话添加额外可访问目录（可重复使用）',
  },
  {
    cmd: '--max-turns 5',
    badge: null,
    badgeClass: '',
    desc: '限制 Agent 最大执行轮数（非交互模式）',
  },
  {
    cmd: '--max-budget-usd 2.00',
    badge: null,
    badgeClass: '',
    desc: '设置 API 调用花费上限，超出则停止（非交互模式）',
  },
]

// ── CC CLAUDE.md ───────────────────────────────────────────
const ccClaudeMdPriority = [
  { level: '全局', path: '~/.claude/CLAUDE.md' },
  { level: '父目录', path: '从根目录到项目路径各层 CLAUDE.md（逐层叠加）' },
  { level: '项目根', path: '<project>/CLAUDE.md' },
  { level: '项目子目录', path: '<project>/<subdir>/CLAUDE.md（当前工作目录中的）' },
]

// ── CC Permission Modes ────────────────────────────────────
const ccPermissionModes = [
  {
    mode: '普通模式',
    how: '默认启动',
    badgeClass: 'bg-gray-100 text-gray-600 dark:bg-dark-600 dark:text-gray-300',
    desc: '每次执行文件操作、运行命令前都会询问确认，最安全',
  },
  {
    mode: '计划模式',
    how: 'Shift+Tab 或 --permission-mode plan',
    badgeClass: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    desc: '只读分析，可查看和阅读文件，但不执行任何写操作，适合先评审方案',
  },
  {
    mode: '自动接受模式',
    how: 'Shift+Tab 或 --permission-mode auto-accept',
    badgeClass: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    desc: '自动批准所有操作，无需确认，适合信任场景下的高效执行',
  },
]
</script>
