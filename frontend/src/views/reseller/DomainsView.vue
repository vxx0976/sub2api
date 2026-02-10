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
                          <td class="px-3 py-2">{{ serverHost }}</td>
                          <td class="px-1 py-2">
                            <button @click="copyToClipboard(serverHost)" class="rounded p-1 text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
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
                          <td class="px-3 py-2">{{ serverHost }}</td>
                          <td class="px-1 py-2">
                            <button @click="copyToClipboard(serverHost)" class="rounded p-1 text-gray-400 hover:text-gray-700 dark:hover:text-gray-300">
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

          <!-- Tab 2: Appearance -->
          <div v-if="activeTab === 'appearance'" class="space-y-5">
            <div>
              <label class="label">{{ t('reseller.sites.siteLogo') }}</label>
              <input
                v-model="editForm.site_logo"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.siteLogoPlaceholder')"
              />
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.brandColor') }}</label>
              <div class="flex items-center gap-2">
                <input v-model="editForm.brand_color" type="color" class="h-10 w-10 cursor-pointer rounded-lg border-0" />
                <input v-model="editForm.brand_color" type="text" class="input flex-1" placeholder="#4F46E5" />
              </div>
            </div>
            <div>
              <label class="label">{{ t('reseller.sites.customCSS') }}</label>
              <textarea
                v-model="editForm.custom_css"
                class="input font-mono text-sm"
                rows="6"
                :placeholder="t('reseller.sites.customCSSPlaceholder')"
              />
            </div>
          </div>

          <!-- Tab 3: Homepage -->
          <div v-if="activeTab === 'homepage'" class="space-y-5">
            <div>
              <label class="label">{{ t('reseller.sites.homeTemplate') }}</label>
              <div class="space-y-3">
                <label class="flex items-start gap-3 rounded-lg border border-gray-200 p-3 cursor-pointer transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-dark-800" :class="{ 'border-primary-500 bg-primary-50 dark:bg-primary-900/10': editForm.home_template === '' || editForm.home_template === 'default' }">
                  <input v-model="editForm.home_template" type="radio" value="default" class="mt-0.5" />
                  <div>
                    <span class="text-sm font-medium text-gray-900 dark:text-white">{{ t('reseller.sites.homeTemplateDefault') }}</span>
                    <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.sites.homeTemplateDefaultHint') }}</p>
                  </div>
                </label>
                <label class="flex items-start gap-3 rounded-lg border border-gray-200 p-3 cursor-pointer transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-dark-800" :class="{ 'border-primary-500 bg-primary-50 dark:bg-primary-900/10': editForm.home_template === 'minimal' }">
                  <input v-model="editForm.home_template" type="radio" value="minimal" class="mt-0.5" />
                  <div>
                    <span class="text-sm font-medium text-gray-900 dark:text-white">{{ t('reseller.sites.homeTemplateMinimal') }}</span>
                    <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.sites.homeTemplateMinimalHint') }}</p>
                  </div>
                </label>
                <label class="flex items-start gap-3 rounded-lg border border-gray-200 p-3 cursor-pointer transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-dark-800" :class="{ 'border-primary-500 bg-primary-50 dark:bg-primary-900/10': editForm.home_template === 'custom_html' }">
                  <input v-model="editForm.home_template" type="radio" value="custom_html" class="mt-0.5" />
                  <div>
                    <span class="text-sm font-medium text-gray-900 dark:text-white">{{ t('reseller.sites.homeTemplateCustom') }}</span>
                  </div>
                </label>
                <label class="flex items-start gap-3 rounded-lg border border-gray-200 p-3 cursor-pointer transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-dark-800" :class="{ 'border-primary-500 bg-primary-50 dark:bg-primary-900/10': editForm.home_template === 'external_url' }">
                  <input v-model="editForm.home_template" type="radio" value="external_url" class="mt-0.5" />
                  <div>
                    <span class="text-sm font-medium text-gray-900 dark:text-white">{{ t('reseller.sites.homeTemplateExternal') }}</span>
                  </div>
                </label>
              </div>
            </div>

            <!-- Conditional content area -->
            <div v-if="editForm.home_template === 'custom_html'">
              <label class="label">{{ t('reseller.sites.homeContent') }}</label>
              <textarea
                v-model="editForm.home_content"
                class="input font-mono text-sm"
                rows="10"
                :placeholder="t('reseller.sites.homeContentHtmlPlaceholder')"
              />
            </div>
            <div v-else-if="editForm.home_template === 'external_url'">
              <label class="label">{{ t('reseller.sites.homeContent') }}</label>
              <input
                v-model="editForm.home_content"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.homeContentUrlPlaceholder')"
              />
            </div>
            <div v-else class="rounded-lg bg-gray-50 p-4 text-sm text-gray-500 dark:bg-dark-800 dark:text-gray-400">
              {{ t('reseller.sites.homeTemplateDefaultHint') }}
            </div>

            <!-- Preview button -->
            <div>
              <button
                @click="previewSite"
                class="btn btn-secondary inline-flex items-center gap-2"
              >
                <Icon name="externalLink" size="sm" />
                {{ t('reseller.sites.preview') }}
              </button>
            </div>
          </div>

          <!-- Tab 4: Features & SEO -->
          <div v-if="activeTab === 'features'" class="space-y-5">
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

            <!-- Documentation URL -->
            <div>
              <label class="label">{{ t('reseller.sites.docUrl') }}</label>
              <input
                v-model="editForm.doc_url"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.docUrlPlaceholder')"
              />
            </div>

            <!-- Purchase Page -->
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

            <!-- Login Redirect -->
            <div>
              <label class="label">{{ t('reseller.sites.loginRedirect') }}</label>
              <input
                v-model="editForm.login_redirect"
                type="text"
                class="input"
                :placeholder="t('reseller.sites.loginRedirectPlaceholder')"
              />
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.sites.loginRedirectHint') }}</p>
            </div>

            <!-- SEO Section -->
            <div class="border-t border-gray-200 pt-5 dark:border-gray-700">
              <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">SEO</h3>
              <div class="space-y-4">
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

