<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- Current Balance Card -->
      <div class="card overflow-hidden">
        <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
          <div class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm">
            <Icon name="creditCard" size="xl" class="text-white" />
          </div>
          <p class="text-sm font-medium text-primary-100">{{ t('recharge.currentBalance') }}</p>
          <p class="mt-2 text-4xl font-bold text-white">${{ user?.balance?.toFixed(2) || '0.00' }}</p>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </div>

      <!-- No method available -->
      <div v-else-if="availableMethods.length === 0" class="card p-12 text-center">
        <Icon name="exclamationCircle" size="xl" class="mx-auto text-gray-400 dark:text-dark-500" />
        <h3 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('recharge.disabled') }}</h3>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">{{ t('recharge.disabledDesc') }}</p>
      </div>

      <!-- Recharge -->
      <template v-else>
        <div class="card">
          <div class="p-6">
            <div class="mb-5 flex items-center justify-between">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('recharge.formTitle') }}</h2>
              <router-link
                to="/recharge-orders"
                class="inline-flex items-center gap-1 text-sm font-medium text-primary-500 hover:text-primary-600 dark:text-primary-400"
              >
                {{ t('recharge.viewAllOrders') }}
                <Icon name="chevronRight" size="sm" />
              </router-link>
            </div>

            <!-- Payment method switcher (gated + locale-defaulted) -->
            <div v-if="availableMethods.length > 1" class="mb-5">
              <label class="input-label">{{ t('recharge.selectMethod') }}</label>
              <div class="mt-2 grid gap-3" :class="availableMethods.length === 3 ? 'grid-cols-3' : 'grid-cols-2'">
                <button
                  v-for="m in availableMethods"
                  :key="m"
                  type="button"
                  :class="[
                    'flex flex-col items-center justify-center gap-2 rounded-xl border-2 px-3 py-4 text-sm font-medium transition-all',
                    activeMethod === m
                      ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-300'
                      : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-dark-600 dark:text-dark-300 dark:hover:border-dark-500'
                  ]"
                  @click="activeMethod = m"
                >
                  <span :class="methodIconClass(m)" v-html="methodIcon(m)" />
                  <span>{{ methodLabel(m) }}</span>
                </button>
              </div>
            </div>

            <!-- Active panel -->
            <AlipayRechargePanel v-if="activeMethod === 'alipay' && alimpayConfig" :config="alimpayConfig" @paid="onPaid" />
            <WechatRechargePanel v-else-if="activeMethod === 'wechat' && rechargeConfig" :config="rechargeConfig" @paid="onPaid" />
            <UsdtRechargePanel v-else-if="activeMethod === 'usdt' && usdtConfig" :config="usdtConfig" @paid="onPaid" />
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
/**
 * 统一充值页：把原本三个独立菜单（余额充值/支付宝个人免签/USDT）合并为一个页面。
 * - 支付宝 → AliMPay 个人免签
 * - 微信   → 易支付 EPAY（pay_type=wxpay）
 * - USDT   → 自建多链
 * 标签按各通道 enabled 动态显隐；默认选中：中文优先支付宝，其他语言优先 USDT。
 */
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { alimpayAPI, type AliMPayConfig } from '@/api/alimpayOrder'
import { rechargeAPI, type RechargeConfig } from '@/api/recharge'
import { usdtAPI, type UsdtConfig } from '@/api/usdtOrder'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import AlipayRechargePanel from './recharge/AlipayRechargePanel.vue'
import WechatRechargePanel from './recharge/WechatRechargePanel.vue'
import UsdtRechargePanel from './recharge/UsdtRechargePanel.vue'

type Method = 'alipay' | 'wechat' | 'usdt'

const { t, locale } = useI18n()
const authStore = useAuthStore()

const user = computed(() => authStore.user)

const loading = ref(true)
const alimpayConfig = ref<AliMPayConfig | null>(null)
const rechargeConfig = ref<RechargeConfig | null>(null)
const usdtConfig = ref<UsdtConfig | null>(null)
const activeMethod = ref<Method>('alipay')

const availableMethods = computed<Method[]>(() => {
  const methods: Method[] = []
  if (alimpayConfig.value?.enabled) methods.push('alipay')
  // 微信仅来自易支付 EPAY；需 EPAY 启用且其支付方式含 wxpay
  if (rechargeConfig.value?.enabled && rechargeConfig.value.pay_types?.includes('wxpay')) methods.push('wechat')
  if (usdtConfig.value?.enabled) methods.push('usdt')
  return methods
})

