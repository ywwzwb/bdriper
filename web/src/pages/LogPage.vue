<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
      <div>
        <h1 style="font-size:28px;font-weight:700;">日志</h1>
        <p style="color:#8A8F98;margin-top:4px;font-size:14px;">系统日志与转码输出</p>
      </div>
      <div style="display:flex;gap:8px;">
        <button class="btn-ghost" @click="clearLogs">清空日志</button>
        <a :href="api.logs.downloadUrl()" class="btn-ghost" style="text-decoration:none;display:inline-flex;align-items:center;gap:4px;">
          <PhDownload size="14" /> 下载
        </a>
      </div>
    </div>

    <div style="display:flex;gap:8px;margin-bottom:16px;flex-wrap:wrap;">
      <button v-for="f in levelFilters" :key="f.key" class="pill" :class="{ active: activeLevel === f.key }" @click="activeLevel = f.key">
        {{ f.label }}
      </button>
    </div>

    <div ref="logContainer" class="glass" style="padding:20px;font-family:'JetBrains Mono','Fira Code',monospace;max-height:70vh;overflow-y:auto;">
      <div v-if="filteredLogs.length">
        <div v-for="(log, i) in filteredLogs" :key="i" style="display:flex;gap:12px;padding:4px 0;font-size:13px;line-height:1.6;">
          <span style="color:#475569;flex-shrink:0;min-width:80px;">{{ log.time }}</span>
          <span :style="{color: levelColor(log.level), flexShrink: '0', minWidth: '44px', fontWeight: 500}">{{ log.level.toUpperCase() }}</span>
          <span style="color:#8A8F98;flex-shrink:0;">{{ log.source }}</span>
          <span style="color:#C4C7CE;word-break:break-all;">{{ log.message || log.msg }}</span>
        </div>
      </div>
      <div v-else style="text-align:center;padding:40px;color:#8A8F98;">暂无日志</div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { PhDownload } from '@phosphor-icons/vue'
import { api } from '@/api'
import { on as wsOn, connect } from '@/ws'

interface LogEntry { time: string; level: string; source: string; message: string; msg?: string }

const logLines = ref<LogEntry[]>([])
const logContainer = ref<HTMLElement>()
const autoScroll = ref(true)

const activeLevel = ref('all')
const levelFilters = [
  { key: 'all', label: '全部' },
  { key: 'info', label: 'INFO' },
  { key: 'warn', label: 'WARN' },
  { key: 'error', label: 'ERROR' },
  { key: 'debug', label: 'DEBUG' },
]

const filteredLogs = computed(() =>
  activeLevel.value === 'all' ? logLines.value : logLines.value.filter(l => l.level === activeLevel.value)
)

function levelColor(l: string) {
  const m: Record<string,string> = { info: '#5E6AD2', warn: '#EAB308', error: '#EF4444', debug: '#475569' }
  return m[l] || '#8A8F98'
}

function clearLogs() { logLines.value = [] }

onMounted(async () => {
  connect()
  try {
    const data = await api.logs.get(500)
    if (data?.lines) {
      const lines = Array.isArray(data.lines) ? data.lines : data.lines.split('\n')
      logLines.value = lines.map((line: string) => {
        const match = line.match(/^(\S+)\s+\[(\w+)\]\s+\[(\w+)\]\s+(.+)$/)
        if (match) {
          return { time: match[1], level: match[2].toLowerCase(), source: match[3], message: match[4] }
        }
        return { time: '', level: 'info', source: '', message: line }
      })
    }
  } catch {}

  wsOn('log', (entry: any) => {
    logLines.value.push(entry)
    if (logLines.value.length > 5000) logLines.value = logLines.value.slice(-3000)
    if (autoScroll.value && logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
})

watch(logContainer, (el) => {
  if (el) {
    el.addEventListener('scroll', () => {
      autoScroll.value = el.scrollTop + el.clientHeight >= el.scrollHeight - 20
    })
  }
})
</script>
