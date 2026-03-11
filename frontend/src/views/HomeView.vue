<template>
  <!-- External URL: iframe mode -->
  <div v-if="homeTemplate === 'external_url' && homeContent" class="min-h-screen">
    <iframe
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
  </div>

  <!-- Custom HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
  <div v-else-if="homeTemplate === 'custom_html' && homeContent" class="min-h-screen" v-html="homeContent"></div>

  <!-- Hero template: gradient hero + features + CTA -->
  <div v-else-if="homeTemplate === 'hero'" class="relative flex min-h-screen flex-col overflow-hidden">
    <!-- Gradient background -->
    <div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-slate-900 via-primary-950 to-slate-900"></div>
    <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_80%_50%_at_50%_-20%,rgba(120,119,198,0.25),transparent)]"></div>
    <PublicHeader class="relative z-10" />
    <main class="relative z-10 flex flex-1 flex-col items-center justify-center px-6 py-20 text-center">
      <div v-if="appStore.siteLogo" class="mb-8 flex justify-center">
        <img :src="appStore.siteLogo" :alt="siteName" class="h-20 w-20 rounded-2xl object-contain shadow-2xl ring-1 ring-white/10" />
      </div>
      <h1 class="mx-auto max-w-3xl text-4xl font-bold tracking-tight text-white sm:text-6xl">
        {{ siteName }}
      </h1>
      <p v-if="appStore.cachedPublicSettings?.site_subtitle" class="mx-auto mt-5 max-w-xl text-lg text-slate-300">
        {{ appStore.cachedPublicSettings.site_subtitle }}
      </p>
      <div class="mt-10 flex flex-wrap justify-center gap-3">
        <router-link :to="isAuthenticated ? dashboardPath : loginPath" class="inline-flex items-center gap-2 rounded-xl bg-white px-7 py-3 text-sm font-semibold text-slate-900 shadow-lg transition hover:bg-slate-100">
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
          <Icon name="arrowRight" size="sm" :stroke-width="2.5" />
        </router-link>
        <router-link v-if="isResellerDomain && docUrl" to="/docs" class="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/10 px-7 py-3 text-sm font-semibold text-white backdrop-blur transition hover:bg-white/20">
          <Icon name="book" size="sm" />
          {{ t('home.docs') }}
        </router-link>
        <a v-else-if="!isResellerDomain && docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/10 px-7 py-3 text-sm font-semibold text-white backdrop-blur transition hover:bg-white/20">
          <Icon name="book" size="sm" />
          {{ t('home.docs') }}
        </a>
        <router-link v-if="isResellerDomain" to="/key-query" class="inline-flex items-center gap-2 rounded-xl border border-white/20 bg-white/10 px-7 py-3 text-sm font-semibold text-white backdrop-blur transition hover:bg-white/20">
          <Icon name="search" size="sm" />
          {{ t('keyQuery.title') }}
        </router-link>
      </div>
      <!-- Feature pills -->
      <div class="mt-16 flex flex-wrap justify-center gap-3">
        <span v-for="feat in ['OpenAI Compatible', 'Claude / Gemini / GPT / Codex', '高并发', '低延迟', '按量计费']" :key="feat"
          class="rounded-full border border-white/10 bg-white/5 px-4 py-1.5 text-xs font-medium text-slate-300 backdrop-blur">
          {{ feat }}
        </span>
      </div>
    </main>
    <footer class="relative z-10 py-6 text-center text-sm text-slate-500">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <!-- Brand template: logo + contact + clean CTA -->
  <div v-else-if="homeTemplate === 'brand'" class="relative flex min-h-screen flex-col bg-white dark:bg-dark-950">
    <PublicHeader class="relative z-10" />
    <main class="flex flex-1 flex-col items-center justify-center px-6 py-20">
      <div class="mx-auto w-full max-w-md text-center">
        <div v-if="appStore.siteLogo" class="mb-8 flex justify-center">
          <img :src="appStore.siteLogo" :alt="siteName" class="h-24 w-24 rounded-3xl object-contain shadow-xl" />
        </div>
        <div v-else class="mb-8 flex justify-center">
          <div class="flex h-24 w-24 items-center justify-center rounded-3xl bg-primary-600 shadow-xl">
            <span class="text-3xl font-bold text-white">{{ siteName.charAt(0) }}</span>
          </div>
        </div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ siteName }}</h1>
        <p v-if="appStore.cachedPublicSettings?.site_subtitle" class="mt-3 text-base text-gray-500 dark:text-gray-400">
          {{ appStore.cachedPublicSettings.site_subtitle }}
        </p>
        <router-link :to="isAuthenticated ? dashboardPath : loginPath" class="mt-8 inline-flex w-full items-center justify-center gap-2 rounded-2xl bg-primary-600 py-3.5 text-sm font-semibold text-white shadow-lg transition hover:bg-primary-700">
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
          <Icon name="arrowRight" size="sm" :stroke-width="2.5" />
        </router-link>
        <div class="mt-4 flex flex-col items-center gap-2">
          <router-link v-if="isResellerDomain" to="/key-query" class="text-sm text-gray-500 transition hover:text-gray-700 dark:text-gray-400 dark:hover:text-white">
            {{ t('keyQuery.title') }}
          </router-link>
          <router-link v-if="isResellerDomain && docUrl" to="/docs" class="text-sm text-gray-500 transition hover:text-gray-700 dark:text-gray-400 dark:hover:text-white">
            {{ t('home.docs') }}
          </router-link>
          <a v-else-if="!isResellerDomain && docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="text-sm text-gray-500 transition hover:text-gray-700 dark:text-gray-400 dark:hover:text-white">
            {{ t('home.docs') }}
          </a>
        </div>
        <!-- Contact info -->
        <div v-if="appStore.cachedPublicSettings?.contact_info" class="mt-8 rounded-xl border border-gray-100 bg-gray-50 p-4 text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400">
          {{ appStore.cachedPublicSettings.contact_info }}
        </div>
      </div>
    </main>
    <footer class="py-6 text-center text-sm text-gray-400 dark:text-gray-600">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <!-- Minimal template -->
  <div v-else-if="homeTemplate === 'minimal'" class="relative flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <PublicHeader />
    <main class="relative z-10 flex flex-1 items-center justify-center">
      <div class="mx-auto max-w-lg px-6 py-16 text-center">
        <div v-if="appStore.siteLogo" class="mb-6 flex justify-center">
          <img :src="appStore.siteLogo" :alt="siteName" class="h-16 w-16 rounded-2xl object-contain" />
        </div>
        <h1 class="text-3xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
          {{ siteName }}
        </h1>
        <p v-if="appStore.cachedPublicSettings?.site_subtitle" class="mt-3 text-base text-gray-600 dark:text-dark-400">
          {{ appStore.cachedPublicSettings.site_subtitle }}
        </p>
        <div class="mt-8 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <router-link :to="isAuthenticated ? dashboardPath : loginPath" class="btn btn-primary btn-lg px-8">
            {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
            <Icon name="arrowRight" size="md" :stroke-width="2" />
          </router-link>
          <router-link v-if="isResellerDomain" to="/key-query" class="btn btn-secondary btn-lg px-8">
            <Icon name="search" size="sm" />
            {{ t('keyQuery.title') }}
          </router-link>
          <router-link
            v-if="isResellerDomain && docUrl"
            to="/docs"
            class="btn btn-secondary btn-lg px-8"
          >
            <Icon name="book" size="sm" />
            {{ t('home.docs') }}
          </router-link>
          <a
            v-else-if="!isResellerDomain && docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-lg px-8"
          >
            <Icon name="book" size="sm" />
            {{ t('home.docs') }}
          </a>
        </div>
      </div>
    </main>
    <footer class="relative z-10 border-t border-gray-200/60 px-6 py-6 dark:border-dark-800/60">
      <p class="text-center text-sm text-gray-500 dark:text-dark-400">
        &copy; {{ currentYear }} {{ siteName }}
      </p>
    </footer>
  </div>

  <!-- Default Home Page -->
  <div v-else class="relative flex min-h-screen flex-col overflow-hidden bg-gray-50 dark:bg-dark-950">
    <!-- Background (light) -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute inset-0 bg-mesh-gradient opacity-60 dark:opacity-30"></div>
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(15,23,42,0.04)_1px,transparent_1px),linear-gradient(90deg,rgba(15,23,42,0.04)_1px,transparent_1px)] bg-[size:72px_72px] dark:bg-[linear-gradient(rgba(148,163,184,0.06)_1px,transparent_1px),linear-gradient(90deg,rgba(148,163,184,0.06)_1px,transparent_1px)]"
      ></div>
    </div>

    <!-- Header -->
    <PublicHeader />

    <!-- Main Content -->
    <main class="relative z-10 flex-1">
      <!-- Hero -->
      <section class="relative overflow-hidden">
        <div class="pointer-events-none absolute inset-0">
          <div
            class="absolute -top-24 left-1/2 h-[28rem] w-[50rem] -translate-x-1/2 rounded-full bg-primary-500/15 blur-3xl"
          ></div>
          <div class="absolute -top-32 right-[-12rem] h-80 w-80 rounded-full bg-primary-400/10 blur-3xl"></div>
          <div class="absolute -bottom-36 left-[-12rem] h-80 w-80 rounded-full bg-primary-600/15 blur-3xl"></div>
        </div>

        <div class="relative mx-auto max-w-7xl px-6 pb-12 pt-14 lg:pb-16 lg:pt-20">
          <div class="grid gap-10 lg:grid-cols-2 lg:items-center">
            <!-- Copy -->
            <div class="text-center lg:text-left">
              <div
                class="inline-flex items-center gap-2 rounded-full border border-gray-200/60 bg-white/70 px-3 py-1 text-xs font-semibold text-gray-700 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40 dark:text-dark-200"
              >
                <Icon name="sparkles" size="sm" class="text-primary-600 dark:text-primary-300" />
                <span>{{ t('home.hero.badge') }}</span>
              </div>

              <h1
                class="mt-6 text-balance text-4xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-5xl lg:text-6xl"
              >
                {{ t('home.hero.title', { siteName }) }}
              </h1>

              <p class="mt-4 max-w-2xl text-pretty text-lg text-gray-600 dark:text-dark-400 sm:text-xl">
                {{ t('home.hero.subtitle') }}
              </p>

              <p class="mt-2 max-w-2xl text-pretty text-base text-gray-500 dark:text-dark-500">
                {{ t('home.hero.subtitleDesc') }}
              </p>

              <div class="mt-7 flex flex-col items-center gap-3 sm:flex-row lg:justify-start">
                <router-link :to="isAuthenticated ? dashboardPath : loginPath" class="btn btn-primary btn-lg px-7">
                  {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                  <Icon name="arrowRight" size="md" :stroke-width="2" />
                </router-link>

                <!-- Key Query: shown on reseller domains -->
                <router-link v-if="isResellerDomain" to="/key-query" class="btn btn-secondary btn-lg px-7">
                  <Icon name="search" size="md" />
                  {{ t('keyQuery.title') }}
                </router-link>


                <!-- Docs: in-app for reseller, external for main site -->
                <router-link
                  v-if="isResellerDomain && docUrl"
                  to="/docs"
                  class="inline-flex items-center gap-2 text-sm font-semibold text-gray-700 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
                >
                  <Icon name="book" size="sm" />
                  {{ t('home.docs') }}
                </router-link>
                <a
                  v-else-if="!isResellerDomain && docUrl"
                  :href="docUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="inline-flex items-center gap-2 text-sm font-semibold text-gray-700 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
                >
                  <Icon name="book" size="sm" />
                  {{ t('home.docs') }}
                </a>
              </div>

              <div class="mt-6 space-y-3">
                <div class="space-y-1">
                  <div class="flex flex-wrap items-center justify-center gap-2 text-xs lg:justify-start">
                    <span class="font-semibold text-gray-700 dark:text-dark-200">{{ t('home.hero.supportedToolsLabel') }}</span>
                    <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Terminal</span>
                    <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">VS Code</span>
                    <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Codex CLI</span>
                  </div>
                </div>

                <div class="flex flex-wrap items-center justify-center gap-2 text-xs lg:justify-start">
                  <span class="font-semibold text-gray-700 dark:text-dark-200">{{ t('home.hero.platformsLabel') }}</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">macOS</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Windows</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Linux</span>
                </div>
              </div>
            </div>

            <!-- Terminal (typewriter) -->
            <div
              id="quickstart"
              class="w-full rounded-4xl border border-gray-200/60 bg-white/70 p-5 shadow-2xl backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40 lg:p-6"
            >
              <div class="flex items-center gap-2">
                <span class="h-2.5 w-2.5 rounded-full bg-red-400/80"></span>
                <span class="h-2.5 w-2.5 rounded-full bg-amber-400/80"></span>
                <span class="h-2.5 w-2.5 rounded-full bg-emerald-400/80"></span>
                <span class="ml-2 text-xs font-semibold text-gray-700 dark:text-dark-200">
                  {{ t('home.hero.terminal.title') }}
                </span>
              </div>

              <div class="mt-4 overflow-hidden rounded-3xl bg-gray-950 p-4 ring-1 ring-black/10 sm:p-5 dark:ring-white/10">
                <TypewriterTerminal :lines="terminalLines" :loop="true" />
              </div>

              <p class="mt-3 text-xs leading-relaxed text-gray-500 dark:text-dark-400">
                {{ t('home.hero.terminal.hint') }}
              </p>

              <p class="mt-2 flex items-center gap-1.5 text-xs font-medium text-primary-600 dark:text-primary-400">
                <Icon name="check" size="sm" />
                {{ t('home.hero.terminal.compatibility') }}
              </p>
            </div>
          </div>

        </div>
      </section>

      <!-- Features -->
      <section id="features" class="mx-auto max-w-7xl px-6 py-12 lg:py-16">
        <div class="mx-auto max-w-3xl text-center">
          <h2 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-3xl">
            {{ t('home.why.title', { siteName }) }}
          </h2>
          <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400 sm:text-base">
            {{ t('home.why.subtitle') }}
          </p>
        </div>

        <div class="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <div class="card card-hover p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 to-blue-600 shadow-lg shadow-blue-500/20"
              >
                <Icon name="server" size="lg" class="text-white" />
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('home.features.unifiedGateway') }}
                </h3>
                <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
                  {{ t('home.features.unifiedGatewayDesc') }}
                </p>
              </div>
            </div>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 shadow-lg shadow-primary-500/20"
              >
                <Icon name="users" size="lg" class="text-white" />
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('home.features.multiAccount') }}
                </h3>
                <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
                  {{ t('home.features.multiAccountDesc') }}
                </p>
              </div>
            </div>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-500 to-violet-600 shadow-lg shadow-violet-500/20"
              >
                <Icon name="calculator" size="lg" class="text-white" />
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('home.features.balanceQuota') }}
                </h3>
                <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
                  {{ t('home.features.balanceQuotaDesc') }}
                </p>
              </div>
            </div>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-amber-500 to-orange-500 shadow-lg shadow-amber-500/20"
              >
                <Icon name="shield" size="lg" class="text-white" />
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('home.features.rateLimit') }}
                </h3>
                <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
                  {{ t('home.features.rateLimitDesc') }}
                </p>
              </div>
            </div>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-emerald-500 to-green-600 shadow-lg shadow-emerald-500/20"
              >
                <Icon name="cpu" size="lg" class="text-white" />
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('home.features.concurrency') }}
                </h3>
                <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
                  {{ t('home.features.concurrencyDesc') }}
                </p>
              </div>
            </div>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-start gap-4">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-gray-700 to-gray-900 shadow-lg shadow-gray-900/20 dark:from-dark-700 dark:to-dark-950"
              >
                <Icon name="cog" size="lg" class="text-white" />
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('home.features.adminConsole') }}
                </h3>
                <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
                  {{ t('home.features.adminConsoleDesc') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Metrics -->
      <section class="mx-auto max-w-7xl px-6 pb-12 pt-2 lg:pb-16">
        <div class="mx-auto max-w-3xl text-center">
          <h2 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-3xl">
            {{ t('home.metrics.title') }}
          </h2>
          <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400 sm:text-base">
            {{ t('home.metrics.subtitle') }}
          </p>
        </div>

        <div class="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <div class="card card-hover p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                <Icon name="bolt" size="md" />
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.metrics.items.setup') }}</div>
            </div>
            <div class="mt-4 h-2.5 rounded-full bg-gray-200 dark:bg-dark-800">
              <div class="h-2.5 w-[85%] rounded-full bg-gradient-to-r from-primary-500 to-sky-500"></div>
            </div>
            <p class="mt-3 text-xs leading-relaxed text-gray-600 dark:text-dark-400">{{ t('home.metrics.items.setupDesc') }}</p>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                <Icon name="dollar" size="md" />
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.metrics.items.cost') }}</div>
            </div>
            <div class="mt-4 h-2.5 rounded-full bg-gray-200 dark:bg-dark-800">
              <div class="h-2.5 w-[80%] rounded-full bg-gradient-to-r from-emerald-500 to-primary-500"></div>
            </div>
            <p class="mt-3 text-xs leading-relaxed text-gray-600 dark:text-dark-400">{{ t('home.metrics.items.costDesc') }}</p>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                <Icon name="chart" size="md" />
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.metrics.items.transparent') }}</div>
            </div>
            <div class="mt-4 h-2.5 rounded-full bg-gray-200 dark:bg-dark-800">
              <div class="h-2.5 w-[90%] rounded-full bg-gradient-to-r from-sky-500 to-primary-500"></div>
            </div>
            <p class="mt-3 text-xs leading-relaxed text-gray-600 dark:text-dark-400">{{ t('home.metrics.items.transparentDesc') }}</p>
          </div>

          <div class="card card-hover p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                <Icon name="grid" size="md" />
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.metrics.items.coverage') }}</div>
            </div>
            <div class="mt-4 h-2.5 rounded-full bg-gray-200 dark:bg-dark-800">
              <div class="h-2.5 w-[88%] rounded-full bg-gradient-to-r from-violet-500 to-sky-500"></div>
            </div>
            <p class="mt-3 text-xs leading-relaxed text-gray-600 dark:text-dark-400">{{ t('home.metrics.items.coverageDesc') }}</p>
          </div>
        </div>
      </section>

      <!-- Tool Cards -->
      <section class="mx-auto max-w-7xl px-6 pb-14 pt-2">
        <div class="mx-auto max-w-3xl text-center">
          <h2 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-3xl">
            {{ t('home.tools.title') }}
          </h2>
          <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400 sm:text-base">
            {{ t('home.tools.subtitle') }}
          </p>
        </div>
        <div class="mx-auto mt-8 grid max-w-3xl gap-5 sm:grid-cols-2">
          <router-link
            to="/docs/claude"
            class="group flex flex-col rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-primary-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-700"
          >
            <div class="mb-4 flex items-center gap-3">
              <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 shadow-md shadow-primary-500/25">
                <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 7.5l3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0021 18V6a2.25 2.25 0 00-2.25-2.25H5.25A2.25 2.25 0 003 6v12a2.25 2.25 0 002.25 2.25z" />
                </svg>
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">Claude Code</h3>
                <p class="text-xs text-gray-400 dark:text-gray-500">claude-code CLI</p>
              </div>
            </div>
            <p class="flex-1 text-sm leading-relaxed text-gray-500 dark:text-gray-400">{{ t('docs.entry.claudeDesc') }}</p>
            <div class="mt-4 flex items-center gap-1 text-sm font-medium text-primary-600 transition-colors group-hover:text-primary-700 dark:text-primary-400">
              {{ t('docs.entry.viewGuide') }}
              <svg class="h-4 w-4 transition-transform duration-150 group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
              </svg>
            </div>
          </router-link>

          <router-link
            to="/docs/codex"
            class="group flex flex-col rounded-2xl border border-gray-200 bg-white p-6 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-emerald-300 hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:hover:border-emerald-700"
          >
            <div class="mb-4 flex items-center gap-3">
              <div class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 shadow-md shadow-emerald-500/25">
                <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M17.25 6.75L22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3l-4.5 16.5" />
                </svg>
              </div>
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">Codex CLI</h3>
                <p class="text-xs text-gray-400 dark:text-gray-500">@openai/codex</p>
              </div>
            </div>
            <p class="flex-1 text-sm leading-relaxed text-gray-500 dark:text-gray-400">{{ t('docs.entry.codexDesc') }}</p>
            <div class="mt-4 flex items-center gap-1 text-sm font-medium text-emerald-600 transition-colors group-hover:text-emerald-700 dark:text-emerald-400">
              {{ t('docs.entry.viewGuide') }}
              <svg class="h-4 w-4 transition-transform duration-150 group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
              </svg>
            </div>
          </router-link>
        </div>
      </section>

      <!-- FAQ -->
      <section id="faq" class="mx-auto max-w-7xl px-6 pb-14 pt-2 lg:pb-20">
        <div class="mx-auto max-w-3xl text-center">
          <h2 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-3xl">
            {{ t('home.faq.title') }}
          </h2>
          <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400 sm:text-base">
            {{ t('home.faq.subtitle') }}
          </p>
        </div>

        <div class="mx-auto mt-8 max-w-3xl space-y-3">
          <details
            open
            class="group rounded-4xl border border-gray-200/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40"
          >
            <summary class="flex cursor-pointer list-none items-center justify-between gap-4">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.faq.q0') }}</span>
              <Icon name="chevronDown" size="sm" class="text-gray-500 transition-transform group-open:rotate-180 dark:text-dark-400" />
            </summary>
            <p class="mt-3 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.faq.a0') }}
            </p>
          </details>

          <details
            class="group rounded-4xl border border-gray-200/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40"
          >
            <summary class="flex cursor-pointer list-none items-center justify-between gap-4">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.faq.q1') }}</span>
              <Icon name="chevronDown" size="sm" class="text-gray-500 transition-transform group-open:rotate-180 dark:text-dark-400" />
            </summary>
            <p class="mt-3 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.faq.a1') }}
            </p>
          </details>

          <details
            class="group rounded-4xl border border-gray-200/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40"
          >
            <summary class="flex cursor-pointer list-none items-center justify-between gap-4">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.faq.q2') }}</span>
              <Icon name="chevronDown" size="sm" class="text-gray-500 transition-transform group-open:rotate-180 dark:text-dark-400" />
            </summary>
            <p class="mt-3 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.faq.a2') }}
            </p>
          </details>

          <details
            class="group rounded-4xl border border-gray-200/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40"
          >
            <summary class="flex cursor-pointer list-none items-center justify-between gap-4">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.faq.q3') }}</span>
              <Icon name="chevronDown" size="sm" class="text-gray-500 transition-transform group-open:rotate-180 dark:text-dark-400" />
            </summary>
            <p class="mt-3 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.faq.a3') }}
            </p>
          </details>

          <details
            class="group rounded-4xl border border-gray-200/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40"
          >
            <summary class="flex cursor-pointer list-none items-center justify-between gap-4">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.faq.q4') }}</span>
              <Icon name="chevronDown" size="sm" class="text-gray-500 transition-transform group-open:rotate-180 dark:text-dark-400" />
            </summary>
            <p class="mt-3 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.faq.a4') }}
            </p>
          </details>

          <details
            class="group rounded-4xl border border-gray-200/60 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40"
          >
            <summary class="flex cursor-pointer list-none items-center justify-between gap-4">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.faq.q5') }}</span>
              <Icon name="chevronDown" size="sm" class="text-gray-500 transition-transform group-open:rotate-180 dark:text-dark-400" />
            </summary>
            <p class="mt-3 text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.faq.a5') }}
            </p>
          </details>
        </div>

        <div v-if="docUrl" class="mt-6 text-center">
          <router-link
            v-if="isResellerDomain"
            to="/docs"
            class="inline-flex items-center gap-2 text-sm font-semibold text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
          >
            {{ t('home.faq.more') }}
            <Icon name="arrowRight" size="sm" />
          </router-link>
          <a
            v-else
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-2 text-sm font-semibold text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
          >
            {{ t('home.faq.more') }}
            <Icon name="arrowRight" size="sm" />
          </a>
        </div>
      </section>

      <!-- CTA -->
      <section class="mx-auto max-w-7xl px-6 pb-16 lg:pb-24">
        <div
          class="relative overflow-hidden rounded-4xl bg-gradient-to-br from-gray-900 via-dark-900 to-dark-950 px-8 py-10 text-white shadow-2xl"
        >
          <div class="pointer-events-none absolute inset-0 opacity-50">
            <div class="absolute -left-20 -top-20 h-72 w-72 rounded-full bg-primary-500/25 blur-3xl"></div>
            <div class="absolute -bottom-28 -right-24 h-80 w-80 rounded-full bg-sky-500/20 blur-3xl"></div>
          </div>

          <div class="relative flex flex-col items-center justify-between gap-8 text-center lg:flex-row lg:text-left">
            <div class="max-w-2xl">
              <h3 class="text-2xl font-semibold tracking-tight sm:text-3xl">{{ t('home.cta.title') }}</h3>
              <p class="mt-3 text-sm leading-relaxed text-gray-300 sm:text-base">
                {{ t('home.cta.description') }}
              </p>
            </div>

            <div class="flex w-full flex-col gap-3 sm:w-auto sm:flex-row">
              <router-link
                :to="isAuthenticated ? dashboardPath : loginPath"
                class="btn btn-primary btn-lg w-full px-7 sm:w-auto"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" :stroke-width="2" />
              </router-link>
              <router-link
                v-if="isResellerDomain && docUrl"
                to="/docs"
                class="btn btn-secondary btn-lg w-full border-white/10 bg-white/10 text-white hover:bg-white/15 sm:w-auto"
              >
                <Icon name="book" size="md" />
                {{ t('home.docs') }}
              </router-link>
              <a
                v-else-if="!isResellerDomain && docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="btn btn-secondary btn-lg w-full border-white/10 bg-white/10 text-white hover:bg-white/15 sm:w-auto"
              >
                <Icon name="book" size="md" />
                {{ t('home.docs') }}
              </a>
            </div>
          </div>
        </div>
      </section>
    </main>

    <!-- Footer -->
    <footer class="relative z-10 border-t border-gray-200/60 px-6 py-8 dark:border-dark-800/60">
      <div
        class="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 text-center sm:flex-row sm:text-left"
      >
        <p class="text-sm text-gray-500 dark:text-dark-400">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-4">
          <router-link
            v-if="isResellerDomain && docUrl"
            to="/docs"
            class="text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-white"
          >
            {{ t('home.docs') }}
          </router-link>
          <a
            v-else-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-gray-500 transition-colors hover:text-gray-700 dark:text-dark-400 dark:hover:text-white"
          >
            {{ t('home.docs') }}
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAuthStore, useAppStore } from '@/stores'
import TypewriterTerminal, { type TerminalLine } from '@/components/common/TypewriterTerminal.vue'
import PublicHeader from '@/components/layout/PublicHeader.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const route = useRoute()

const authStore = useAuthStore()
const appStore = useAppStore()

// Get referral code from URL if present
const refCode = computed(() => route.query.ref as string | undefined)
const registerPath = computed(() => refCode.value ? `/register?ref=${refCode.value}` : '/register')
const loginPath = computed(() => refCode.value ? `/login?redirect=${encodeURIComponent(registerPath.value)}` : '/login')

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '码驿站')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const homeTemplate = computed(() => appStore.cachedPublicSettings?.home_template || 'default')
const apiBaseRoot = computed(() => {
  const fallback = typeof window !== 'undefined' ? window.location.origin : 'https://YOUR_HOST'
  const raw = (appStore.cachedPublicSettings?.api_base_url || fallback).trim()
  const trimmed = raw.replace(/\/+$/, '')
  return trimmed.replace(/\/v1beta$/, '').replace(/\/v1$/, '')
})

// Reseller domain detection
const isResellerDomain = computed(() => !!appStore.cachedPublicSettings?.reseller_id)

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

function buildTerminalLines(baseUrl: string): TerminalLine[] {
  const token = 'sk-xxx'
  return [
    { kind: 'comment', text: t('home.hero.terminal.claude.line1') },
    { kind: 'prompt', text: `export ANTHROPIC_BASE_URL="${baseUrl}"` },
    { kind: 'prompt', text: `export ANTHROPIC_AUTH_TOKEN="${token}"` },
    { kind: 'comment', text: t('home.hero.terminal.claude.line2') },
    { kind: 'prompt', text: 'claude' },
    { kind: 'success', text: t('home.hero.terminal.claude.success') }
  ]
}

const terminalLines = computed<TerminalLine[]>(() => buildTerminalLines(apiBaseRoot.value))

// Initialize theme on mount
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
