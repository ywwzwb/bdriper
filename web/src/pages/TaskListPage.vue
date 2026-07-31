<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-3xl font-bold text-fg">任务管理</h1>
        <p class="text-muted mt-1">管理所有转码任务</p>
      </div>
      <div class="flex gap-3">
        <button v-if="tasks.some(t => t.status === 'completed')" @click="deleteCompleted" class="px-4 py-2 text-sm border border-border/50 rounded-lg hover:bg-muted transition-all duration-200">一键删除已完成</button>
        <button @click="showWizard = true" class="px-5 py-2.5 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 flex items-center gap-2 shadow-lg shadow-accent/20">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span>新建任务</span>
        </button>
      </div>
    </div>

    <div class="flex gap-1.5 mb-6 bg-muted/50 rounded-xl p-1 w-fit">
      <button v-for="status in statuses" :key="status.value" @click="filter = status.value" class="px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200" :class="filter === status.value ? 'bg-card text-fg shadow-sm' : 'text-muted hover:text-fg'">
        {{ status.label }}
      </button>
    </div>

    <div v-if="error" class="text-destructive mb-4">{{ error }}</div>

    <transition name="slide-up">
      <div v-if="selectedIds.length > 0" class="sticky bottom-0 z-10 flex gap-3 p-4 bg-card border border-border/50 rounded-xl mb-4 items-center shadow-xl">
        <span class="text-sm text-muted">已选择 {{ selectedIds.length }} 项</span>
        <button @click="batchAction('pause')" class="px-3 py-1.5 text-sm border border-border/50 rounded-lg hover:bg-muted transition-all duration-200">批量暂停</button>
        <button @click="batchAction('delete')" class="px-3 py-1.5 text-sm border border-destructive/50 text-destructive rounded-lg hover:bg-destructive/10 transition-all duration-200">批量删除</button>
      </div>
    </transition>

    <div v-if="filteredTasks.length === 0 && !error" class="flex flex-col items-center justify-center py-20 text-muted">
      <svg class="w-16 h-16 mb-4 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="9" y1="9" x2="15" y2="9"/><line x1="9" y1="13" x2="13" y2="13"/></svg>
      <p class="text-lg">暂无任务</p>
      <p class="text-sm mt-1">点击「新建任务」开始转码</p>
    </div>

    <div class="space-y-2">
      <div v-for="task in filteredTasks" :key="task.id" class="bg-card border border-border/50 rounded-xl overflow-hidden hover:border-accent/30 transition-all duration-200">
        <div class="flex items-center gap-3 p-4 cursor-pointer" @click="toggleExpand(task.id)">
          <div class="w-1 h-12 rounded-full flex-shrink-0" :class="statusStripe(task.status)" />
          <input type="checkbox" :checked="selectedIds.includes(task.id)" @change="toggleSelect(task.id)" @click.stop class="w-4 h-4 rounded border-border bg-muted accent-accent cursor-pointer" />

          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1">
              <span class="font-medium text-fg truncate">{{ task.name }}</span>
              <span class="px-2 py-0.5 rounded-md text-xs font-medium flex-shrink-0" :class="statusBadge(task.status)">
                <span class="inline-block w-1.5 h-1.5 rounded-full mr-1.5" :class="statusDot(task.status)" />
                {{ statusLabel(task.status) }}
              </span>
            </div>
            <div class="flex items-center gap-3 text-xs text-muted">
              <span class="truncate">{{ task.source_path || '--' }}</span>
            </div>
            <div v-if="task.status === 'running'" class="mt-2 w-full bg-muted rounded-full h-1.5">
              <div class="bg-accent h-1.5 rounded-full transition-all duration-500" :style="{ width: (task.progress ?? 0) + '%' }" />
            </div>
          </div>

          <div class="flex items-center gap-3 flex-shrink-0">
            <span v-if="task.status === 'running'" class="text-xs font-mono text-muted">{{ task.progress ?? 0 }}%</span>
            <span v-if="task.eta && task.status === 'running'" class="text-xs text-muted hidden sm:inline">{{ task.eta }}</span>
            <div class="flex gap-1">
              <button v-if="task.status === 'running'" @click.stop="action(task.id, 'pause')" class="p-1.5 rounded-lg hover:bg-muted text-muted hover:text-fg transition-all duration-200" title="暂停">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="6" y="4" width="3" height="16"/><rect x="15" y="4" width="3" height="16"/></svg>
              </button>
              <button v-if="task.status === 'failed'" @click.stop="action(task.id, 'retry')" class="p-1.5 rounded-lg hover:bg-muted text-muted hover:text-fg transition-all duration-200" title="重试">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><polyline points="1,4 1,10 7,10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
              </button>
              <button @click.stop="action(task.id, 'delete')" class="p-1.5 rounded-lg hover:bg-destructive/10 text-muted hover:text-destructive transition-all duration-200" title="删除">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <svg class="w-4 h-4 text-muted transition-transform duration-200 flex-shrink-0" :class="expanded.has(task.id) ? 'rotate-180' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><polyline points="6,9 12,15 18,9"/></svg>
          </div>
        </div>

        <transition name="expand">
          <div v-if="expanded.has(task.id)" class="px-4 pb-4 border-t border-border/50">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mt-3 text-sm">
              <div class="flex flex-col gap-0.5">
                <span class="text-xs text-muted">源文件</span>
                <span class="text-fg truncate">{{ task.source_path || '--' }}</span>
              </div>
              <div class="flex flex-col gap-0.5">
                <span class="text-xs text-muted">目标路径</span>
                <span class="text-fg truncate">{{ task.output_path || '--' }}</span>
              </div>
              <div class="flex flex-col gap-0.5">
                <span class="text-xs text-muted">编码配置</span>
                <span class="text-fg">{{ task.config_name || '--' }}</span>
              </div>
              <div class="flex flex-col gap-0.5">
                <span class="text-xs text-muted">进程 PID</span>
                <span class="text-fg">{{ task.pid ?? '--' }}</span>
              </div>
              <div v-if="task.error" class="sm:col-span-2 flex flex-col gap-0.5">
                <span class="text-xs text-muted">错误信息</span>
                <span class="text-destructive text-xs font-mono">{{ task.error }}</span>
              </div>
            </div>
          </div>
        </transition>
      </div>
    </div>

    <Teleport to="body">
      <WizardContainer v-if="showWizard" @close="closeWizard" />
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api'
import { on as wsOn } from '@/ws'
import WizardContainer from '@/wizard/WizardContainer.vue'

