<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark-900">
    <!-- Header -->
    <header class="border-b border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
      <div class="mx-auto flex max-w-3xl items-center justify-between px-4 py-4">
        <div class="flex items-center gap-3">
          <img
            v-if="appStore.siteLogo"
            :src="appStore.siteLogo"
            alt="Logo"
            class="h-8 w-8 rounded-lg object-contain"
          />
          <h1 class="text-xl font-bold text-gray-900 dark:text-white">
            {{ appStore.siteName || 'Sub2API' }}
          </h1>
        </div>
        <!-- Language Switcher -->
        <div class="relative" ref="langDropdownRef">
          <button
            @click="showLangDropdown = !showLangDropdown"
            class="flex items-center gap-1.5 rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-600 transition hover:bg-gray-50 dark:border-dark-600 dark:text-gray-300 dark:hover:bg-dark-700"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 003 12c0-1.605.42-3.113 1.157-4.418" />
            </svg>
            <span>{{ currentLocaleName }}</span>
            <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
            </svg>
          </button>
          <div
            v-if="showLangDropdown"
            class="absolute right-0 top-full z-50 mt-1 w-40 rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
          >
            <button
              v-for="loc in availableLocales"
              :key="loc.code"
              @click="switchLocale(loc.code)"
              class="flex w-full items-center px-3 py-2 text-left text-sm transition hover:bg-gray-50 dark:hover:bg-dark-700"
              :class="locale === loc.code ? 'text-primary-600 font-medium dark:text-primary-400' : 'text-gray-700 dark:text-gray-300'"
            >
              {{ loc.name }}
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="mx-auto max-w-3xl px-4 py-8">
      <!-- Title -->
      <div class="mb-6 text-center">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('keyQuery.title') }}
        </h2>
      </div>

      <!-- Search Form -->
      <div class="mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex gap-3">
          <input
            v-model="apiKey"
            type="text"
            class="input flex-1 font-mono text-sm"
            :placeholder="t('keyQuery.placeholder')"
            @keyup.enter="handleQuery"
          />
          <button
            @click="handleQuery"
            :disabled="querying || !apiKey.trim()"
            class="btn btn-primary flex-shrink-0"
          >
            <svg
              v-if="querying"
              class="mr-1.5 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{ querying ? t('keyQuery.querying') : t('keyQuery.submit') }}
          </button>
        </div>

        <!-- Error Message -->
        <p v-if="error" class="mt-3 text-sm text-red-600 dark:text-red-400">
          {{ error }}
        </p>
      </div>

      <!-- Results -->
      <div v-if="result" class="space-y-4">
        <!-- Key Info Card -->
        <div class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('keyQuery.keyInfo') }}
          </h3>
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.keyName') }}</span>
              <span class="font-medium text-gray-900 dark:text-white">{{ result.key.name }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.status') }}</span>
              <span
                class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium"
                :class="statusClass(result.key.status)"
              >
                {{ statusText(result.key.status) }}
              </span>
            </div>
            <div>
              <div class="mb-1 flex items-center justify-between">
                <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.quota') }}</span>
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  <template v-if="!result.key.quota">
                    {{ t('keyQuery.unlimited') }}
                  </template>
                  <template v-else>
                    ${{ (result.key.quota_used ?? 0).toFixed(2) }} / ${{ result.key.quota.toFixed(2) }}
                  </template>
                </span>
              </div>
              <div v-if="result.key.quota" class="h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="h-full rounded-full transition-all"
                  :class="quotaBarClass(result.key.quota_used ?? 0, result.key.quota)"
                  :style="{ width: Math.min(((result.key.quota_used ?? 0) / result.key.quota) * 100, 100) + '%' }"
                ></div>
              </div>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.expiresAt') }}</span>
              <span class="text-sm text-gray-700 dark:text-gray-300">
                <template v-if="!result.key.expires_at">
                  {{ t('keyQuery.neverExpires') }}
                </template>
                <template v-else>
                  {{ t('keyQuery.daysRemaining', { days: result.key.days_until_expiry }) }}
                </template>
              </span>
            </div>
          </div>
        </div>

        <!-- Subscription Card -->
        <div class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('keyQuery.subscription') }}
          </h3>
          <div v-if="!result.subscription" class="py-4 text-center">
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.noSubscription') }}</p>
          </div>
          <div v-else class="space-y-4">
            <div class="flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ result.subscription.group_name }}</span>
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('keyQuery.daysRemaining', { days: result.subscription.days_remaining }) }}
              </span>
            </div>
            <!-- Daily Usage -->
            <div v-if="result.subscription.daily_limit_usd !== null">
              <div class="mb-1 flex items-center justify-between">
                <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.dailyUsage') }}</span>
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  ${{ (result.subscription.daily_usage_usd ?? 0).toFixed(2) }} / ${{ result.subscription.daily_limit_usd.toFixed(2) }}
                </span>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="h-full rounded-full transition-all"
                  :class="quotaBarClass(result.subscription.daily_usage_usd ?? 0, result.subscription.daily_limit_usd)"
                  :style="{ width: usagePercent(result.subscription.daily_usage_usd ?? 0, result.subscription.daily_limit_usd) + '%' }"
                ></div>
              </div>
            </div>
            <div v-else class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.dailyUsage') }}</span>
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('keyQuery.noLimit') }}</span>
            </div>
            <!-- Weekly Usage -->
            <div v-if="result.subscription.weekly_limit_usd !== null">
              <div class="mb-1 flex items-center justify-between">
                <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.weeklyUsage') }}</span>
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  ${{ (result.subscription.weekly_usage_usd ?? 0).toFixed(2) }} / ${{ result.subscription.weekly_limit_usd.toFixed(2) }}
                </span>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="h-full rounded-full transition-all"
                  :class="quotaBarClass(result.subscription.weekly_usage_usd ?? 0, result.subscription.weekly_limit_usd)"
                  :style="{ width: usagePercent(result.subscription.weekly_usage_usd ?? 0, result.subscription.weekly_limit_usd) + '%' }"
                ></div>
              </div>
            </div>
            <div v-else class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.weeklyUsage') }}</span>
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('keyQuery.noLimit') }}</span>
            </div>
            <!-- Monthly Usage -->
            <div v-if="result.subscription.monthly_limit_usd !== null">
              <div class="mb-1 flex items-center justify-between">
                <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.monthlyUsage') }}</span>
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  ${{ (result.subscription.monthly_usage_usd ?? 0).toFixed(2) }} / ${{ result.subscription.monthly_limit_usd.toFixed(2) }}
                </span>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="h-full rounded-full transition-all"
                  :class="quotaBarClass(result.subscription.monthly_usage_usd ?? 0, result.subscription.monthly_limit_usd)"
                  :style="{ width: usagePercent(result.subscription.monthly_usage_usd ?? 0, result.subscription.monthly_limit_usd) + '%' }"
                ></div>
              </div>
            </div>
            <div v-else class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('keyQuery.monthlyUsage') }}</span>
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('keyQuery.noLimit') }}</span>
            </div>
          </div>
        </div>

        <!-- Time Range Selector -->
        <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
          <div class="flex flex-wrap items-center gap-2">
            <button
              v-for="range in timeRanges"
              :key="range"
              @click="setTimeRange(range)"
              class="rounded-lg px-3 py-1.5 text-sm font-medium transition"
              :class="selectedTimeRange === range
                ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-700'"
            >
              {{ t(`keyQuery.timeRange.${range}`) }}
            </button>
            <div v-if="selectedTimeRange === 'custom'" class="flex flex-wrap items-center gap-2 ml-2">
              <input
                type="date"
                v-model="customStartDate"
                @change="fetchTimeRangeData"
                class="rounded-lg border border-gray-300 bg-white px-2 py-1 text-sm text-gray-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300"
              />
              <span class="text-gray-400">-</span>
              <input
                type="date"
                v-model="customEndDate"
                @change="fetchTimeRangeData"
                class="rounded-lg border border-gray-300 bg-white px-2 py-1 text-sm text-gray-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300"
              />
            </div>
          </div>
        </div>

        <!-- Tabs -->
        <div class="flex border-b border-gray-200 dark:border-dark-700">
          <button
            @click="activeTab = 'overview'"
            class="px-4 py-2.5 text-sm font-medium transition"
            :class="activeTab === 'overview'
              ? 'border-b-2 border-primary-500 text-primary-600 dark:text-primary-400'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
          >
            {{ t('keyQuery.tabs.overview') }}
          </button>
          <button
            @click="activeTab = 'usage'"
            class="px-4 py-2.5 text-sm font-medium transition"
            :class="activeTab === 'usage'
              ? 'border-b-2 border-primary-500 text-primary-600 dark:text-primary-400'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
          >
            {{ t('keyQuery.tabs.usageDetails') }}
          </button>
        </div>

        <!-- Overview Tab -->
        <div v-if="activeTab === 'overview'" class="space-y-4">
          <!-- Stats Loading -->
          <div v-if="statsLoading" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <template v-else>
            <!-- Stats Cards -->
            <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
              <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('keyQuery.requestCount') }}</p>
                <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">{{ (stats?.total_requests ?? 0).toLocaleString() }}</p>
              </div>
              <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('keyQuery.totalCost') }}</p>
                <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">${{ (stats?.total_cost ?? 0).toFixed(4) }}</p>
              </div>
              <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('keyQuery.stats.totalTokens') }}</p>
                <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">{{ (stats?.total_tokens ?? 0).toLocaleString() }}</p>
              </div>
              <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('keyQuery.stats.avgDuration') }}</p>
                <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">{{ (stats?.average_duration_ms ?? 0).toFixed(0) }}ms</p>
              </div>
            </div>

            <!-- Token Breakdown -->
            <div class="rounded-xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <div class="grid grid-cols-3 gap-4 text-center">
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('keyQuery.stats.inputTokens') }}</p>
                  <p class="mt-1 font-mono text-sm font-medium text-gray-900 dark:text-white">{{ (stats?.total_input_tokens ?? 0).toLocaleString() }}</p>
                </div>
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('keyQuery.stats.outputTokens') }}</p>
                  <p class="mt-1 font-mono text-sm font-medium text-gray-900 dark:text-white">{{ (stats?.total_output_tokens ?? 0).toLocaleString() }}</p>
                </div>
                <div>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('keyQuery.stats.cacheTokens') }}</p>
                  <p class="mt-1 font-mono text-sm font-medium text-gray-900 dark:text-white">{{ (stats?.total_cache_tokens ?? 0).toLocaleString() }}</p>
                </div>
              </div>
            </div>

            <!-- Model Breakdown -->
            <div class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('keyQuery.models.title') }}
              </h3>
              <div v-if="models.length === 0" class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('keyQuery.usage.noData') }}
              </div>
              <div v-else class="space-y-3">
                <div
                  v-for="model in models"
                  :key="model.model"
                  class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700"
                >
                  <div class="mb-2 flex items-center justify-between">
                    <span class="font-mono text-sm font-medium text-gray-900 dark:text-white">{{ model.model }}</span>
                    <span class="text-sm font-medium text-gray-700 dark:text-gray-300">${{ (model.cost ?? 0).toFixed(4) }}</span>
                  </div>
                  <div class="mb-2 h-1.5 overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                    <div
                      class="h-full rounded-full bg-primary-500"
                      :style="{ width: modelBarWidth(model.cost ?? 0) }"
                    ></div>
                  </div>
                  <div class="flex gap-4 text-xs text-gray-500 dark:text-gray-400">
                    <span>{{ t('keyQuery.models.requests') }}: {{ (model.requests ?? 0).toLocaleString() }}</span>
                    <span>{{ t('keyQuery.models.tokens') }}: {{ (model.total_tokens ?? 0).toLocaleString() }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Daily Trend -->
            <div class="rounded-xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('keyQuery.trend.title') }}
              </h3>
              <div v-if="trend.length === 0" class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('keyQuery.usage.noData') }}
              </div>
              <div v-else class="space-y-2">
                <div
                  v-for="point in trend"
                  :key="point.date"
                  class="flex items-center gap-3"
                >
                  <span class="w-20 flex-shrink-0 font-mono text-xs text-gray-500 dark:text-gray-400">{{ formatTrendDate(point.date) }}</span>
                  <div class="flex-1">
                    <div class="h-5 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
                      <div
                        class="flex h-full items-center rounded bg-primary-500/20 pl-2"
                        :style="{ width: trendBarWidth(point.cost ?? 0) }"
                      >
                        <span v-if="(point.cost ?? 0) > 0" class="whitespace-nowrap text-xs font-medium text-primary-700 dark:text-primary-300">${{ (point.cost ?? 0).toFixed(4) }}</span>
                      </div>
                    </div>
                  </div>
                  <span class="w-12 flex-shrink-0 text-right font-mono text-xs text-gray-500 dark:text-gray-400">{{ (point.requests ?? 0) }}r</span>
                </div>
              </div>
            </div>
          </template>
        </div>

        <!-- Usage Details Tab -->
        <div v-if="activeTab === 'usage'" class="space-y-4">
          <!-- Usage Loading -->
          <div v-if="usageLoading" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <template v-else>
            <!-- Usage Table -->
            <div class="overflow-x-auto rounded-xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-800">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900/50">
                    <th class="whitespace-nowrap px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('keyQuery.usage.time') }}</th>
                    <th class="whitespace-nowrap px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('keyQuery.usage.model') }}</th>
                    <th class="whitespace-nowrap px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('keyQuery.usage.inputTokens') }}</th>
                    <th class="whitespace-nowrap px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('keyQuery.usage.outputTokens') }}</th>
                    <th class="whitespace-nowrap px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('keyQuery.usage.cacheTokens') }}</th>
                    <th class="whitespace-nowrap px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('keyQuery.usage.cost') }}</th>
                    <th class="whitespace-nowrap px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('keyQuery.usage.duration') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="usageLogs.length === 0">
                    <td colspan="7" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                      {{ t('keyQuery.usage.noData') }}
                    </td>
                  </tr>
                  <tr
                    v-for="log in usageLogs"
                    :key="log.id"
                    class="border-b border-gray-100 transition hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-700/50"
                  >
                    <td class="whitespace-nowrap px-4 py-2.5 font-mono text-xs text-gray-700 dark:text-gray-300">{{ formatDateTime(log.created_at) }}</td>
                    <td class="whitespace-nowrap px-4 py-2.5 font-mono text-xs text-gray-900 dark:text-white">{{ log.model }}</td>
                    <td class="whitespace-nowrap px-4 py-2.5 text-right font-mono text-xs text-gray-700 dark:text-gray-300">{{ (log.input_tokens ?? 0).toLocaleString() }}</td>
                    <td class="whitespace-nowrap px-4 py-2.5 text-right font-mono text-xs text-gray-700 dark:text-gray-300">{{ (log.output_tokens ?? 0).toLocaleString() }}</td>
                    <td class="whitespace-nowrap px-4 py-2.5 text-right font-mono text-xs text-gray-700 dark:text-gray-300">{{ ((log.cache_creation_tokens ?? 0) + (log.cache_read_tokens ?? 0)).toLocaleString() }}</td>
                    <td class="whitespace-nowrap px-4 py-2.5 text-right font-mono text-xs text-gray-700 dark:text-gray-300">${{ (log.actual_cost ?? 0).toFixed(4) }}</td>
                    <td class="whitespace-nowrap px-4 py-2.5 text-right font-mono text-xs text-gray-700 dark:text-gray-300">{{ log.duration_ms != null ? log.duration_ms + 'ms' : '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Pagination -->
            <div v-if="pagination && pagination.pages > 1" class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ pagination.total }} {{ t('keyQuery.requests') }}
              </span>
              <div class="flex items-center gap-2">
                <button
                  @click="goToPage(pagination!.page - 1)"
                  :disabled="pagination.page <= 1"
                  class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm transition disabled:opacity-50 dark:border-dark-600"
                  :class="pagination.page <= 1 ? 'text-gray-400 dark:text-gray-600' : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'"
                >
                  {{ t('keyQuery.usage.prev') }}
                </button>
                <span class="text-sm text-gray-600 dark:text-gray-400">{{ pagination.page }} / {{ pagination.pages }}</span>
                <button
                  @click="goToPage(pagination!.page + 1)"
                  :disabled="pagination.page >= pagination.pages"
                  class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm transition disabled:opacity-50 dark:border-dark-600"
                  :class="pagination.page >= pagination.pages ? 'text-gray-400 dark:text-gray-600' : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'"
                >
                  {{ t('keyQuery.usage.next') }}
                </button>
              </div>
            </div>
          </template>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { availableLocales, setLocale } from '@/i18n'
