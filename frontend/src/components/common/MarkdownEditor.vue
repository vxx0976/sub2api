<template>
  <div class="markdown-editor overflow-hidden rounded-lg border border-gray-300 dark:border-dark-600" :class="{ 'ring-2 ring-primary-500': focused }">
    <!-- Toolbar -->
    <div class="flex items-center justify-between border-b border-gray-200 bg-gray-50 px-3 py-1.5 dark:border-dark-600 dark:bg-dark-800">
      <!-- Format buttons -->
      <div class="flex items-center gap-0.5">
        <button type="button" @click="insertFormat('bold')" class="toolbar-btn" :title="t('markdownEditor.bold')">
          <span class="text-xs font-bold">B</span>
        </button>
        <button type="button" @click="insertFormat('italic')" class="toolbar-btn" :title="t('markdownEditor.italic')">
          <span class="text-xs italic">I</span>
        </button>
        <button type="button" @click="insertFormat('strikethrough')" class="toolbar-btn" :title="t('markdownEditor.strikethrough')">
          <span class="text-xs line-through">S</span>
        </button>
        <span class="mx-1 h-4 w-px bg-gray-300 dark:bg-dark-600"></span>
        <button type="button" @click="insertFormat('heading')" class="toolbar-btn" :title="t('markdownEditor.heading')">
          <span class="text-xs font-bold">H</span>
        </button>
        <button type="button" @click="insertFormat('quote')" class="toolbar-btn" :title="t('markdownEditor.quote')">
          <span class="text-xs">"</span>
        </button>
        <button type="button" @click="insertFormat('code')" class="toolbar-btn" :title="t('markdownEditor.code')">
          <span class="text-xs font-mono">&lt;/&gt;</span>
        </button>
        <span class="mx-1 h-4 w-px bg-gray-300 dark:bg-dark-600"></span>
        <button type="button" @click="insertFormat('ul')" class="toolbar-btn !w-auto px-1.5" :title="t('markdownEditor.unorderedList')">
          <span class="text-xs">- list</span>
        </button>
        <button type="button" @click="insertFormat('ol')" class="toolbar-btn !w-auto px-1.5" :title="t('markdownEditor.orderedList')">
          <span class="text-xs">1. list</span>
        </button>
        <button type="button" @click="insertFormat('link')" class="toolbar-btn" :title="t('markdownEditor.link')">
          <Icon name="link" size="sm" />
        </button>
      </div>

      <!-- Tab switcher -->
      <div class="flex items-center rounded-md bg-gray-200 p-0.5 dark:bg-dark-700">
        <button
          type="button"
          @click="activeTab = 'edit'"
          class="rounded px-2.5 py-1 text-xs font-medium transition-colors"
          :class="activeTab === 'edit' ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-600 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
        >
          {{ t('markdownEditor.edit') }}
        </button>
        <button
          type="button"
          @click="activeTab = 'preview'"
          class="rounded px-2.5 py-1 text-xs font-medium transition-colors"
          :class="activeTab === 'preview' ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-600 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
        >
          {{ t('markdownEditor.preview') }}
        </button>
      </div>
    </div>

    <!-- Edit mode -->
    <textarea
      v-show="activeTab === 'edit'"
      ref="textareaRef"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
      @focus="focused = true"
      @blur="focused = false"
      @invalid="activeTab = 'edit'"
      :rows="rows"
      :placeholder="placeholder"
      :required="required"
      class="w-full resize-y border-0 bg-white px-4 py-3 font-mono text-sm text-gray-900 placeholder-gray-400 focus:outline-none focus:ring-0 dark:bg-dark-900 dark:text-white dark:placeholder-gray-500"
    ></textarea>

    <!-- Preview mode -->
    <div
      v-show="activeTab === 'preview'"
      class="min-h-[120px] bg-white px-4 py-3 dark:bg-dark-900"
      :style="{ minHeight: rows * 1.5 + 1.5 + 'rem' }"
    >
      <div
        v-if="modelValue"
        class="markdown-body prose prose-sm max-w-none dark:prose-invert"
        v-html="renderedHtml"
      ></div>
      <p v-else class="text-sm text-gray-400 dark:text-gray-500">{{ t('markdownEditor.emptyPreview') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Marked } from 'marked'
import DOMPurify from 'dompurify'
import Icon from '@/components/icons/Icon.vue'

const md = new Marked({ breaks: true, gfm: true })

const props = withDefaults(defineProps<{
  modelValue: string
  rows?: number
  placeholder?: string
  required?: boolean
}>(), {
  rows: 8,
  placeholder: '',
  required: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { t } = useI18n()
const activeTab = ref<'edit' | 'preview'>('edit')
const focused = ref(false)
const textareaRef = ref<HTMLTextAreaElement | null>(null)

const renderedHtml = computed(() => {
  if (!props.modelValue) return ''
  const html = md.parse(props.modelValue) as string
  return DOMPurify.sanitize(html)
})

function insertFormat(type: string) {
  const textarea = textareaRef.value
  if (!textarea) return

  activeTab.value = 'edit'

  // Wait for textarea to be visible
  requestAnimationFrame(() => {
    const start = textarea.selectionStart
    const end = textarea.selectionEnd
    const text = props.modelValue
    const selected = text.slice(start, end)

    let before = ''
    let after = ''
    let insert = ''

    switch (type) {
      case 'bold':
        before = '**'; after = '**'
        insert = selected || t('markdownEditor.boldText')
        break
      case 'italic':
        before = '*'; after = '*'
        insert = selected || t('markdownEditor.italicText')
        break
      case 'strikethrough':
        before = '~~'; after = '~~'
        insert = selected || t('markdownEditor.strikethroughText')
        break
      case 'heading':
        before = lineStart(text, start) ? '## ' : '\n## '
        insert = selected || t('markdownEditor.headingText')
        break
      case 'quote':
        before = lineStart(text, start) ? '> ' : '\n> '
        insert = selected || t('markdownEditor.quoteText')
        break
      case 'code':
        if (selected.includes('\n')) {
          before = '```\n'; after = '\n```'
          insert = selected
        } else {
          before = '`'; after = '`'
          insert = selected || t('markdownEditor.codeText')
        }
        break
      case 'ul':
        before = lineStart(text, start) ? '- ' : '\n- '
        insert = selected || t('markdownEditor.listItem')
        break
      case 'ol':
        before = lineStart(text, start) ? '1. ' : '\n1. '
        insert = selected || t('markdownEditor.listItem')
        break
      case 'link':
        before = '['; after = '](url)'
        insert = selected || t('markdownEditor.linkText')
        break
    }

    const newText = text.slice(0, start) + before + insert + after + text.slice(end)
    emit('update:modelValue', newText)

    // Select the inserted text so user can immediately type over placeholder
    requestAnimationFrame(() => {
      textarea.focus()
      textarea.setSelectionRange(start + before.length, start + before.length + insert.length)
    })
  })
}

function lineStart(text: string, pos: number): boolean {
  return pos === 0 || text[pos - 1] === '\n'
}

</script>

<style scoped>
.toolbar-btn {
  @apply flex h-7 w-7 items-center justify-center rounded text-gray-600 transition-colors hover:bg-gray-200 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-600 dark:hover:text-white;
}
</style>
