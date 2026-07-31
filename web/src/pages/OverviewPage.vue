<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-3xl font-bold text-fg">概览</h1>
        <p class="text-muted mt-1">系统资源与任务状态</p>
      </div>
      <RouterLink to="/tasks?wizard=open" class="px-5 py-2.5 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 flex items-center gap-2 shadow-lg shadow-accent/20">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
        <span>新建任务</span>
      </RouterLink>
    </div>

    <div v-if="loading" class="text-muted">加载中...</div>
    <div v-else-if="error" class="text-destructive mb-4">{{ error }}</div>

    <div v-else class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      <div class="bg-card border border-border/50 rounded-xl p-5 hover:border-accent/30 transition-all duration-200">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-lg bg-accent/10 flex items-center justify-center">
            <svg class="w-5 h-5 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="1" x2="15" y2="4"/><line x1="9" y1="20" x2="9" y2="23"/><line x1="15" y1="20" x2="15" y2="23"/><line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/><line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/></svg>
          </div>
          <span class="text-sm text-muted">CPU 使用率</span>
        </div>
        <div class="flex items-baseline gap-2">
          <span class="text-3xl font-bold text-fg">{{ animatedCPU }}</span>
          <span class="text-sm text-muted">%</span>
        </div>
        <div class="mt-3 w-full bg-muted rounded-full h-1.5">
          <div class="bg-accent rounded-full h-1.5 transition-all duration-500" :style="{ width: Math.min(animatedCPU, 100) + '%' }" />
        </div>
      </div>

      <div class="bg-card border border-border/50 rounded-xl p-5 hover:border-accent/30 transition-all duration-200">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-lg bg-accent/10 flex items-center justify-center">
            <svg class="w-5 h-5 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
          </div>
          <span class="text-sm text-muted">GPU</span>
        </div>
        <div class="flex items-baseline gap-2">
          <span class="text-3xl font-bold text-fg">{{ stats.gpu_available ? (stats.gpu_usage > 0 ? animatedGPU : '--') : 'N/A' }}</span>
          <span v-if="stats.gpu_available && stats.gpu_usage > 0" class="text-sm text-muted">%</span>
        </div>
        <div class="mt-1 text-xs text-muted">{{ stats.gpu_available ? stats.gpu_vendor : '无可用 GPU' }}</div>
      </div>

      <div class="bg-card border border-border/50 rounded-xl p-5 hover:border-accent/30 transition-all duration-200">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center">
            <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><polyline points="12,6 12,12 16,14"/></svg>
          </div>
          <span class="text-sm text-muted">运行中</span>
        </div>
        <div class="flex items-baseline gap-2">
          <span class="text-3xl font-bold text-blue-400">{{ animatedRunning }}</span>
          <span class="text-sm text-muted">/ {{ stats.total }} 个任务</span>
        </div>
        <div v-if="stats.running > 0" class="mt-3 flex gap-1">
          <span v-for="i in Math.min(stats.running, 5)" :key="i" class="w-2 h-2 rounded-full bg-blue-400 animate-pulse" />
        </div>
      </div>

      <div class="bg-card border border-border/50 rounded-xl p-5 hover:border-accent/30 transition-all duration-200">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-lg bg-purple-500/10 flex items-center justify-center">
            <svg class="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="4" height="4"/><rect x="9" y="15" width="4" height="4"/><rect x="14" y="9" width="2" height="10"/></svg>
          </div>
          <span class="text-sm text-muted">内存</span>
        </div>
        <div class="flex items-baseline gap-2">
          <span class="text-3xl font-bold text-fg">{{ animatedMem }}</span>
          <span class="text-sm text-muted">MB</span>
        </div>
        <div class="mt-1 text-xs text-muted">{{ stats.goroutines }} 个协程</div>
      </div>
    </div>

    <div v-if="stats.running > 0" class="bg-card border border-border/50 rounded-xl p-4">
      <div class="flex items-center gap-3 text-sm">
        <span class="w-2 h-2 rounded-full bg-accent animate-pulse" />
        <span class="text-fg">正在运行 {{ stats.running }} 个转码任务</span>
        <RouterLink to="/tasks" class="text-accent hover:underline ml-auto">查看详情</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watchEffect } from 'vue'
import { api } from '@/api'

const stats = ref<any>({
  cpu_usage: 0,
  gpu_available: false,
  gpu_vendor: '',
  gpu_usage: 0,
  running: 0,
  total: 0,
  goroutines: 0,
  mem_mb: 0,
})
const loading = ref(true)
const error = ref('')
const animatedCPU = ref(0)
const animatedGPU = ref(0)
const animatedRunning = ref(0)
const animatedMem = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

function animateTo(target: number, ref: any, duration = 400) {
  const start = ref.value
  const diff = target - start
  const startTime = performance.now()
  function step(ts: number) {
    const elapsed = ts - startTime
    const progress = Math.min(elapsed / duration, 1)
    ref.value = Math.round(start + diff * progress)
    if (progress < 1) requestAnimationFrame(step)
  }
  requestAnimationFrame(step)
}

async function fetchStatus() {
  try {
    const data = await api.overview()
    const old = stats.value
    stats.value = { ...data }
    animateTo(data.cpu_usage ?? 0, animatedCPU)
    animateTo(data.gpu_usage ?? 0, animatedGPU)
    animateTo(data.running ?? 0, animatedRunning)
    animateTo(data.mem_mb ?? 0, animatedMem)
    error.value = ''
  } catch (e: any) {
    error.value = e.message || '加载失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStatus()
  timer = setInterval(fetchStatus, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
