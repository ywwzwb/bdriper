<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-semibold">系统日志</h1>
      <div class="flex gap-3 items-center">
        <div class="flex gap-2">
          <label v-for="l in levels" :key="l" class="flex items-center gap-1.5 text-sm cursor-pointer select-none">
            <input type="checkbox" :checked="activeLevels.has(l)" @change="toggleLevel(l)" class="w-3.5 h-3.5 rounded border-border bg-muted accent-accent" />
            <span :class="logColor(l)">{{ levelLabel(l) }}</span>
          </label>
        </div>
        <a :href="api.logs.downloadUrl()" class="px-3 py-1.5 text-sm border border-border rounded-lg hover:bg-muted transition">下载日志</a>
      </div>
    </div>

    <div class="bg-card border border-border rounded-lg overflow-hidden">
      <div v-if="truncated" class="px-4 py-2 bg-muted/50 border-b border-border text-sm text-muted flex items-center justify-between">
        <span>... 已隐藏 {{ hiddenLines }} 行旧日志 ...</span>
        <button @click="showAll" class="text-accent hover:underline">显示全部</button>
      </div>
      <div ref="logContainer" class="p-4 font-mono text-xs leading-relaxed max-h-[calc(100vh-220px)] overflow-auto" @scroll="onScroll">
        <div v-for="(line, i) in displayedLines" :key="i" class="whitespace-pre-wrap" :class="logColor(getLevel(line))">{{ line }}</div>
        <div v-if="filteredLines.length === 0" class="text-muted">暂无日志</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { api } from '@/api'
import { connectLogs } from '@/ws'

const MAX_LINES = 5000
const levels = ['debug', 'info', 'warn', 'error']
const levelLabels: Record<string, string> = {
  debug: '调试',
  info: '信息',
  warn: '警告',
  error: '错误',
}
const activeLevels = ref(new Set(['info', 'warn', 'error']))
const allLines = ref<string[]>([])
const truncated = ref(false)
const hiddenLines = ref(0)
const autoScroll = ref(true)
const logContainer = ref<HTMLElement | null>(null)
let logCleanup: (() => void) | null = null

const filteredLines = computed(() => {
  return allLines.value.filter(line => {
    const lvl = getLevel(line)
    return activeLevels.value.has(lvl)
  })
})

const displayedLines = computed(() => {
  if (truncated.value && filteredLines.value.length > MAX_LINES) {
    return filteredLines.value.slice(-MAX_LINES)
  }
  return filteredLines.value
})

function getLevel(line: string): string {
  const match = line.match(/\[(debug|info|warn|error)\]/i)
  return match ? match[1].toLowerCase() : 'debug'
}

function levelLabel(level: string): string {
  return levelLabels[level] || level.toUpperCase()
}

function logColor(level: string) {
  const map: Record<string, string> = { debug: 'text-slate-400', info: 'text-fg', warn: 'text-yellow-400', error: 'text-destructive' }
  return map[level] || 'text-fg'
}

function toggleLevel(l: string) {
  const next = new Set(activeLevels.value)
  if (next.has(l)) next.delete(l)
  else next.add(l)
  activeLevels.value = next
}

function showAll() {
  truncated.value = false
}

function onScroll() {
  if (!logContainer.value) return
  const el = logContainer.value
  autoScroll.value = el.scrollHeight - el.scrollTop - el.clientHeight < 40
}

async function scrollToBottom() {
  await nextTick()
  if (logContainer.value && autoScroll.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

onMounted(async () => {
  try {
    const text = await api.logs.get(1000)
    allLines.value = text.split('\n').filter(Boolean)
    if (allLines.value.length > MAX_LINES) {
      truncated.value = true
      hiddenLines.value = allLines.value.length - MAX_LINES
    }
    await scrollToBottom()
  } catch { /* ignore initial load error */ }

  logCleanup = connectLogs((line: string) => {
    allLines.value.push(line)
    if (allLines.value.length > MAX_LINES && !truncated.value) {
      truncated.value = true
      hiddenLines.value = allLines.value.length - MAX_LINES
    } else if (truncated.value) {
      hiddenLines.value = allLines.value.length - MAX_LINES
    }
    scrollToBottom()
  })
})

onUnmounted(() => {
  if (logCleanup) logCleanup()
})
</script>