import {
  queryApiKey,
  queryKeyUsage,
  queryKeyStats,
  queryKeyModels,
  queryKeyTrend,
  type KeyQueryResponse,
  type UsageStatsResponse,
  type ModelStat,
  type TrendPoint,
  type UsageLog,
  type Pagination
} from '@/api/key-query'

const { t, locale } = useI18n()
const appStore = useAppStore()

const STORAGE_KEY = 'key_query_api_key'
const apiKey = ref(localStorage.getItem(STORAGE_KEY) || '')
const querying = ref(false)
const error = ref('')
const result = ref<KeyQueryResponse | null>(null)

// Language dropdown
const showLangDropdown = ref(false)
const langDropdownRef = ref<HTMLElement | null>(null)

const currentLocaleName = computed(() => {
  const loc = availableLocales.find((l) => l.code === locale.value)
  return loc?.name ?? locale.value
})

function switchLocale(code: string) {
  setLocale(code)
  showLangDropdown.value = false
}

function handleClickOutside(e: MouseEvent) {
  if (langDropdownRef.value && !langDropdownRef.value.contains(e.target as Node)) {
    showLangDropdown.value = false
  }
}

// Time range
type TimeRange = 'today' | 'thisMonth' | 'all' | 'custom'
const timeRanges: TimeRange[] = ['today', 'thisMonth', 'all', 'custom']
const selectedTimeRange = ref<TimeRange>('thisMonth')
const customStartDate = ref('')
const customEndDate = ref('')

