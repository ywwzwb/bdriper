<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-3xl font-bold text-fg">系统日志</h1>
        <p class="text-muted mt-1">实时服务器日志输出</p>
      </div>
      <div class="flex gap-3 items-center">
        <div class="flex gap-1 bg-muted/50 rounded-xl p-1">
          <button v-for="level in levels" :key="level" @click="toggleLevel(level)" class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 flex items-center gap-1.5" :class="activeLevels.has(level) ? 'bg-card text-fg shadow-sm' : 'text-muted hover:text-fg'">
            <span class="w-1.5 h-1.5 rounded-full" :class="levelDot(level)" />
            {{ levelLabel(level) }}
          </button>
        </div>
        <a :href="api.logs.downloadUrl()" class="px-4 py-2 text-sm border border-border/50 rounded-lg hover:bg-muted transition-all duration-200 flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4"/><polyline points="7,10 12,15 17,10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          下载日志
        </a>
      </div>
    </div>

    <div class="bg-card border border-border/50 rounded-xl overflow-hidden">
      <div v-if="truncated" class="px-4 py-2.5 bg-muted/30 border-b border-border/50 text-sm text-muted flex items-center justify-between">
        <span>... 已隐藏 {{ hiddenLines }} 行旧日志 ...</span>
        <button @click="showAll" class="text-accent hover:underline text-sm">显示全部</button>
      </div>
      <div ref="logContainer" class="p-4 font-mono text-xs leading-relaxed max-h-[calc(100vh-240px)] overflow-auto bg-[#0a0e17]" @scroll="onScroll">
        <div v-for="(line, i) in displayedLines" :key="i" class="whitespace-pre-wrap py-0.5" :class="logLineColor(getLevel(line))">{{ line }}</div>
        <div v-if="filteredLines.length === 0" class="text-muted py-4 text-center">暂无匹配日志</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { api } from '@/api'
import { connectLogs } from '@/ws'

const MAX_LINES = 5000
const levels = ['debug', 'info', 'warn', 'error']
const levelLabels: Record<string, string> = { debug: '调试', info: '信息', warn: '警告', error: '错误' }
const activeLevels = ref(new Set(['info', 'warn', 'error']))
const allLines = ref<string[]>([])
const truncated = ref(false)
const hiddenLines = ref(0)
const autoScroll = ref(true)
const logContainer = ref<HTMLElement | null>(null)
let logCleanup: (() => void) | null = null

const filteredLines = computed(() => allLines.value.filter(line => {
  const lvl = getLevel(line)
  return activeLevels.value.has(lvl)
}))

const displayedLines = computed(() => {
  if (truncated.value && filteredLines.value.length > MAX_LINES) return filteredLines.value.slice(-MAX_LINES)
  return filteredLines.value
})

function getLevel(line: string): string {
  const match = line.match(/\[(debug|info|warn|error)\]/i)
  return match ? match[1].toLowerCase() : 'debug'
}

function levelLabel(level: string): string { return levelLabels[level] || level.toUpperCase() }

function levelDot(level: string) {
  const map: Record<string, string> = { debug: 'bg-slate-400', info: 'bg-fg', warn: 'bg-yellow-400', error: 'bg-destructive' }
  return map[level] || 'bg-fg'
}

function logLineColor(level: string) {
  const map: Record<string, string> = { debug: 'text-slate-500', info: 'text-slate-200', warn: 'text-yellow-400', error: 'text-destructive font-medium' }
  return map[level] || 'text-fg'
}

function toggleLevel(lvl: string) {
  const next = new Set(activeLevels.value)
  if (next.has(lvl)) next.delete(lvl)
  else next.add(lvl)
  activeLevels.value = next
}

function showAll() { truncated.value = false }

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

onUnmounted(() => { if (logCleanup) logCleanup() })
</script>
