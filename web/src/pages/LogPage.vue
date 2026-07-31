<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
      <div>
        <h1 style="font-size:28px;font-weight:700;">日志</h1>
        <p style="color:#8A8F98;margin-top:4px;font-size:14px;">系统日志与转码输出</p>
      </div>
      <div style="display:flex;gap:8px;">
        <button class="btn-ghost" @click="clearLogs">清空日志</button>
        <button class="btn-ghost" @click="downloadLogs">
          <PhDownload size="14" /> 下载
        </button>
      </div>
    </div>

    <div style="display:flex;gap:8px;margin-bottom:16px;flex-wrap:wrap;">
      <button v-for="f in levelFilters" :key="f.key" class="pill" :class="{ active: activeLevel === f.key }" @click="activeLevel = f.key">
        {{ f.label }}
      </button>
    </div>

    <div class="glass" style="padding:20px;font-family:'JetBrains Mono','Fira Code',monospace;max-height:70vh;overflow-y:auto;">
      <div v-if="filteredLogs.length">
        <div v-for="(log, i) in filteredLogs" :key="i" style="display:flex;gap:12px;padding:4px 0;font-size:13px;line-height:1.6;">
          <span style="color:#475569;flex-shrink:0;min-width:80px;">{{ log.time }}</span>
          <span :style="{color: levelColor(log.level), flexShrink: '0', minWidth: '44px', fontWeight: 500}">{{ log.level.toUpperCase() }}</span>
          <span style="color:#8A8F98;flex-shrink:0;">{{ log.source }}</span>
          <span style="color:#C4C7CE;word-break:break-all;">{{ log.message }}</span>
        </div>
      </div>
      <div v-else style="text-align:center;padding:40px;color:#8A8F98;">暂无日志</div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, computed } from 'vue'
import { PhDownload } from '@phosphor-icons/vue'

interface LogEntry { time: string; level: string; source: string; message: string }

const logs = ref<LogEntry[]>([
  { time: '14:32:01', level: 'info', source: 'server', message: 'BDRiper 服务启动成功，监听端口 8080' },
  { time: '14:32:02', level: 'info', source: 'worker', message: '工作池初始化完成，最大并发: 3' },
  { time: '14:32:15', level: 'info', source: 'task', message: '[任务 #1] 四月は君の嘘 EP01 开始转码 → x265 HQ Anime' },
  { time: '14:32:16', level: 'debug', source: 'ffmpeg', message: '[#1] ffmpeg -i 00000.m2ts -c:v libx265 -preset slow -crf 18 ...' },
  { time: '14:33:20', level: 'info', source: 'task', message: '[任务 #2] 化物語 EP02 加入队列等待中' },
  { time: '14:35:44', level: 'warn', source: 'ffmpeg', message: '[#1] Frame rate mismatch detected: 23.976 vs 24000/1001' },
  { time: '14:38:01', level: 'info', source: 'task', message: '[任务 #3] Charlotte EP01 转码完成，耗时 12 分钟' },
  { time: '14:40:12', level: 'error', source: 'ffmpeg', message: '[#4] Kill la Kill EP01 编码失败: Unknown stream type in track 3' },
  { time: '14:40:12', level: 'error', source: 'task', message: '[任务 #4] Kill la Kill EP01 转码失败: 编码器错误' },
  { time: '14:41:05', level: 'info', source: 'task', message: '[任务 #5] Steins;Gate EP01 开始转码 → x264 High Quality' },
  { time: '14:45:30', level: 'info', source: 'task', message: '[任务 #5] Steins;Gate EP01 转码完成，耗时 4.4 分钟' },
  { time: '14:46:00', level: 'info', source: 'task', message: '[任务 #1] 四月は君の嘘 EP01 进度 45%，剩余 12 分钟' },
  { time: '14:48:01', level: 'info', source: 'system', message: '自动保存配置完成' },
  { time: '14:50:00', level: 'debug', source: 'worker', message: '内存使用: 24MB, 协程: 12, GC 暂停: 0.2ms' },
  { time: '14:52:30', level: 'warn', source: 'server', message: '磁盘空间低于 20%，剩余 15.3 GB' },
])

const activeLevel = ref('all')
const levelFilters = [
  { key: 'all', label: '全部' },
  { key: 'info', label: 'INFO' },
  { key: 'warn', label: 'WARN' },
  { key: 'error', label: 'ERROR' },
  { key: 'debug', label: 'DEBUG' },
]

const filteredLogs = computed(() =>
  activeLevel.value === 'all' ? logs.value : logs.value.filter(l => l.level === activeLevel.value)
)

function levelColor(l: string) {
  const m: Record<string,string> = { info: '#5E6AD2', warn: '#EAB308', error: '#EF4444', debug: '#475569' }
  return m[l] || '#8A8F98'
}
function clearLogs() { logs.value = [] }
function downloadLogs() {
  const text = logs.value.map(l => `${l.time} [${l.level.toUpperCase()}] [${l.source}] ${l.message}`).join('\n')
  const blob = new Blob([text], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url; a.download = 'bdriper.log'; a.click()
  URL.revokeObjectURL(url)
}
</script>
