<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl">
      <!-- Header -->
      <div class="mb-8 text-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Claude Code 接入指南</h1>
        <p class="mt-3 text-gray-600 dark:text-gray-400">仅需三步，让官方工具连接加速网络</p>
      </div>

      <!-- Warning Box -->
      <div class="mb-8 rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-900/20">
        <div class="flex items-start gap-3">
          <svg class="h-6 w-6 flex-shrink-0 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div>
            <h3 class="font-semibold text-amber-800 dark:text-amber-200">重要提示：不支持 Web 和 Cursor</h3>
            <p class="mt-1 text-sm text-amber-700 dark:text-amber-300">
              本接口为 Anthropic 原生协议，专为 claude-code 命令行工具设计。<br>
              不支持网页版 (claude.ai) 登录，也不支持 Cursor 等需要 OpenAI 格式的软件。
            </p>
          </div>
        </div>
      </div>

      <!-- Step 1: Install Claude CLI -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">1</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">安装 Claude CLI</h2>
        </div>

        <p class="mb-4 text-gray-600 dark:text-gray-400">根据您的操作系统选择对应的安装方式：</p>

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
          <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500"># 下载并安装 Claude CLI</span>
<span class="text-green-400">curl</span> <span class="text-white">-fsSL</span> <span class="text-yellow-300">https://claude.ai/install.sh</span> <span class="text-white">|</span> <span class="text-green-400">bash</span></code></pre>
          <button
            @click="copyCode('install')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'install' ? '已复制' : '复制' }}
          </button>
        </div>

        <!-- Install Code Block - Windows PowerShell -->
        <div v-else-if="activeInstallTab === 'powershell'" class="relative rounded-lg bg-gray-900 p-4">
          <div class="mb-2 text-xs text-gray-500">powershell</div>
          <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500"># 下载并安装 Claude CLI</span>
<span class="text-green-400">irm</span> <span class="text-yellow-300">https://claude.ai/install.ps1</span> <span class="text-white">|</span> <span class="text-green-400">iex</span></code></pre>
          <button
            @click="copyCode('install')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'install' ? '已复制' : '复制' }}
          </button>
        </div>

        <!-- Install Code Block - Windows CMD -->
        <div v-else class="relative rounded-lg bg-gray-900 p-4">
          <div class="mb-2 text-xs text-gray-500">cmd</div>
          <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">:: 使用 npm 安装 Claude CLI</span>
<span class="text-green-400">npm</span> <span class="text-white">install -g</span> <span class="text-yellow-300">@anthropic-ai/claude-code</span></code></pre>
          <button
            @click="copyCode('install')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'install' ? '已复制' : '复制' }}
          </button>
        </div>

        <!-- Install Tip -->
        <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-yellow-500">💡</span>
          <span v-if="activeInstallTab === 'mac'">
            安装完成后可能需要重启终端或运行
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">source ~/.bashrc</code>
          </span>
          <span v-else-if="activeInstallTab === 'powershell'">
            安装完成后可能需要重启 PowerShell 终端
          </span>
          <span v-else>
            请确保已安装 Node.js，安装完成后重启 CMD 窗口
          </span>
        </p>

        <!-- Verify Installation -->
        <div class="mt-4 rounded-lg border border-green-200 bg-green-50 p-4 dark:border-green-800 dark:bg-green-900/20">
          <div class="mb-2 text-sm font-medium text-green-700 dark:text-green-300">验证安装</div>
          <div class="rounded bg-gray-900 p-3">
            <code class="text-sm text-green-400">claude --version</code>
          </div>
          <p class="mt-2 text-xs text-green-600 dark:text-green-400">如果显示版本号，说明安装成功</p>
        </div>
      </section>

      <!-- Step 2: Create API Key and Configure -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">2</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">创建密钥并配置环境变量</h2>
        </div>

        <ol class="mb-6 space-y-3 text-gray-700 dark:text-gray-300">
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">1.</span>
            <span>登录后进入 <router-link to="/keys" class="text-primary-500 hover:underline">后台 → API 密钥</router-link></span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">2.</span>
            <span>点击右上角「创建密钥」，输入名称后确认</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">3.</span>
            <span>点击新密钥右侧的 <span class="inline-block rounded bg-primary-500 px-2 py-0.5 text-xs text-white">使用密钥</span> 按钮</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">4.</span>
            <span>选择您的操作系统，复制环境变量命令</span>
          </li>
          <li class="flex items-start gap-2">
            <span class="font-medium text-gray-500">5.</span>
            <span>在终端中粘贴并执行命令</span>
          </li>
        </ol>

        <!-- Example Code Section -->
        <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-700/50">
          <p class="mb-3 text-sm text-gray-600 dark:text-gray-400">「使用密钥」弹窗示例：</p>

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
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500"># 复制以下命令到终端执行</span>
<span class="text-blue-400">export</span> <span class="text-white">ANTHROPIC_BASE_URL=</span><span class="text-yellow-300">"{{ baseUrl }}"</span>
<span class="text-blue-400">export</span> <span class="text-white">ANTHROPIC_API_KEY=</span><span class="text-yellow-300">"sk-您的密钥"</span></code></pre>
            <button
              @click="copyCode('config')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'config' ? '已复制' : '复制' }}
            </button>
          </div>

          <!-- Config Code Block - Windows PowerShell -->
          <div v-else-if="activeConfigTab === 'powershell'" class="relative rounded-lg bg-gray-900 p-4">
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500"># 复制以下命令到 PowerShell 执行</span>
<span class="text-blue-400">$env:</span><span class="text-white">ANTHROPIC_BASE_URL</span> <span class="text-white">=</span> <span class="text-yellow-300">"{{ baseUrl }}"</span>
<span class="text-blue-400">$env:</span><span class="text-white">ANTHROPIC_API_KEY</span> <span class="text-white">=</span> <span class="text-yellow-300">"sk-您的密钥"</span></code></pre>
            <button
              @click="copyCode('config')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'config' ? '已复制' : '复制' }}
            </button>
          </div>

          <!-- Config Code Block - Windows CMD -->
          <div v-else class="relative rounded-lg bg-gray-900 p-4">
            <pre class="overflow-x-auto text-sm leading-relaxed"><code><span class="text-gray-500">:: 复制以下命令到 CMD 执行</span>
