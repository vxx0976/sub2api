<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- ==================== LIST VIEW ==================== -->
      <template v-if="!editingSite">
        <!-- Header -->
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.sites.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.sites.description') }}</p>
          </div>
          <button
            @click="showCreateModal = true"
            class="inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
          >
            <Icon name="plus" size="sm" />
            {{ t('reseller.sites.add') }}
          </button>
        </div>

        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-12">
          <LoadingSpinner />
        </div>

        <!-- Sites List -->
        <div v-else-if="domains.length > 0" class="space-y-4">
          <div
            v-for="domain in domains"
            :key="domain.id"
            class="card p-4"
          >
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center gap-3">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-xl"
                  :class="domain.verified
                    ? 'bg-green-100 dark:bg-green-900/30'
                    : 'bg-amber-100 dark:bg-amber-900/30'"
                >
                  <Icon
                    :name="domain.verified ? 'check' : 'clock'"
                    size="md"
                    :class="domain.verified
                      ? 'text-green-600 dark:text-green-400'
                      : 'text-amber-600 dark:text-amber-400'"
                  />
                </div>
                <div>
                  <p class="font-medium text-gray-900 dark:text-white">{{ domain.domain }}</p>
                  <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="domain.verified
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                        : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'"
                    >
                      {{ domain.verified ? t('reseller.domains.verified') : t('reseller.domains.unverified') }}
                    </span>
                    <span v-if="domain.site_name">{{ domain.site_name }}</span>
                  </div>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <!-- Verify Button (only for unverified domains) -->
                <button
                  v-if="!domain.verified"
                  @click="showVerifyInfo(domain)"
                  class="inline-flex items-center gap-1.5 rounded-lg border border-amber-300 px-3 py-1.5 text-xs font-medium text-amber-700 transition-colors hover:bg-amber-50 dark:border-amber-700 dark:text-amber-400 dark:hover:bg-amber-900/20"
                >
                  <Icon name="check" size="xs" />
                  {{ t('reseller.domains.verify') }}
                </button>
                <button
                  @click="openEditView(domain)"
                  class="rounded-lg p-1.5 text-blue-600 transition-colors hover:bg-blue-50 dark:hover:bg-blue-900/20"
                  :title="t('common.edit')"
                >
                  <Icon name="edit" size="sm" />
                </button>
                <button
                  @click="confirmDelete(domain)"
                  class="rounded-lg p-1.5 text-red-600 transition-colors hover:bg-red-50 dark:hover:bg-red-900/20"
                  :title="t('common.delete')"
                >
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </div>

            <!-- Verify Info Box (shown when verify button is clicked) -->
            <div v-if="verifyInfoDomain?.id === domain.id" class="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-900/20">
              <p class="mb-3 text-sm font-medium text-amber-800 dark:text-amber-300">
                {{ t('reseller.domains.verifyInstructions') }}
              </p>
              <div class="space-y-4 text-sm text-amber-700 dark:text-amber-400">
                <!-- Step 1: A Record -->
                <div>
                  <p class="mb-1.5 font-medium">{{ t('reseller.domains.verifyStep1_dns') }}</p>
                  <div class="overflow-hidden rounded-lg border border-amber-200 bg-white dark:border-amber-800 dark:bg-dark-800">
                    <table class="w-full text-xs">
                      <thead>
                        <tr class="border-b border-amber-100 dark:border-amber-800">
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsType') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsHost') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsValue') }}</th>
                          <th class="w-8"></th>
                        </tr>
                      </thead>
                      <tbody class="font-mono">
                        <tr>
                          <td class="px-3 py-2">A</td>
                          <td class="px-3 py-2">@</td>
                          <td class="px-3 py-2">{{ serverIP }}</td>
                          <td class="px-1 py-2">
                            <button @click="copyToClipboard(serverIP)" class="rounded p-1 text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                              <Icon name="copy" size="xs" />
                            </button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p class="mt-1 text-xs text-amber-600/70 dark:text-amber-500/70">{{ t('reseller.domains.verifyStep1_hint') }}</p>
                </div>

                <!-- Step 2: TXT Record -->
                <div>
                  <p class="mb-1.5 font-medium">{{ t('reseller.domains.verifyStep2_txt') }}</p>
                  <div class="overflow-hidden rounded-lg border border-amber-200 bg-white dark:border-amber-800 dark:bg-dark-800">
                    <table class="w-full text-xs">
                      <thead>
                        <tr class="border-b border-amber-100 dark:border-amber-800">
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsType') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsHost') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsValue') }}</th>
                          <th class="w-8"></th>
                        </tr>
                      </thead>
                      <tbody class="font-mono">
                        <tr>
                          <td class="px-3 py-2">TXT</td>
                          <td class="px-3 py-2">_domain-verify</td>
                          <td class="px-3 py-2 break-all">domain-verify={{ domain.verify_token }}</td>
                          <td class="px-1 py-2">
                            <button @click="copyToClipboard('domain-verify=' + (domain.verify_token || ''))" class="rounded p-1 text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                              <Icon name="copy" size="xs" />
                            </button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p class="mt-1 text-xs text-amber-600/70 dark:text-amber-500/70">{{ t('reseller.domains.verifyStep2_hint') }}</p>
                </div>

                <!-- Step 3: Verify -->
                <p class="font-medium">{{ t('reseller.domains.verifyStep3') }}</p>
              </div>
              <button
                @click="verifyDomain(domain)"
                class="mt-3 inline-flex items-center gap-1.5 rounded-lg bg-amber-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-amber-700"
                :disabled="verifying"
              >
                {{ verifying ? t('reseller.domains.verifying') : t('reseller.domains.verifyNow') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
          <Icon name="globe" size="lg" class="mb-3 text-gray-300 dark:text-gray-600" />
          <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.domains.empty') }}</p>
          <button
            @click="showCreateModal = true"
            class="mt-4 inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
          >
            <Icon name="plus" size="sm" />
            {{ t('reseller.sites.add') }}
          </button>
        </div>
      </template>

      <!-- ==================== EDIT VIEW ==================== -->
      <template v-if="editingSite">
        <!-- Header with back button -->
        <div class="flex items-center gap-4">
          <button
            @click="closeEditView"
            class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-dark-800"
          >
            <Icon name="chevronLeft" size="sm" />
            {{ t('reseller.sites.backToList') }}
          </button>
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.sites.editSite') }}</h1>
            <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400">{{ editingSite.domain }}</p>
          </div>
        </div>

        <!-- Tabs -->
        <div class="border-b border-gray-200 dark:border-gray-700">
          <nav class="-mb-px flex space-x-6">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              :class="[
                activeTab === tab.id
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400',
                'whitespace-nowrap border-b-2 px-1 py-3 text-sm font-medium transition-colors'
              ]"
            >
              {{ tab.label }}
            </button>
          </nav>
        </div>

        <!-- Tab Content -->
        <div class="card p-6">
          <!-- Tab 1: Basic Info -->
          <div v-if="activeTab === 'basic'" class="space-y-5">
            <div>
              <label class="label">{{ t('reseller.sites.domainName') }}</label>
              <input
                :value="editingSite.domain"
                type="text"
                class="input bg-gray-50 dark:bg-dark-800"
                disabled
              />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.siteName') }}</label>
              <input
                v-model="editForm.site_name"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.siteNamePlaceholder')"
              />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.subtitle') }}</label>
              <input
                v-model="editForm.subtitle"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.subtitlePlaceholder')"
              />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.siteLogo') }}</label>
              <ImageUpload
                v-model="editForm.site_logo"
                mode="image"
                :upload-label="t('reseller.sites.uploadLogo')"
                :remove-label="t('reseller.sites.removeLogo')"
                :hint="t('reseller.sites.logoHint')"
                :max-size="300 * 1024"
              />
            </div>

            <!-- Default Language -->
            <div>
              <label class="label">{{ t('reseller.sites.defaultLocale') }}</label>
              <select v-model="editForm.default_locale" class="input">
                <option value="">{{ t('reseller.sites.defaultLocaleBrowser') }}</option>
                <option v-for="locale in availableLocales" :key="locale.code" :value="locale.code">
                  {{ locale.name }}
                </option>
              </select>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.sites.defaultLocaleHint') }}</p>
            </div>

            <!-- Verification Status -->
            <div class="rounded-xl border p-4" :class="editingSite.verified ? 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-900/20' : 'border-amber-200 bg-amber-50 dark:border-amber-800 dark:bg-amber-900/20'">
              <div class="flex items-center gap-2">
                <Icon
                  :name="editingSite.verified ? 'check' : 'clock'"
                  size="sm"
                  :class="editingSite.verified ? 'text-green-600 dark:text-green-400' : 'text-amber-600 dark:text-amber-400'"
                />
                <span class="text-sm font-medium" :class="editingSite.verified ? 'text-green-800 dark:text-green-300' : 'text-amber-800 dark:text-amber-300'">
                  {{ editingSite.verified ? t('reseller.domains.verified') : t('reseller.domains.unverified') }}
                </span>
              </div>
              <div v-if="!editingSite.verified" class="mt-3">
                <button
                  @click="showVerifyInfo(editingSite)"
                  class="inline-flex items-center gap-1.5 rounded-lg bg-amber-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-amber-700"
                >
                  <Icon name="check" size="xs" />
                  {{ t('reseller.domains.verify') }}
                </button>
              </div>
            </div>

            <!-- Inline Verify Info (in edit view) -->
            <div v-if="verifyInfoDomain?.id === editingSite.id" class="rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-900/20">
              <p class="mb-3 text-sm font-medium text-amber-800 dark:text-amber-300">
                {{ t('reseller.domains.verifyInstructions') }}
              </p>
              <div class="space-y-4 text-sm text-amber-700 dark:text-amber-400">
                <div>
                  <p class="mb-1.5 font-medium">{{ t('reseller.domains.verifyStep1_dns') }}</p>
                  <div class="overflow-hidden rounded-lg border border-amber-200 bg-white dark:border-amber-800 dark:bg-dark-800">
                    <table class="w-full text-xs">
                      <thead>
                        <tr class="border-b border-amber-100 dark:border-amber-800">
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsType') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsHost') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsValue') }}</th>
                          <th class="w-8"></th>
                        </tr>
                      </thead>
                      <tbody class="font-mono">
                        <tr>
                          <td class="px-3 py-2">A</td>
                          <td class="px-3 py-2">@</td>
                          <td class="px-3 py-2">{{ serverIP }}</td>
                          <td class="px-1 py-2">
                            <button @click="copyToClipboard(serverIP)" class="rounded p-1 text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                              <Icon name="copy" size="xs" />
                            </button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p class="mt-1 text-xs text-amber-600/70 dark:text-amber-500/70">{{ t('reseller.domains.verifyStep1_hint') }}</p>
                </div>
                <div>
                  <p class="mb-1.5 font-medium">{{ t('reseller.domains.verifyStep2_txt') }}</p>
                  <div class="overflow-hidden rounded-lg border border-amber-200 bg-white dark:border-amber-800 dark:bg-dark-800">
                    <table class="w-full text-xs">
                      <thead>
                        <tr class="border-b border-amber-100 dark:border-amber-800">
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsType') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsHost') }}</th>
                          <th class="px-3 py-1.5 text-left text-gray-500">{{ t('reseller.domains.dnsValue') }}</th>
                          <th class="w-8"></th>
                        </tr>
                      </thead>
                      <tbody class="font-mono">
                        <tr>
                          <td class="px-3 py-2">TXT</td>
                          <td class="px-3 py-2">_domain-verify</td>
                          <td class="px-3 py-2 break-all">domain-verify={{ editingSite.verify_token }}</td>
                          <td class="px-1 py-2">
                            <button @click="copyToClipboard('domain-verify=' + (editingSite.verify_token || ''))" class="rounded p-1 text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
                              <Icon name="copy" size="xs" />
                            </button>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </div>
                  <p class="mt-1 text-xs text-amber-600/70 dark:text-amber-500/70">{{ t('reseller.domains.verifyStep2_hint') }}</p>
                </div>
                <p class="font-medium">{{ t('reseller.domains.verifyStep3') }}</p>
              </div>
              <button
                @click="verifyDomain(editingSite)"
                class="mt-3 inline-flex items-center gap-1.5 rounded-lg bg-amber-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-amber-700"
                :disabled="verifying"
              >
                {{ verifying ? t('reseller.domains.verifying') : t('reseller.domains.verifyNow') }}
              </button>
            </div>

          </div>

          <!-- Tab 3: Homepage -->
          <div v-if="activeTab === 'homepage'" class="space-y-6">
            <div>
              <p class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('reseller.sites.homeTemplate') }}</p>
              <!-- Template card grid: 3 cols on mobile, 5 on sm+ -->
              <div class="grid grid-cols-3 gap-2 sm:grid-cols-5">
                <button
                  v-for="tpl in homeTemplates"
                  :key="tpl.id"
                  type="button"
                  @click="editForm.home_template = tpl.id; previewTemplate = tpl.id"
                  :class="[
                    'group relative flex flex-col overflow-hidden rounded-xl border-2 text-left transition-all',
                    editForm.home_template === tpl.id
                      ? 'border-primary-500 shadow-md'
                      : 'border-gray-200 hover:border-gray-300 dark:border-dark-600 dark:hover:border-dark-500'
                  ]"
                >
                  <!-- Thumbnail -->
                  <div class="relative h-20 w-full overflow-hidden">
                    <template v-if="tpl.id === 'default'">
                      <div class="h-full bg-slate-50 dark:bg-dark-900">
                        <div class="flex items-center justify-between bg-white px-2 py-1 shadow-sm dark:bg-dark-800">
                          <div class="h-1.5 w-8 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                          <div class="h-3.5 w-7 rounded bg-primary-500 text-[5px] leading-[14px] text-white text-center">Login</div>
                        </div>
                        <div class="flex flex-col items-center gap-0.5 px-2 pt-2">
                          <div class="h-2 w-20 rounded-full bg-slate-700 dark:bg-slate-300"></div>
                          <div class="h-1 w-24 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                          <div class="mt-1 h-3.5 w-14 rounded bg-primary-500"></div>
                        </div>
                        <div class="mt-1.5 grid grid-cols-3 gap-1 px-2">
                          <div v-for="i in 3" :key="i" class="rounded bg-white p-1 shadow-sm dark:bg-dark-800">
                            <div class="mb-0.5 h-2 w-2 rounded-sm bg-primary-200 dark:bg-primary-900"></div>
                            <div class="h-0.5 w-full rounded-full bg-slate-200 dark:bg-dark-700"></div>
                          </div>
                        </div>
                      </div>
                    </template>
                    <template v-else-if="tpl.id === 'hero'">
                      <div class="flex h-full flex-col bg-gradient-to-br from-slate-900 via-indigo-900 to-slate-800">
                        <div class="flex items-center justify-between px-2 py-1">
                          <div class="h-1 w-7 rounded-full bg-white/40"></div>
                          <div class="h-3 w-6 rounded bg-white/20 text-[5px] leading-3 text-white/70 text-center">Login</div>
                        </div>
                        <div class="flex flex-1 flex-col items-center justify-center gap-1 px-2 text-center">
                          <div class="h-2 w-20 rounded-full bg-white/90"></div>
                          <div class="h-1 w-24 rounded-full bg-white/40"></div>
                          <div class="mt-1 flex gap-1">
                            <div class="h-3.5 w-10 rounded bg-white/90"></div>
                            <div class="h-3.5 w-8 rounded border border-white/30 bg-white/10"></div>
                          </div>
                        </div>
                      </div>
                    </template>
                    <template v-else-if="tpl.id === 'brand'">
                      <div class="flex h-full flex-col bg-white dark:bg-dark-900">
                        <div class="flex items-center justify-between px-2 py-1 shadow-sm">
                          <div class="h-1 w-7 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                          <div class="h-3 w-6 rounded bg-primary-100 text-[5px] leading-3 text-primary-600 text-center dark:bg-primary-900 dark:text-primary-400">Login</div>
                        </div>
                        <div class="flex flex-1 flex-col items-center justify-center gap-1.5 px-2 text-center">
                          <div class="h-7 w-7 rounded-lg bg-primary-100 dark:bg-primary-900/50"></div>
                          <div class="h-1.5 w-16 rounded-full bg-slate-700 dark:bg-slate-300"></div>
                          <div class="h-3.5 w-20 rounded-lg bg-primary-500"></div>
                        </div>
                      </div>
                    </template>
                    <template v-else-if="tpl.id === 'custom_html'">
                      <div class="h-full bg-slate-900 p-2 font-mono">
                        <div class="mb-1 flex gap-0.5">
                          <div class="h-1.5 w-1.5 rounded-full bg-red-500/70"></div>
                          <div class="h-1.5 w-1.5 rounded-full bg-yellow-500/70"></div>
                          <div class="h-1.5 w-1.5 rounded-full bg-green-500/70"></div>
                        </div>
                        <div class="space-y-0.5">
                          <div><span class="text-[6px] text-purple-400">&lt;html&gt;</span></div>
                          <div class="ml-2"><span class="text-[6px] text-blue-400">&lt;body&gt;</span></div>
                          <div class="ml-4 h-1 w-16 rounded-full bg-green-500/40"></div>
                          <div class="ml-4 h-1 w-12 rounded-full bg-yellow-500/40"></div>
                          <div class="ml-4 h-1 w-14 rounded-full bg-green-500/40"></div>
                          <div class="ml-2"><span class="text-[6px] text-blue-400">&lt;/body&gt;</span></div>
                          <div><span class="text-[6px] text-purple-400">&lt;/html&gt;</span></div>
                        </div>
                      </div>
                    </template>
                    <template v-else-if="tpl.id === 'external_url'">
                      <div class="h-full bg-slate-100 dark:bg-dark-800">
                        <div class="flex items-center gap-1 bg-slate-200 px-2 py-1 dark:bg-dark-700">
                          <div class="h-1 w-1 rounded-full bg-red-400"></div>
                          <div class="h-1 w-1 rounded-full bg-yellow-400"></div>
                          <div class="h-1 w-1 rounded-full bg-green-400"></div>
                          <div class="ml-1 flex flex-1 items-center rounded bg-white px-1 py-0.5 text-[5px] text-slate-400 dark:bg-dark-900">
                            https://your-site.com
                          </div>
                        </div>
                        <div class="flex h-[calc(100%-20px)] flex-col items-center justify-center gap-1.5">
                          <svg class="h-6 w-6 text-slate-400 dark:text-dark-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18M9 21V9"/></svg>
                          <div class="h-1 w-14 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                        </div>
                      </div>
                    </template>
                  </div>
                  <!-- Label -->
                  <div class="flex items-center justify-between px-2 py-1.5" :class="editForm.home_template === tpl.id ? 'bg-primary-50 dark:bg-primary-900/20' : 'bg-white dark:bg-dark-800'">
                    <p class="truncate text-[10px] font-semibold" :class="editForm.home_template === tpl.id ? 'text-primary-700 dark:text-primary-400' : 'text-gray-700 dark:text-gray-300'">{{ tpl.name }}</p>
                    <div v-if="editForm.home_template === tpl.id" class="ml-1 flex h-3.5 w-3.5 flex-shrink-0 items-center justify-center rounded-full bg-primary-500">
                      <svg class="h-2 w-2" viewBox="0 0 10 10" fill="none"><path d="M2 5l2.5 2.5L8 3" stroke="white" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    </div>
                  </div>
                </button>
              </div>

              <!-- Preview Panel -->
              <div v-if="previewTemplate" class="mt-3 overflow-hidden rounded-xl border border-gray-200 dark:border-dark-600">
                <div class="flex items-center justify-between bg-gray-50 px-3 py-2 dark:bg-dark-800/60">
                  <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ t('reseller.sites.templatePreview') }}：{{ currentPreviewTpl?.name }}</span>
                  <span class="text-[11px] text-gray-400 dark:text-gray-500">{{ currentPreviewTpl?.hint }}</span>
                </div>
                <div class="relative h-52 w-full overflow-hidden">
                  <!-- Default -->
                  <template v-if="previewTemplate === 'default'">
                    <div class="h-full bg-slate-50 dark:bg-dark-900">
                      <div class="flex items-center justify-between bg-white px-5 py-2.5 shadow-sm dark:bg-dark-800">
                        <div class="h-3 w-20 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                        <div class="flex gap-2">
                          <div class="h-2.5 w-14 rounded-full bg-slate-200 dark:bg-dark-700"></div>
                          <div class="h-2.5 w-14 rounded-full bg-slate-200 dark:bg-dark-700"></div>
                          <div class="h-7 w-16 rounded bg-primary-500 text-[9px] leading-7 text-white text-center">Login</div>
                        </div>
                      </div>
                      <div class="flex flex-col items-center gap-2 px-6 pt-5">
                        <div class="h-4 w-52 rounded-full bg-slate-700 dark:bg-slate-300"></div>
                        <div class="h-2.5 w-64 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                        <div class="mt-2 h-8 w-32 rounded-lg bg-primary-500"></div>
                      </div>
                      <div class="mt-4 grid grid-cols-3 gap-3 px-5">
                        <div v-for="i in 3" :key="i" class="rounded-lg bg-white p-3 shadow-sm dark:bg-dark-800">
                          <div class="mb-2 h-5 w-5 rounded bg-primary-200 dark:bg-primary-900"></div>
                          <div class="h-2 w-full rounded-full bg-slate-200 dark:bg-dark-700"></div>
                          <div class="mt-1 h-1.5 w-3/4 rounded-full bg-slate-100 dark:bg-dark-700"></div>
                        </div>
                      </div>
                    </div>
                  </template>
                  <!-- Hero -->
                  <template v-else-if="previewTemplate === 'hero'">
                    <div class="flex h-full flex-col bg-gradient-to-br from-slate-900 via-indigo-900 to-slate-800">
                      <div class="flex items-center justify-between px-5 py-3">
                        <div class="h-2.5 w-16 rounded-full bg-white/40"></div>
                        <div class="h-6 w-14 rounded bg-white/20 text-[9px] leading-6 text-white/70 text-center">Login</div>
                      </div>
                      <div class="flex flex-1 flex-col items-center justify-center gap-2.5 px-6 text-center">
                        <div class="h-4 w-56 rounded-full bg-white/90"></div>
                        <div class="h-2.5 w-72 rounded-full bg-white/40"></div>
                        <div class="mt-2 flex gap-2.5">
                          <div class="h-8 w-24 rounded-lg bg-white/90"></div>
                          <div class="h-8 w-20 rounded-lg border border-white/30 bg-white/10"></div>
                        </div>
                        <div class="mt-1.5 flex flex-wrap justify-center gap-1.5">
                          <div v-for="i in 4" :key="i" class="h-5 w-16 rounded-full border border-white/20 bg-white/10"></div>
                        </div>
                      </div>
                    </div>
                  </template>
                  <!-- Brand -->
                  <template v-else-if="previewTemplate === 'brand'">
                    <div class="flex h-full flex-col bg-white dark:bg-dark-900">
                      <div class="flex items-center justify-between px-5 py-2.5 shadow-sm">
                        <div class="h-2.5 w-16 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                        <div class="h-6 w-14 rounded bg-primary-100 text-[9px] leading-6 text-primary-600 text-center dark:bg-primary-900 dark:text-primary-400">Login</div>
                      </div>
                      <div class="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
                        <div class="h-14 w-14 rounded-2xl bg-primary-100 dark:bg-primary-900/50"></div>
                        <div class="h-4 w-40 rounded-full bg-slate-700 dark:bg-slate-300"></div>
                        <div class="h-2.5 w-52 rounded-full bg-slate-200 dark:bg-dark-700"></div>
                        <div class="h-9 w-48 rounded-xl bg-primary-500"></div>
                      </div>
                    </div>
                  </template>
                  <!-- Custom HTML -->
                  <template v-else-if="previewTemplate === 'custom_html'">
                    <div class="h-full bg-slate-900 p-5 font-mono">
                      <div class="mb-3 flex gap-1.5">
                        <div class="h-3 w-3 rounded-full bg-red-500/70"></div>
                        <div class="h-3 w-3 rounded-full bg-yellow-500/70"></div>
                        <div class="h-3 w-3 rounded-full bg-green-500/70"></div>
                      </div>
                      <div class="space-y-1.5">
                        <div><span class="text-xs text-purple-400">&lt;html&gt;</span></div>
                        <div class="ml-5 flex items-center gap-2"><span class="text-xs text-blue-400">&lt;head&gt;</span><div class="h-2 w-24 rounded-full bg-slate-600"></div><span class="text-xs text-blue-400">&lt;/head&gt;</span></div>
                        <div class="ml-5"><span class="text-xs text-blue-400">&lt;body&gt;</span></div>
                        <div class="ml-10 h-2 w-40 rounded-full bg-green-500/40"></div>
                        <div class="ml-10 h-2 w-28 rounded-full bg-yellow-500/40"></div>
                        <div class="ml-10 h-2 w-36 rounded-full bg-green-500/40"></div>
                        <div class="ml-5"><span class="text-xs text-blue-400">&lt;/body&gt;</span></div>
                        <div><span class="text-xs text-purple-400">&lt;/html&gt;</span></div>
                      </div>
                    </div>
                  </template>
                  <!-- External URL -->
                  <template v-else-if="previewTemplate === 'external_url'">
                    <div class="h-full bg-slate-100 dark:bg-dark-800">
                      <div class="flex items-center gap-2 bg-slate-200 px-4 py-2.5 dark:bg-dark-700">
                        <div class="h-2.5 w-2.5 rounded-full bg-red-400"></div>
                        <div class="h-2.5 w-2.5 rounded-full bg-yellow-400"></div>
                        <div class="h-2.5 w-2.5 rounded-full bg-green-400"></div>
                        <div class="ml-2 flex flex-1 items-center gap-1.5 rounded-md bg-white px-2.5 py-1 text-[10px] text-slate-400 dark:bg-dark-900 dark:text-dark-400">
                          <svg class="h-3 w-3 flex-shrink-0" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="7"/></svg>
                          https://your-site.com
                        </div>
                      </div>
                      <div class="flex h-[calc(100%-38px)] flex-col items-center justify-center gap-3">
                        <svg class="h-14 w-14 text-slate-400 dark:text-dark-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18M9 21V9"/></svg>
                        <div class="h-2.5 w-36 rounded-full bg-slate-300 dark:bg-dark-600"></div>
                        <div class="h-2 w-24 rounded-full bg-slate-200 dark:bg-dark-700"></div>
                      </div>
                    </div>
                  </template>
                </div>
              </div>
            </div>

            <!-- Conditional content input -->
            <div v-if="editForm.home_template === 'custom_html'">
              <label class="label">{{ t('reseller.sites.homeContent') }}</label>
              <textarea v-model="editForm.home_content" class="input font-mono text-sm" rows="10" :placeholder="t('reseller.sites.homeContentHtmlPlaceholder')" />
            </div>
            <div v-else-if="editForm.home_template === 'external_url'">
              <label class="label">{{ t('reseller.sites.homeContent') }}</label>
              <input v-model="editForm.home_content" type="text" class="input" :placeholder="t('reseller.sites.homeContentUrlPlaceholder')" />
            </div>

            <!-- Preview button -->
            <div v-if="editingSite?.verified">
              <button @click="previewSite" class="btn btn-secondary inline-flex items-center gap-2">
                <Icon name="externalLink" size="sm" />
                {{ t('reseller.sites.preview') }}
              </button>
            </div>
          </div>

          <!-- Tab 3: Docs -->
          <div v-if="activeTab === 'docs'" class="space-y-5">
            <div>
              <label class="label">{{ t('reseller.sites.docUrl') }}</label>
              <input
                v-model="editForm.doc_url"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.docUrlPlaceholder')"
              />
            </div>
          </div>

          <!-- Tab 4: Purchase -->
          <div v-if="activeTab === 'purchase'" class="space-y-5">
            <div class="space-y-3">
              <div class="flex items-center gap-3">
                <input
                  v-model="editForm.purchase_enabled"
                  type="checkbox"
                  id="purchase-enabled"
                  class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                />
                <label for="purchase-enabled" class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ t('reseller.sites.purchaseEnabled') }}
                </label>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.sites.purchaseEnabledHint') }}</p>
              <div v-if="editForm.purchase_enabled">
                <label class="label">{{ t('reseller.sites.purchaseUrl') }}</label>
                <input
                  v-model="editForm.purchase_url"
                  type="text"
                  class="input"
                  :placeholder="t('reseller.sites.purchaseUrlPlaceholder')"
                />
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.sites.purchaseUrlHint') }}</p>
              </div>
            </div>
          </div>

          <!-- Tab 5: SEO -->
          <div v-if="activeTab === 'seo'" class="space-y-5">
            <div>
              <label class="label">{{ t('reseller.sites.seoTitle') }}</label>
              <input
                v-model="editForm.seo_title"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.seoTitlePlaceholder')"
              />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.seoDescription') }}</label>
              <textarea
                v-model="editForm.seo_description"
                class="input"
                rows="3"
                :placeholder="t('reseller.sites.seoDescriptionPlaceholder')"
              />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.seoKeywords') }}</label>
              <input
                v-model="editForm.seo_keywords"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.seoKeywordsPlaceholder')"
              />
            </div>
          </div>
        </div>

        <!-- Save Button -->
        <div class="flex justify-end">
          <button
            @click="saveEditForm"
            class="btn btn-primary"
            :disabled="submitting"
          >
            {{ submitting ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </div>

    <!-- Create Site Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeCreateModal">
        <div class="mx-4 w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-900">
          <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('reseller.sites.addTitle') }}
          </h3>
          <form @submit.prevent="createDomain" class="space-y-4">
            <div>
              <label class="label">{{ t('reseller.sites.domainName') }}</label>
              <input v-model="createForm.domain" type="text" class="input" placeholder="example.com" required />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.siteName') }}</label>
              <input v-model="createForm.site_name" type="text" class="input" :placeholder="t('reseller.sites.siteNamePlaceholder')" />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.subtitle') }}</label>
              <input v-model="createForm.subtitle" type="text" class="input" :placeholder="t('reseller.sites.subtitlePlaceholder')" />
            </div>
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" @click="closeCreateModal" class="btn btn-secondary">{{ t('common.cancel') }}</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? t('common.saving') : t('common.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Delete Confirmation Modal -->
    <Teleport to="body">
      <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showDeleteModal = false">
        <div class="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-900">
          <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('common.confirmDelete') }}</h3>
          <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">
            {{ t('reseller.domains.deleteConfirm', { domain: deleteTarget?.domain }) }}
          </p>
          <div class="flex justify-end gap-3">
            <button @click="showDeleteModal = false" class="btn btn-secondary">{{ t('common.cancel') }}</button>
            <button @click="deleteDomain" class="btn btn-danger" :disabled="submitting">
              {{ submitting ? t('common.deleting') : t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { ResellerDomain } from '@/api/reseller/domains'
import { useAppStore } from '@/stores'
import { availableLocales } from '@/i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import ImageUpload from '@/components/common/ImageUpload.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Server IP for A record guidance
const serverIP = ref('')

async function loadServerInfo() {
  try {
    const info = await resellerAPI.domains.getServerInfo()
    serverIP.value = info.ip || info.domain
  } catch {
    serverIP.value = window.location.hostname
  }
}

// List state
const loading = ref(true)
const domains = ref<ResellerDomain[]>([])

// Edit view state
const editingSite = ref<ResellerDomain | null>(null)
const activeTab = ref('basic')
const previewTemplate = ref('default')

// Home template definitions
const homeTemplates = computed(() => [
  { id: 'default',      name: t('reseller.sites.homeTemplateDefault'),  hint: t('reseller.sites.homeTemplateDefaultHint')  },
  { id: 'hero',         name: t('reseller.sites.homeTemplateHero'),     hint: t('reseller.sites.homeTemplateHeroHint')     },
  { id: 'brand',        name: t('reseller.sites.homeTemplateBrand'),    hint: t('reseller.sites.homeTemplateBrandHint')    },
  { id: 'custom_html',  name: t('reseller.sites.homeTemplateCustom'),   hint: t('reseller.sites.homeTemplateCustomHint')   },
  { id: 'external_url', name: t('reseller.sites.homeTemplateExternal'), hint: t('reseller.sites.homeTemplateExternalHint') },
])

const currentPreviewTpl = computed(() => homeTemplates.value.find(t => t.id === previewTemplate.value))

// Modal state
const showCreateModal = ref(false)
const showDeleteModal = ref(false)
const submitting = ref(false)
const deleteTarget = ref<ResellerDomain | null>(null)
const verifyInfoDomain = ref<ResellerDomain | null>(null)
const verifying = ref(false)

// Tabs definition
const tabs = computed(() => [
  { id: 'basic', label: t('reseller.sites.tabs.basic') },
  { id: 'homepage', label: t('reseller.sites.tabs.homepage') },
  { id: 'docs', label: t('reseller.sites.tabs.docs') },
  { id: 'purchase', label: t('reseller.sites.tabs.purchase') },
  { id: 'seo', label: t('reseller.sites.tabs.seo') }
])

// Create form (simplified)
const createForm = ref({
  domain: '',
  site_name: '',
  subtitle: ''
})

// Edit form (all fields)
const editForm = ref({
  site_name: '',
  site_logo: '',
  subtitle: '',
  doc_url: '',
  home_content: '',
  home_template: 'default',
  purchase_enabled: false,
  purchase_url: '',
  default_locale: '',
  seo_title: '',
  seo_description: '',
  seo_keywords: ''
})

async function loadDomains() {
  loading.value = true
  try {
    const result = await resellerAPI.domains.list(1, 100)
    domains.value = result.items || []
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

function openEditView(domain: ResellerDomain) {
  editingSite.value = domain
  activeTab.value = 'basic'
  verifyInfoDomain.value = null
  previewTemplate.value = domain.home_template || 'default'
  editForm.value = {
    site_name: domain.site_name || '',
    site_logo: domain.site_logo || '',
    subtitle: domain.subtitle || '',
    doc_url: domain.doc_url || '',
    home_content: domain.home_content || '',
    home_template: domain.home_template || 'default',
    purchase_enabled: domain.purchase_enabled || false,
    purchase_url: domain.purchase_url || '',
    default_locale: domain.default_locale || '',
    seo_title: domain.seo_title || '',
    seo_description: domain.seo_description || '',
    seo_keywords: domain.seo_keywords || ''
  }
}

function closeEditView() {
  editingSite.value = null
  verifyInfoDomain.value = null
}

function confirmDelete(domain: ResellerDomain) {
  deleteTarget.value = domain
  showDeleteModal.value = true
}

function showVerifyInfo(domain: ResellerDomain) {
  verifyInfoDomain.value = verifyInfoDomain.value?.id === domain.id ? null : domain
}

function closeCreateModal() {
  showCreateModal.value = false
  createForm.value = { domain: '', site_name: '', subtitle: '' }
}

function previewSite() {
  if (editingSite.value) {
    window.open(`https://${editingSite.value.domain}`, '_blank')
  }
}

async function createDomain() {
  submitting.value = true
  try {
    await resellerAPI.domains.create({
      domain: createForm.value.domain,
      site_name: createForm.value.site_name || undefined,
      subtitle: createForm.value.subtitle || undefined
    })
    appStore.showSuccess(t('reseller.domains.createSuccess'))
    closeCreateModal()
    loadDomains()
  } catch (error: any) {
    appStore.showError(error.message || t('common.saveFailed'))
  } finally {
    submitting.value = false
  }
}

async function saveEditForm() {
  if (!editingSite.value) return
  submitting.value = true
  try {
    const updated = await resellerAPI.domains.update(editingSite.value.id, {
      site_name: editForm.value.site_name,
      site_logo: editForm.value.site_logo,
      subtitle: editForm.value.subtitle,
      doc_url: editForm.value.doc_url,
      home_content: editForm.value.home_content,
      home_template: editForm.value.home_template,
      purchase_enabled: editForm.value.purchase_enabled,
      purchase_url: editForm.value.purchase_url,
      default_locale: editForm.value.default_locale,
      seo_title: editForm.value.seo_title,
      seo_description: editForm.value.seo_description,
      seo_keywords: editForm.value.seo_keywords
    })
    // Update the editingSite with fresh data
    editingSite.value = updated
    appStore.showSuccess(t('common.saveSuccess'))
    loadDomains()
  } catch (error: any) {
    appStore.showError(error.message || t('common.saveFailed'))
  } finally {
    submitting.value = false
  }
}

async function deleteDomain() {
  if (!deleteTarget.value) return
  submitting.value = true
  try {
    await resellerAPI.domains.delete(deleteTarget.value.id)
    appStore.showSuccess(t('reseller.domains.deleteSuccess'))
    showDeleteModal.value = false
    // If we deleted the currently editing site, go back to list
    if (editingSite.value?.id === deleteTarget.value.id) {
      closeEditView()
    }
    loadDomains()
  } catch (error: any) {
    appStore.showError(error.message || t('common.saveFailed'))
  } finally {
    submitting.value = false
  }
}

async function verifyDomain(domain: ResellerDomain) {
  verifying.value = true
  try {
    const updated = await resellerAPI.domains.verify(domain.id)
    appStore.showSuccess(t('reseller.domains.verifySuccess'))
    verifyInfoDomain.value = null
    // If verifying from edit view, update the editingSite
    if (editingSite.value?.id === domain.id) {
      editingSite.value = updated
    }
    loadDomains()
  } catch (error: any) {
    appStore.showError(error.message || t('reseller.domains.verifyFailed'))
  } finally {
    verifying.value = false
  }
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(t('common.copied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

onMounted(() => {
  loadDomains()
  loadServerInfo()
})
</script>