const route = useRoute()

const tasks = ref<any[]>([])
const error = ref('')
const filter = ref('all')
const selectedIds = ref<number[]>([])
const expanded = ref(new Set<number>())
const showWizard = ref(false)

const statuses = [
  { label: '全部', value: 'all' },
  { label: '运行中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
  { label: '等待', value: 'pending' },
]

const filteredTasks = computed(() => {
  if (filter.value === 'all') return tasks.value
  return tasks.value.filter(t => t.status === filter.value)
})

const statusLabels: Record<string, string> = {
  running: '运行中',
  completed: '已完成',
  failed: '失败',
  pending: '等待中',
  paused: '已暂停',
}

function statusLabel(status: string): string { return statusLabels[status] || status }

function statusStripe(status: string) {
  const map: Record<string, string> = {
    running: 'bg-accent',
    completed: 'bg-blue-400',
    failed: 'bg-destructive',
    pending: 'bg-yellow-400',
    paused: 'bg-yellow-400',
  }
  return map[status] || 'bg-muted'
}

function statusDot(status: string) {
  const map: Record<string, string> = {
    running: 'bg-accent',
    completed: 'bg-blue-400',
    failed: 'bg-destructive',
    pending: 'bg-yellow-400',
    paused: 'bg-yellow-400',
  }
  return map[status] || 'bg-muted'
}

function statusBadge(status: string) {
  const map: Record<string, string> = {
    running: 'bg-accent/10 text-accent',
    completed: 'bg-blue-500/10 text-blue-400',
    failed: 'bg-destructive/10 text-destructive',
    pending: 'bg-yellow-500/10 text-yellow-400',
    paused: 'bg-yellow-500/10 text-yellow-400',
  }
  return map[status] || 'bg-muted text-muted'
}

function toggleExpand(id: number) {
  const next = new Set(expanded.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expanded.value = next
}

function toggleSelect(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

async function fetchTasks() {
  try {
    tasks.value = await api.tasks.list(filter.value === 'all' ? undefined : filter.value)
    error.value = ''
  } catch (e: any) {
    error.value = e.message || '加载失败'
  }
}

async function action(id: number, act: string) {
  try {
    if (act === 'delete') await api.tasks.delete(id)
    else if (act === 'pause') await api.tasks.update(id, { status: 'paused' })
    else if (act === 'retry') await api.tasks.retry(id)
    await fetchTasks()
  } catch (e: any) {
    error.value = e.message || '操作失败'
  }
}

async function batchAction(act: string) {
  try {
    await api.tasks.batch(selectedIds.value, act)
    selectedIds.value = []
    await fetchTasks()
  } catch (e: any) {
    error.value = e.message || '批量操作失败'
  }
}

async function deleteCompleted() {
  try {
    await api.tasks.deleteCompleted()
    await fetchTasks()
  } catch (e: any) {
    error.value = e.message || '删除失败'
  }
}

function closeWizard() { showWizard.value = false; fetchTasks() }

watch(filter, () => fetchTasks())

onMounted(() => {
  fetchTasks()
  if (route.query.wizard === 'open') showWizard.value = true
  wsOn('task_updated', fetchTasks)
})
</script>

<style scoped>
.slide-up-enter-active, .slide-up-leave-active { transition: all 0.3s ease; }
.slide-up-enter-from, .slide-up-leave-to { transform: translateY(20px); opacity: 0; }

.expand-enter-active, .expand-leave-active { transition: all 0.25s ease; overflow: hidden; }
.expand-enter-from, .expand-leave-to { max-height: 0; opacity: 0; }
.expand-enter-to, .expand-leave-from { max-height: 500px; opacity: 1; }
</style>
