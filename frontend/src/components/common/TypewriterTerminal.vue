<template>
  <div class="font-mono">
    <div class="sr-only" aria-live="polite">{{ fullText }}</div>

    <div aria-hidden="true" class="space-y-1">
      <div v-for="(line, index) in lines" :key="index" class="flex items-start gap-2">
        <span
          v-if="line.kind === 'prompt'"
          class="select-none text-emerald-400"
          >{{ line.prompt ?? '$' }}</span
        >
        <span
          class="min-w-0 flex-1 whitespace-pre-wrap break-all"
          :class="lineClass(line.kind)"
          >{{ renderedLine(index) }}<span v-if="shouldShowCursor(index)" class="tw-cursor"> </span
        ></span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'

export type TerminalLineKind = 'prompt' | 'comment' | 'success' | 'output'

export interface TerminalLine {
  kind?: TerminalLineKind
  prompt?: string
  text: string
}

const props = withDefaults(defineProps<{
  lines: TerminalLine[]
  startDelayMs?: number
  charDelayMs?: number
  lineDelayMs?: number
  loop?: boolean
  loopDelayMs?: number
}>(), {
  startDelayMs: 350,
  charDelayMs: 22,
  lineDelayMs: 220,
  loop: false,
  loopDelayMs: 1500
})

const prefersReducedMotion = ref(false)
const typed = ref<string[]>([])
const currentLineIndex = ref(0)
const currentCharIndex = ref(0)
const isTyping = ref(false)

let timer: number | null = null

const fullText = computed(() => props.lines.map((l) => l.text).join('\n'))

function clearTimer() {
  if (timer !== null) {
    window.clearTimeout(timer)
    timer = null
  }
}

function reset() {
  typed.value = props.lines.map(() => '')
  currentLineIndex.value = 0
  currentCharIndex.value = 0
  isTyping.value = false
}

function finishImmediately() {
  typed.value = props.lines.map((l) => l.text)
  currentLineIndex.value = Math.max(0, props.lines.length - 1)
  currentCharIndex.value = props.lines[props.lines.length - 1]?.text.length ?? 0
  isTyping.value = false
}

function scheduleNext(delayMs: number) {
  clearTimer()
  timer = window.setTimeout(step, delayMs)
}

function scheduleLoopRestart() {
  clearTimer()
  timer = window.setTimeout(() => start(), props.loopDelayMs)
}

function step() {
  if (prefersReducedMotion.value) {
    finishImmediately()
    return
  }

  const line = props.lines[currentLineIndex.value]
  if (!line) {
    isTyping.value = false
    if (props.loop) scheduleLoopRestart()
    return
  }

  isTyping.value = true

  const text = line.text
  const nextCharIndex = currentCharIndex.value + 1

  if (nextCharIndex <= text.length) {
    typed.value[currentLineIndex.value] = text.slice(0, nextCharIndex)
    currentCharIndex.value = nextCharIndex
    scheduleNext(props.charDelayMs)
    return
  }

  // Line done
  currentLineIndex.value += 1
  currentCharIndex.value = 0

  if (currentLineIndex.value >= props.lines.length) {
    isTyping.value = false
    if (props.loop) scheduleLoopRestart()
    return
  }

  scheduleNext(props.lineDelayMs)
}

function start() {
  reset()
  if (prefersReducedMotion.value) {
    finishImmediately()
    return
  }
  scheduleNext(props.startDelayMs)
}

function renderedLine(index: number) {
  return typed.value[index] ?? ''
}

function shouldShowCursor(index: number) {
  if (!props.lines.length) return false
  if (prefersReducedMotion.value) return false

  const typingLineIndex = Math.min(currentLineIndex.value, props.lines.length - 1)

  if (isTyping.value) return index === typingLineIndex
  return index === props.lines.length - 1
}

function lineClass(kind?: TerminalLineKind) {
  switch (kind) {
    case 'comment':
      return 'text-slate-400'
    case 'success':
      return 'text-emerald-300'
    case 'output':
      return 'text-gray-200'
    case 'prompt':
    default:
      return 'text-sky-200'
  }
}

onMounted(() => {
  prefersReducedMotion.value = window.matchMedia?.('(prefers-reduced-motion: reduce)')?.matches ?? false
  start()
})

onBeforeUnmount(() => {
  clearTimer()
})

watch(
  () => props.lines,
  () => start(),
  { deep: true }
)
</script>

<style scoped>
.tw-cursor {
  display: inline-block;
  width: 0.55rem;
  height: 1rem;
  margin-left: 0.15rem;
  vertical-align: -0.1rem;
  background: rgba(209, 117, 86, 0.95);
  animation: tw-blink 1s step-end infinite;
}

@keyframes tw-blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}
</style>
