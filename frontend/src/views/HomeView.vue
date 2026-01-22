<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
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
    <header
      class="sticky top-0 z-30 border-b border-gray-200/60 bg-white/70 backdrop-blur-xl dark:border-dark-800/60 dark:bg-dark-950/50"
    >
      <nav class="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
        <!-- Brand -->
        <router-link to="/" class="flex items-center gap-3">
          <div
            class="h-10 w-10 overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-gray-200/60 dark:bg-dark-900 dark:ring-dark-800"
          >
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <div class="hidden flex-col sm:flex">
            <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ siteName }}</span>
            <span class="text-xs text-gray-500 dark:text-dark-400">{{ siteSubtitle }}</span>
          </div>
        </router-link>

        <!-- Nav -->
        <div class="hidden items-center gap-7 text-sm font-medium lg:flex">
          <a
            href="#quickstart"
            class="text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >{{ t('home.nav.quickstart') }}</a
          >
          <a
            href="#features"
            class="text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >{{ t('home.nav.features') }}</a
          >
          <a
            href="#pricing"
            class="text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >{{ t('home.nav.pricing') }}</a
          >
          <a
            href="#providers"
            class="text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >{{ t('home.nav.providers') }}</a
          >
          <a
            href="#faq"
            class="text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >{{ t('home.nav.faq') }}</a
          >
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-gray-600 transition-colors hover:text-gray-900 dark:text-dark-300 dark:hover:text-white"
            >{{ t('home.docs') }}</a
          >
        </div>

        <!-- Actions -->
        <div class="flex items-center gap-2.5">
          <LocaleSwitcher />

          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="hidden rounded-xl p-2 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-900 dark:hover:text-white sm:inline-flex"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <button
            @click="toggleTheme"
            class="rounded-xl p-2 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-900 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-1.5 rounded-full bg-gray-900 py-1 pl-1 pr-2.5 transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            <span
              class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[10px] font-semibold text-white"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-medium text-white">{{ t('home.dashboard') }}</span>
            <Icon name="arrowRight" size="xs" class="text-gray-300" :stroke-width="2" />
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-full bg-gray-900 px-3 py-1 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

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

              <div class="mt-7 flex flex-col items-center gap-3 sm:flex-row lg:justify-start">
                <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="btn btn-primary btn-lg px-7">
                  {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                  <Icon name="arrowRight" size="md" :stroke-width="2" />
                </router-link>

                <a href="#pricing" class="btn btn-secondary btn-lg px-7">
                  <Icon name="dollar" size="md" />
                  {{ t('home.hero.viewPricing') }}
                </a>

                <a
                  v-if="docUrl"
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
                <div class="flex flex-wrap items-center justify-center gap-2 text-xs lg:justify-start">
                  <span class="font-semibold text-gray-700 dark:text-dark-200">{{ t('home.hero.supportedToolsLabel') }}</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Claude Code</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Codex CLI</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Gemini CLI</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">Cursor</span>
                  <span class="rounded-full border border-gray-200/60 bg-white/70 px-2.5 py-1 text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">OpenCode</span>
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
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div class="flex items-center gap-2">
                  <span class="h-2.5 w-2.5 rounded-full bg-red-400/80"></span>
                  <span class="h-2.5 w-2.5 rounded-full bg-amber-400/80"></span>
                  <span class="h-2.5 w-2.5 rounded-full bg-emerald-400/80"></span>
                  <span class="ml-2 text-xs font-semibold text-gray-700 dark:text-dark-200">
                    {{ t('home.hero.terminal.title') }}
                  </span>
                </div>

                <div class="flex flex-wrap items-center gap-2">
                  <div
                    class="flex items-center gap-1 rounded-full border border-gray-200/60 bg-white/60 p-1 text-[11px] font-semibold text-gray-600 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/50 dark:text-dark-300"
                    role="tablist"
                    aria-label="Quickstart examples"
                  >
                    <button
                      v-for="tab in demoTabs"
                      :key="tab.key"
                      type="button"
                      role="tab"
                      :aria-selected="activeDemo === tab.key"
                      class="rounded-full px-2.5 py-1 transition-colors"
                      :class="activeDemo === tab.key ? 'bg-gray-900 text-white dark:bg-white dark:text-gray-900' : 'hover:bg-gray-100 dark:hover:bg-dark-800/60'"
                      @click="activeDemo = tab.key"
                    >
                      {{ tab.label }}
                    </button>
                  </div>

                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-gray-200/60 bg-white/60 px-3 py-1.5 text-xs font-semibold text-gray-700 shadow-sm transition-colors hover:bg-white dark:border-dark-700/70 dark:bg-dark-900/50 dark:text-dark-200 dark:hover:bg-dark-900"
                    @click="copyTerminalExample"
                  >
                    <Icon name="copy" size="sm" />
                    {{ t('home.hero.terminal.copy') }}
                  </button>
                </div>
              </div>

              <div class="mt-4 overflow-x-auto rounded-3xl bg-gray-950 p-4 ring-1 ring-black/10 sm:p-5 dark:ring-white/10">
                <TypewriterTerminal :lines="terminalLines" :loop="true" />
              </div>

              <p class="mt-3 text-xs leading-relaxed text-gray-500 dark:text-dark-400">
                {{ t('home.hero.terminal.hint') }}
              </p>
            </div>
          </div>

          <!-- Quick stats - compact horizontal badges -->
          <div class="mx-auto mt-10 flex max-w-4xl flex-wrap items-center justify-center gap-3">
            <div class="inline-flex items-center gap-1.5 rounded-full border border-gray-200/60 bg-white/70 px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">
              <Icon name="bolt" size="xs" class="text-primary-600 dark:text-primary-300" />
              <span>{{ t('home.hero.stats.standard') }}</span>
            </div>
            <div class="inline-flex items-center gap-1.5 rounded-full border border-gray-200/60 bg-white/70 px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">
              <Icon name="copy" size="xs" class="text-primary-600 dark:text-primary-300" />
              <span>{{ t('home.hero.stats.copy') }}</span>
            </div>
            <div class="inline-flex items-center gap-1.5 rounded-full border border-gray-200/60 bg-white/70 px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">
              <Icon name="dollar" size="xs" class="text-primary-600 dark:text-primary-300" />
              <span>{{ t('home.hero.stats.transparent') }}</span>
            </div>
            <div class="inline-flex items-center gap-1.5 rounded-full border border-gray-200/60 bg-white/70 px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">
              <Icon name="shield" size="xs" class="text-primary-600 dark:text-primary-300" />
              <span>{{ t('home.hero.stats.protect') }}</span>
            </div>
            <div class="inline-flex items-center gap-1.5 rounded-full border border-gray-200/60 bg-white/70 px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm dark:border-dark-700/70 dark:bg-dark-900/40 dark:text-dark-200">
              <Icon name="server" size="xs" class="text-primary-600 dark:text-primary-300" />
              <span>{{ t('home.hero.stats.stable') }}</span>
            </div>
          </div>

          <!-- 3 steps - more compact -->
          <div class="mx-auto mt-6 max-w-4xl">
            <div class="flex flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-center">
              <div class="flex items-center gap-3 rounded-2xl border border-gray-200/60 bg-white/70 px-4 py-3 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40">
                <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                  <Icon name="login" size="sm" />
                </div>
                <div>
                  <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.steps.oneTitle') }}</div>
                  <div class="text-xs text-gray-600 dark:text-dark-400">{{ t('home.steps.oneDesc') }}</div>
                </div>
              </div>

              <Icon name="arrowRight" size="sm" class="hidden text-gray-400 dark:text-dark-500 sm:block" />

              <div class="flex items-center gap-3 rounded-2xl border border-gray-200/60 bg-white/70 px-4 py-3 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40">
                <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                  <Icon name="grid" size="sm" />
                </div>
                <div>
                  <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.steps.twoTitle') }}</div>
                  <div class="text-xs text-gray-600 dark:text-dark-400">{{ t('home.steps.twoDesc') }}</div>
                </div>
              </div>

              <Icon name="arrowRight" size="sm" class="hidden text-gray-400 dark:text-dark-500 sm:block" />

              <div class="flex items-center gap-3 rounded-2xl border border-gray-200/60 bg-white/70 px-4 py-3 shadow-sm backdrop-blur dark:border-dark-800/60 dark:bg-dark-900/40">
                <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                  <Icon name="play" size="sm" />
                </div>
                <div>
                  <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.steps.threeTitle') }}</div>
                  <div class="text-xs text-gray-600 dark:text-dark-400">{{ t('home.steps.threeDesc') }}</div>
                </div>
              </div>
            </div>
          </div>

          <!-- Free Trial Banner -->
          <div class="mx-auto mt-10 max-w-2xl">
            <div class="relative overflow-hidden rounded-3xl bg-gradient-to-r from-green-500 via-emerald-500 to-teal-500 p-1 shadow-xl shadow-green-500/20">
              <div class="rounded-[22px] bg-white/95 px-6 py-5 backdrop-blur dark:bg-dark-900/95">
                <div class="flex flex-col items-center gap-5 sm:flex-row">
                  <!-- QR Code -->
                  <div class="flex-shrink-0">
                    <div class="overflow-hidden rounded-2xl border-2 border-green-100 bg-white p-1.5 shadow-sm dark:border-green-900/50">
                      <img
                        :src="qrCodeUrl"
                        alt="WeChat QR Code"
                        class="h-28 w-28 object-contain"
                        @error="handleQrError"
                      />
                    </div>
                  </div>
                  <!-- Text -->
                  <div class="flex-1 text-center sm:text-left">
                    <div class="inline-flex items-center gap-1.5 rounded-full bg-green-100 px-3 py-1 text-xs font-semibold text-green-700 dark:bg-green-900/30 dark:text-green-300">
                      <span class="relative flex h-2 w-2">
                        <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"></span>
                        <span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
                      </span>
                      {{ t('home.freeTrial.badge') }}
                    </div>
                    <h3 class="mt-2 text-lg font-bold text-gray-900 dark:text-white">
                      {{ t('home.freeTrial.title') }}
                    </h3>
                    <p class="mt-1 text-sm text-gray-600 dark:text-dark-400">
                      {{ t('home.freeTrial.description') }}
                    </p>
                    <div class="mt-2 inline-flex items-center gap-2 text-sm font-medium text-green-600 dark:text-green-400">
                      <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 01.213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 00.167-.054l1.903-1.114a.864.864 0 01.717-.098 10.16 10.16 0 002.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178A1.17 1.17 0 014.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178 1.17 1.17 0 01-1.162-1.178c0-.651.52-1.18 1.162-1.18z"/>
                      </svg>
                      <span>AI码驿站</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Features -->
      <section id="features" class="mx-auto max-w-7xl px-6 py-12 lg:py-16">
        <div class="mx-auto max-w-3xl text-center">
          <h2 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-3xl">
            {{ t('home.why.title') }}
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

      <!-- Pricing -->
      <section id="pricing" class="relative overflow-hidden py-14 lg:py-20">
        <div class="absolute inset-0 bg-gray-950"></div>
        <div class="pointer-events-none absolute inset-0">
          <div class="absolute -top-36 left-1/2 h-[26rem] w-[44rem] -translate-x-1/2 rounded-full bg-primary-500/15 blur-3xl"></div>
          <div class="absolute -bottom-44 right-[-14rem] h-[28rem] w-[28rem] rounded-full bg-primary-500/10 blur-3xl"></div>
          <div class="absolute inset-0 opacity-60 [mask-image:radial-gradient(closest-side,white,transparent)]">
            <div
              class="h-full w-full bg-[linear-gradient(rgba(255,255,255,0.08)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.08)_1px,transparent_1px)] bg-[size:84px_84px]"
            ></div>
          </div>
        </div>

        <div class="relative mx-auto max-w-7xl px-6">
          <div class="mx-auto max-w-3xl text-center text-white">
            <h2 class="text-balance text-3xl font-semibold tracking-tight sm:text-4xl">
              {{ t('home.pricing.title') }}
            </h2>
            <p class="mt-3 text-sm leading-relaxed text-gray-300 sm:text-base">
              {{ t('home.pricing.subtitle') }}
            </p>
          </div>

          <!-- Plans Grid -->
          <div class="mt-10 grid gap-6 lg:grid-cols-4">
            <div
              v-for="plan in pricingPlans"
              :key="plan.key"
              class="relative flex flex-col overflow-hidden rounded-4xl border border-white/10 bg-white/5 p-6 shadow-[0_18px_60px_rgba(0,0,0,0.35)]"
              :class="plan.recommended ? 'ring-1 ring-primary-500/40' : ''"
            >
              <!-- Recommended Badge -->
              <div v-if="plan.recommended" class="absolute right-4 top-4">
                <span class="inline-flex items-center gap-1 rounded-full bg-primary-500 px-3 py-1 text-xs font-semibold text-white">
                  <Icon name="sparkles" size="xs" />
                  {{ t('home.pricing.recommended') }}
                </span>
              </div>

              <!-- Plan Name -->
              <div class="text-sm font-semibold text-gray-200">{{ plan.name }}</div>

              <!-- Price -->
              <div class="mt-4 flex items-end gap-2">
                <span class="text-xl font-semibold text-gray-300">¥</span>
                <span class="text-5xl font-semibold tracking-tight text-white">{{ plan.price }}</span>
              </div>
              <div class="mt-2 text-xs text-gray-400">
                {{ plan.credit }}
              </div>

              <!-- Promotional Badge -->
              <div class="mt-3 inline-flex items-center gap-1 self-start rounded-md bg-gradient-to-r from-amber-400/20 to-orange-400/20 px-2.5 py-1 text-xs font-medium text-amber-200">
                <Icon name="sparkles" size="xs" />
                <span>{{ plan.unitPrice }}</span>
              </div>

              <!-- Features List -->
              <ul class="mt-6 space-y-3 text-sm text-gray-200">
                <li v-for="(feature, idx) in plan.features" :key="idx" class="flex items-start gap-2">
                  <Icon
                    name="check"
                    size="sm"
                    class="mt-0.5 flex-shrink-0"
                    :class="plan.recommended ? 'text-primary-400' : 'text-emerald-400'"
                  />
                  <span class="leading-relaxed">{{ feature }}</span>
                </li>
              </ul>

              <!-- Buy Button -->
              <div class="mt-auto pt-6">
                <router-link
                  :to="isAuthenticated ? '/plans' : '/login?redirect=/plans'"
                  class="inline-flex w-full items-center justify-center rounded-2xl px-4 py-3 text-sm font-semibold transition-colors"
                  :class="plan.recommended ? 'bg-primary-500 text-white hover:bg-primary-600' : 'border border-white/10 bg-white/5 text-white hover:bg-white/10'"
                >
                  {{ t('home.pricing.buy') }}
                </router-link>
              </div>
            </div>
          </div>

          <p class="mx-auto mt-8 max-w-4xl text-center text-xs leading-relaxed text-gray-500">
            {{ t('home.pricing.note') }}
          </p>
        </div>
      </section>

      <!-- Providers -->
      <section id="providers" class="mx-auto max-w-7xl px-6 pb-12 pt-2 lg:pb-16">
        <div class="mx-auto max-w-3xl text-center">
          <h2 class="text-2xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-3xl">
            {{ t('home.providers.title') }}
          </h2>
          <p class="mt-2 text-sm leading-relaxed text-gray-600 dark:text-dark-400 sm:text-base">
            {{ t('home.providers.description') }}
          </p>
        </div>

        <div class="mt-8 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <div class="card card-hover flex items-center justify-between gap-4 p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-orange-400 to-orange-600">
                <span class="text-xs font-bold text-white">C</span>
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.providers.claude') }}</div>
            </div>
            <span class="rounded-full bg-primary-100 px-2 py-1 text-[10px] font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <div class="card card-hover flex items-center justify-between gap-4 p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-green-500 to-emerald-600">
                <span class="text-xs font-bold text-white">G</span>
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">GPT</div>
            </div>
            <span class="rounded-full bg-primary-100 px-2 py-1 text-[10px] font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <div class="card card-hover flex items-center justify-between gap-4 p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-blue-500 to-sky-600">
                <span class="text-xs font-bold text-white">G</span>
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.providers.gemini') }}</div>
            </div>
            <span class="rounded-full bg-primary-100 px-2 py-1 text-[10px] font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <div class="card card-hover flex items-center justify-between gap-4 p-6">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-rose-500 to-pink-600">
                <span class="text-xs font-bold text-white">A</span>
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.providers.antigravity') }}</div>
            </div>
            <span class="rounded-full bg-primary-100 px-2 py-1 text-[10px] font-semibold text-primary-700 dark:bg-primary-900/40 dark:text-primary-300">
              {{ t('home.providers.supported') }}
            </span>
          </div>

          <div class="card card-hover flex items-center justify-between gap-4 p-6 opacity-80">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 items-center justify-center rounded-2xl bg-gradient-to-br from-gray-500 to-gray-700 dark:from-dark-600 dark:to-dark-800">
                <span class="text-xs font-bold text-white">+</span>
              </div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('home.providers.more') }}</div>
            </div>
            <span class="rounded-full bg-gray-100 px-2 py-1 text-[10px] font-semibold text-gray-600 dark:bg-dark-800 dark:text-dark-300">
              {{ t('home.providers.soon') }}
            </span>
          </div>
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
          <a
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
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary btn-lg w-full px-7 sm:w-auto"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" :stroke-width="2" />
              </router-link>
              <a
                v-if="docUrl"
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
          <a
            v-if="docUrl"
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
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import TypewriterTerminal, { type TerminalLine } from '@/components/common/TypewriterTerminal.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