// Tabs
const activeTab = ref<'overview' | 'usage'>('overview')

// Stats data
const stats = ref<UsageStatsResponse | null>(null)
const models = ref<ModelStat[]>([])
const trend = ref<TrendPoint[]>([])
const statsLoading = ref(false)

// Usage data
const usageLogs = ref<UsageLog[]>([])
const pagination = ref<Pagination | null>(null)
const usageLoading = ref(false)
const currentPage = ref(1)

onMounted(() => {
  appStore.fetchPublicSettings()
  document.addEventListener('click', handleClickOutside)
  // Auto-query if saved key exists
  if (apiKey.value.trim()) {
    handleQuery()
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

function getDateRange(): { start_date?: string; end_date?: string } {
  if (selectedTimeRange.value === 'today') {
    const today = new Date().toISOString().slice(0, 10)
    return { start_date: today, end_date: today }
  }
  if (selectedTimeRange.value === 'thisMonth') {
    const now = new Date()
    const start = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().slice(0, 10)
    const end = now.toISOString().slice(0, 10)
    return { start_date: start, end_date: end }
  }
  if (selectedTimeRange.value === 'custom') {
    const params: { start_date?: string; end_date?: string } = {}
    if (customStartDate.value) params.start_date = customStartDate.value
    if (customEndDate.value) params.end_date = customEndDate.value
    return params
  }
  // 'all'
  return {}
}

async function handleQuery() {
  const key = apiKey.value.trim()
  if (!key) return

  querying.value = true
  error.value = ''
  result.value = null
  stats.value = null
  models.value = []
  trend.value = []
  usageLogs.value = []
  pagination.value = null

  try {
    result.value = await queryApiKey(key)
    localStorage.setItem(STORAGE_KEY, key)
    await fetchTimeRangeData()
  } catch (err: any) {
    error.value = err.message || t('keyQuery.invalidKey')
  } finally {
    querying.value = false
  }
}

async function fetchTimeRangeData() {
  if (!result.value) return
  const key = apiKey.value.trim()
  const dateRange = getDateRange()

  statsLoading.value = true
  usageLoading.value = true
  currentPage.value = 1

  try {
    const [statsRes, modelsRes, trendRes, usageRes] = await Promise.allSettled([
      queryKeyStats({ api_key: key, ...dateRange }),
      queryKeyModels({ api_key: key, ...dateRange }),
      queryKeyTrend({ api_key: key, ...dateRange }),
      queryKeyUsage({ api_key: key, page: 1, page_size: 20, ...dateRange })
    ])

    stats.value = statsRes.status === 'fulfilled' ? statsRes.value : null
    models.value = modelsRes.status === 'fulfilled' ? (modelsRes.value ?? []) : []
    trend.value = trendRes.status === 'fulfilled' ? (trendRes.value ?? []) : []

    if (usageRes.status === 'fulfilled') {
      usageLogs.value = usageRes.value.list ?? []
      pagination.value = usageRes.value.pagination ?? null
    } else {
      usageLogs.value = []
      pagination.value = null
    }
  } catch {
    // Individual errors are handled by allSettled
  } finally {
    statsLoading.value = false
    usageLoading.value = false
  }
}

function setTimeRange(range: TimeRange) {
  selectedTimeRange.value = range
  if (range !== 'custom') {
    fetchTimeRangeData()
  }
}

async function goToPage(page: number) {
  if (!pagination.value || page < 1 || page > pagination.value.pages) return
  const key = apiKey.value.trim()
  const dateRange = getDateRange()

  usageLoading.value = true
  currentPage.value = page

  try {
    const res = await queryKeyUsage({ api_key: key, page, page_size: 20, ...dateRange })
    usageLogs.value = res.list ?? []
    pagination.value = res.pagination ?? null
  } catch {
    // keep current data on error
  } finally {
    usageLoading.value = false
  }
}

// Model bar width relative to max cost
function modelBarWidth(cost: number): string {
  const maxCost = Math.max(...models.value.map((m) => m.cost ?? 0), 0.0001)
  return Math.max((cost / maxCost) * 100, 2) + '%'
}

// Trend bar width relative to max cost
function trendBarWidth(cost: number): string {
  const maxCost = Math.max(...trend.value.map((t) => t.cost ?? 0), 0.0001)
  return Math.max((cost / maxCost) * 100, 2) + '%'
}

function formatTrendDate(dateStr: string): string {
  if (!dateStr) return ''
  return dateStr.slice(5) // Show MM-DD
}

function formatDateTime(dateStr: string): string {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    return d.toLocaleString(undefined, {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch {
    return dateStr
  }
}

function statusClass(status: string): string {
  switch (status) {
    case 'active':
      return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
    case 'inactive':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
    case 'expired':
      return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
    case 'quota_exhausted':
      return 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
  }
}

function statusText(status: string): string {
  switch (status) {
    case 'active':
      return t('keyQuery.statusActive')
    case 'inactive':
      return t('keyQuery.statusInactive')
    case 'expired':
      return t('keyQuery.statusExpired')
    case 'quota_exhausted':
      return t('keyQuery.statusExhausted')
    default:
      return status
  }
}

function quotaBarClass(used: number, total: number): string {
  const percent = (used / total) * 100
  if (percent >= 90) return 'bg-red-500'
  if (percent >= 70) return 'bg-amber-500'
  return 'bg-primary-500'
}

function usagePercent(used: number, limit: number | null): number {
  if (!limit || limit === 0) return 0
  return Math.min((used / limit) * 100, 100)
}
</script>