const ALIPAY_ICON =
  '<svg class="h-8 w-8" viewBox="0 0 1024 1024" fill="currentColor"><path d="M789.024 647.104c-49.088-20.928-103.104-43.392-106.944-45.184 22.336-40.192 40.128-84.928 51.776-133.12h-150.72V416.512h181.12v-30.784h-181.12V301.44h-78.592s-5.568 0-5.568 8.128v76.16H320v30.784h178.976V468.8H352.96v30.784h268.352c-9.792 35.136-23.424 67.84-40.192 96.96a1603.84 1603.84 0 0 0-196.608-56.768C266.816 514.816 192 571.456 192 647.104c0 75.648 71.616 128.832 199.232 128.832 87.36 0 172.288-35.648 238.976-98.432 43.392 23.424 76.288 42.688 76.288 42.688l82.528-73.088zM371.968 736.576c-112.448 0-140.032-45.12-140.032-80.256 0-47.552 52.416-90.24 127.616-90.24 55.104 0 117.888 14.272 178.816 39.616-52.352 77.376-114.496 130.88-166.4 130.88z"/></svg>'
const WECHAT_ICON =
  '<svg class="h-8 w-8" viewBox="0 0 1024 1024" fill="currentColor"><path d="M690.1 377.4c5.9 0 11.8.2 17.6.5-15.8-73.2-88.3-127.8-174.5-127.8-95.5 0-173.3 63.6-173.3 142.7 0 46.4 25.3 84.7 67.2 114.5l-16.8 50.4 58.4-29.2c21.1 4.2 37.9 8.4 58.4 8.4 5.8 0 11.5-.2 17.1-.7-3.6-12.1-5.6-24.8-5.6-37.9 0-67.3 64.2-120.9 151.5-120.9zM563.4 318.8c12.6 0 21.1 8.4 21.1 21.1 0 12.6-8.4 21.1-21.1 21.1-12.6 0-25.2-8.4-25.2-21.1 0-12.6 12.6-21.1 25.2-21.1zm-116.8 42.2c-12.6 0-25.2-8.4-25.2-21.1 0-12.6 12.6-21.1 25.2-21.1 12.6 0 21.1 8.4 21.1 21.1 0 12.6-8.5 21.1-21.1 21.1zm384 134.7c0-67.3-67.2-122-147.6-122-84.7 0-147.6 54.7-147.6 122s62.9 122 147.6 122c16.8 0 37.9-4.2 54.7-8.4l46.4 25.2-12.6-42.2c33.7-25.2 59.1-58.9 59.1-96.6zm-196.8-21.1c-8.4 0-16.8-8.4-16.8-16.8 0-8.4 8.4-16.8 16.8-16.8 12.6 0 21.1 8.4 21.1 16.8 0 8.4-8.5 16.8-21.1 16.8zm100.8 0c-8.4 0-16.8-8.4-16.8-16.8 0-8.4 8.4-16.8 16.8-16.8 12.6 0 21.1 8.4 21.1 16.8 0 8.4-8.5 16.8-21.1 16.8z"/></svg>'
const USDT_ICON =
  '<svg class="h-8 w-8" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2 1.5 8v8L12 22l10.5-6V8L12 2Zm1.5 9.6v1.2c2.4.1 4.2.5 4.2 1s-1.8.9-4.2 1v3h-3v-3c-2.4-.1-4.2-.5-4.2-1s1.8-.9 4.2-1v-1.2c-2 .1-3.6.4-3.6.8v-2.4h6.6v2.4c0-.4-1.6-.7-3.6-.8h3.6Zm-1.5 2.1c2.9 0 5.4-.4 5.4-1s-2.5-1-5.4-1-5.4.4-5.4 1 2.5 1 5.4 1Z"/></svg>'

const methodIcon = (m: Method) => (m === 'alipay' ? ALIPAY_ICON : m === 'wechat' ? WECHAT_ICON : USDT_ICON)
const methodIconClass = (m: Method) => (m === 'alipay' ? 'text-blue-500' : m === 'wechat' ? 'text-green-500' : 'text-teal-500')
const methodLabel = (m: Method) =>
  m === 'alipay' ? t('recharge.methodAlipay') : m === 'wechat' ? t('recharge.methodWechat') : t('recharge.methodUsdt')

// 中文优先支付宝，其他语言优先 USDT
const pickDefaultMethod = (available: Method[]): Method => {
  const isZh = String(locale.value).toLowerCase().startsWith('zh')
  const order: Method[] = isZh ? ['alipay', 'wechat', 'usdt'] : ['usdt', 'alipay', 'wechat']
  return order.find((m) => available.includes(m)) ?? available[0]
}

const onPaid = () => {
  // 余额已在面板内通过 authStore.refreshUser() 刷新
}

onMounted(async () => {
  loading.value = true
  // 三个通道配置并行拉取，单个失败不影响其它通道
  const [alimpayRes, rechargeRes, usdtRes] = await Promise.allSettled([
    alimpayAPI.getConfig(),
    rechargeAPI.getConfig(),
    usdtAPI.getConfig()
  ])
  if (alimpayRes.status === 'fulfilled') alimpayConfig.value = alimpayRes.value
  if (rechargeRes.status === 'fulfilled') rechargeConfig.value = rechargeRes.value
  if (usdtRes.status === 'fulfilled') usdtConfig.value = usdtRes.value
  loading.value = false

  if (availableMethods.value.length > 0) {
    activeMethod.value = pickDefaultMethod(availableMethods.value)
  }
})
</script>