type DemoKey = 'claude' | 'codex' | 'gemini'

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const apiBaseRoot = computed(() => {
  const fallback = typeof window !== 'undefined' ? window.location.origin : 'https://YOUR_HOST'
  const raw = (appStore.cachedPublicSettings?.api_base_url || fallback).trim()
  const trimmed = raw.replace(/\/+$/, '')
  return trimmed.replace(/\/v1beta$/, '').replace(/\/v1$/, '')
})

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// QR code URL (runtime to avoid Vite static analysis)
const qrCodeUrl = '/wechat-qrcode.png'

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

const demoTabs = computed(() => [
  { key: 'claude' as const, label: 'Claude Code' },
  { key: 'codex' as const, label: 'Codex CLI' },
  { key: 'gemini' as const, label: 'Gemini CLI' }
])

const activeDemo = ref<DemoKey>('claude')

function buildTerminalLines(baseUrl: string, demo: DemoKey): TerminalLine[] {
  const token = 'sk-xxx'

  switch (demo) {
    case 'codex':
      return [
        { kind: 'comment', text: t('home.hero.terminal.codex.line1') },
        { kind: 'output', text: '[model_providers.sub2api]' },
        { kind: 'output', text: `base_url = "${baseUrl}"` },
        { kind: 'output', text: 'requires_openai_auth = true' },
        { kind: 'comment', text: t('home.hero.terminal.codex.line2') },
        { kind: 'output', text: '{' },
        { kind: 'output', text: `  "OPENAI_API_KEY": "${token}"` },
        { kind: 'output', text: '}' },
        { kind: 'comment', text: t('home.hero.terminal.codex.line3') },
        { kind: 'prompt', text: 'codex' }
      ]
    case 'gemini':
      return [
        { kind: 'comment', text: t('home.hero.terminal.gemini.line1') },
        { kind: 'prompt', text: `export GOOGLE_GEMINI_BASE_URL="${baseUrl}"` },
        { kind: 'prompt', text: `export GEMINI_API_KEY="${token}"` },
        { kind: 'prompt', text: 'export GEMINI_MODEL="gemini-2.0-flash"' },
        { kind: 'comment', text: t('home.hero.terminal.gemini.line2') },
        { kind: 'prompt', text: 'gemini' },
        { kind: 'success', text: t('home.hero.terminal.gemini.success') }
      ]
    case 'claude':
    default:
      return [
        { kind: 'comment', text: t('home.hero.terminal.claude.line1') },
        { kind: 'prompt', text: `export ANTHROPIC_BASE_URL="${baseUrl}"` },
        { kind: 'prompt', text: `export ANTHROPIC_AUTH_TOKEN="${token}"` },
        { kind: 'comment', text: t('home.hero.terminal.claude.line2') },
        { kind: 'prompt', text: 'claude' },
        { kind: 'success', text: t('home.hero.terminal.claude.success') }
      ]
  }
}