const { t } = useI18n()
const appStore = useAppStore()

// Server host for A record guidance
const serverHost = computed(() => window.location.hostname)

// List state
const loading = ref(true)
const domains = ref<ResellerDomain[]>([])

// Edit view state
const editingSite = ref<ResellerDomain | null>(null)
const activeTab = ref('basic')

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
  { id: 'appearance', label: t('reseller.sites.tabs.appearance') },
  { id: 'homepage', label: t('reseller.sites.tabs.homepage') },
  { id: 'features', label: t('reseller.sites.tabs.features') }
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
  brand_color: '#4F46E5',
  custom_css: '',
  doc_url: '',
  home_content: '',
  home_template: 'default',
  purchase_enabled: false,
  purchase_url: '',
  default_locale: '',
  seo_title: '',
  seo_description: '',
  seo_keywords: '',
  login_redirect: ''
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
  editForm.value = {
    site_name: domain.site_name || '',
    site_logo: domain.site_logo || '',
    subtitle: domain.subtitle || '',
    brand_color: domain.brand_color || '#4F46E5',
    custom_css: domain.custom_css || '',
    doc_url: domain.doc_url || '',
    home_content: domain.home_content || '',
    home_template: domain.home_template || 'default',
    purchase_enabled: domain.purchase_enabled || false,
    purchase_url: domain.purchase_url || '',
    default_locale: domain.default_locale || '',
    seo_title: domain.seo_title || '',
    seo_description: domain.seo_description || '',
    seo_keywords: domain.seo_keywords || '',
    login_redirect: domain.login_redirect || ''
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
      brand_color: editForm.value.brand_color,
      custom_css: editForm.value.custom_css,
      doc_url: editForm.value.doc_url,
      home_content: editForm.value.home_content,
      home_template: editForm.value.home_template,
      purchase_enabled: editForm.value.purchase_enabled,
      purchase_url: editForm.value.purchase_url,
      default_locale: editForm.value.default_locale,
      seo_title: editForm.value.seo_title,
      seo_description: editForm.value.seo_description,
      seo_keywords: editForm.value.seo_keywords,
      login_redirect: editForm.value.login_redirect
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
})
</script>
