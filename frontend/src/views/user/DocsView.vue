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
      <section id="faq" class="scroll-mt-20 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
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
</script>
