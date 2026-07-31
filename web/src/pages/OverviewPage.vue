<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-semibold">概览</h1>
      <RouterLink to="/tasks?wizard=open" class="inline-flex items-center gap-2 px-4 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition">
        <span class="text-lg leading-none">+</span> 新建任务
      </RouterLink>
    </div>

    <div v-if="loading" class="text-muted">加载中...</div>

    <div v-else-if="error" class="text-destructive">{{ error }}</div>

    <div v-else class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div class="bg-card border border-border rounded-lg p-4">
        <div class="text-muted text-sm">运行中任务</div>
        <div class="text-3xl font-bold mt-1">{{ data?.running_tasks ?? 0 }}</div>
        <div class="text-muted text-xs mt-1">/ {{ data?.total_tasks ?? 0 }} 总计</div>
      </div>

      <div class="bg-card border border-border rounded-lg p-4">
        <div class="text-muted text-sm">协程数</div>
        <div class="text-3xl font-bold mt-1">{{ data?.goroutines ?? 0 }}</div>
      </div>

      <div class="bg-card border border-border rounded-lg p-4">
        <div class="text-muted text-sm">内存</div>
        <div class="text-3xl font-bold mt-1">{{ data?.memory_mb ?? 0 }} MB</div>
      </div>

      <div class="bg-card border border-border rounded-lg p-4">
        <div class="text-muted text-sm">CPU 使用率</div>
        <div class="text-3xl font-bold mt-1">{{ data?.cpu_percent ?? 0 }}%</div>
        <div class="w-full bg-muted rounded-full h-2 mt-2">
          <div class="bg-accent h-2 rounded-full transition-all" :style="{ width: (data?.cpu_percent ?? 0) + '%' }"></div>
        </div>
      </div>

      <div v-if="data?.gpu_name" class="bg-card border border-border rounded-lg p-4">
        <div class="text-muted text-sm">GPU</div>
        <div class="text-lg font-semibold mt-1 truncate">{{ data.gpu_name }}</div>
        <div class="text-muted text-xs mt-1">使用率: {{ data.gpu_percent ?? 0 }}%</div>
        <div class="w-full bg-muted rounded-full h-2 mt-2">
          <div class="bg-accent h-2 rounded-full transition-all" :style="{ width: (data.gpu_percent ?? 0) + '%' }"></div>
        </div>
      </div>

      <div class="bg-card border border-border rounded-lg p-4">
        <div class="text-muted text-sm">运行时间</div>
        <div class="text-3xl font-bold mt-1">{{ uptime }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { api } from '@/api'

const data = ref<any>(null)
const loading = ref(true)
const error = ref('')
let timer: ReturnType<typeof setInterval> | null = null

const uptime = computed(() => {
  if (!data.value?.uptime_seconds) return '0s'
  const s = data.value.uptime_seconds
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${s % 60}s`
  return `${s}s`
})

async function fetchStatus() {
  try {
    data.value = await api.overview()
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