const terminalLines = computed<TerminalLine[]>(() => buildTerminalLines(apiBaseRoot.value, activeDemo.value))

const terminalExampleText = computed(() => terminalLines.value.map((l) => l.text).join('\n'))

// Static pricing plans with promotional info
const pricingPlans = computed(() => [
  {
    key: 'starter',
    recommended: false,
    name: '体验版 🎁',
    price: '9.9',
    credit: '$22 月限额',
    unitPrice: '¥0.45/刀',
    badge: '新人专享',
    features: [
      '月限额 $22（约 3300 万 tokens）',
      '30 天有效期',
      '支持全模型',
      '单价仅 ¥0.45/刀',
      '适合轻度使用、快速体验'
    ]
  },
  {
    key: 'lite',
    recommended: false,
    name: '基础版',
    price: '29.9',
    credit: '$70 月限额',
    unitPrice: '¥0.43/刀',
    badge: '日常使用',
    features: [
      '月限额 $70（约 1 亿 tokens）',
      '30 天有效期',
      '支持全模型',
      '单价仅 ¥0.43/刀',
      '满足日常开发需求'
    ]
  },
  {
    key: 'standard',
    recommended: true,
    name: '标准版 ⭐',
    price: '69.9',
    credit: '$170 月限额',
    unitPrice: '¥0.41/刀',
    badge: '性价比之选',
    features: [
      '月限额 $170（约 2.5 亿 tokens）',
      '30 天有效期',
      '支持全模型',
      '单价仅 ¥0.41/刀 💰',
      '最受欢迎，性价比最高'
    ]
  },
  {
    key: 'pro',
    recommended: true,
    name: '专业版 ⭐⭐',
    price: '149.9',
    credit: '$380 月限额',
    unitPrice: '¥0.39/刀',
    badge: '重度使用',
    features: [
      '月限额 $380（约 5.7 亿 tokens）',
      '30 天有效期',
      '支持全模型',
      '单价低至 ¥0.39/刀 🔥',
      '重度使用首选，单价最低'
    ]
  }
])

function copyTerminalExample() {
  copyToClipboard(terminalExampleText.value)
}

// Handle QR code image error
function handleQrError(e: Event) {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
}

// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
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