<span class="text-blue-400">set</span> <span class="text-white">ANTHROPIC_BASE_URL=</span><span class="text-yellow-300">{{ baseUrl }}</span>
<span class="text-blue-400">set</span> <span class="text-white">ANTHROPIC_API_KEY=</span><span class="text-yellow-300">sk-您的密钥</span></code></pre>
            <button
              @click="copyCode('config')"
              class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
            >
              {{ copied === 'config' ? '已复制' : '复制' }}
            </button>
          </div>
        </div>

        <!-- Config Tip -->
        <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-yellow-500">💡</span>
          <span v-if="activeConfigTab === 'mac'">
            建议将命令添加到
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">~/.zshrc</code>
            或
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">~/.bashrc</code>
            实现永久配置
          </span>
          <span v-else-if="activeConfigTab === 'powershell'">
            建议将命令添加到 PowerShell 配置文件
            <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs dark:bg-dark-700">$PROFILE</code>
            实现永久配置
          </span>
          <span v-else>
            建议通过「系统属性 → 环境变量」设置永久环境变量
          </span>
        </p>
      </section>

      <!-- Step 3: Start Claude -->
      <section class="mb-8 rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <div class="mb-4 flex items-center gap-3">
          <span class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-500 text-sm font-bold text-white">3</span>
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">启动 Claude</h2>
        </div>

        <p class="mb-4 text-gray-600 dark:text-gray-400">配置完成后，在终端中运行以下命令启动 Claude：</p>

        <div class="relative rounded-lg border border-gray-200 bg-gray-900 p-4 dark:border-dark-600">
          <code class="text-lg text-green-400">claude</code>
          <button
            @click="copyCode('start')"
            class="absolute right-2 top-2 rounded bg-gray-700 px-2 py-1 text-xs text-gray-300 hover:bg-gray-600"
          >
            {{ copied === 'start' ? '已复制' : '复制' }}
          </button>
        </div>

        <p class="mt-3 flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span class="text-yellow-500">💡</span>
          首次运行可能需要几秒钟初始化，之后即可开始对话
        </p>
      </section>

      <!-- Tips Box -->
      <section class="rounded-xl border border-blue-200 bg-blue-50 p-6 dark:border-blue-800 dark:bg-blue-900/20">
        <h3 class="mb-4 flex items-center gap-2 font-semibold text-blue-800 dark:text-blue-200">
          <span class="text-yellow-500">💡</span>
          提示
        </h3>
        <ul class="space-y-2 text-sm text-blue-700 dark:text-blue-300">
          <li class="flex items-start gap-2">
            <span>•</span>
            <span>环境变量配置后立即生效，无需重启终端</span>
          </li>
          <li class="flex items-start gap-2">
            <span>•</span>
            <span>建议配置永久环境变量，避免每次都要重新设置</span>
          </li>
          <li class="flex items-start gap-2">
            <span>•</span>
            <span>遇到问题？查看 <router-link to="/dashboard" class="font-medium underline">系统状态</router-link> 或联系技术支持</span>
          </li>
        </ul>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'

const copied = ref<string | null>(null)
const activeInstallTab = ref('mac')
const activeConfigTab = ref('mac')

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

const baseUrl = computed(() => {
  return window.location.origin
})

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
  switch (activeConfigTab.value) {
    case 'mac':
      return `export ANTHROPIC_BASE_URL="${baseUrl.value}"
export ANTHROPIC_API_KEY="sk-您的密钥"`
    case 'powershell':
      return `$env:ANTHROPIC_BASE_URL = "${baseUrl.value}"
$env:ANTHROPIC_API_KEY = "sk-您的密钥"`
    case 'cmd':
      return `set ANTHROPIC_BASE_URL=${baseUrl.value}
set ANTHROPIC_API_KEY=sk-您的密钥`
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
  } else if (type === 'start') {
    code = 'claude'
  }

  navigator.clipboard.writeText(code)
  copied.value = type
  setTimeout(() => {
    copied.value = null
  }, 2000)
}
</script>
