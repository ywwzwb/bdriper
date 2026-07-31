<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-semibold">任务管理</h1>
      <div class="flex gap-3">
        <button @click="deleteCompleted" class="px-3 py-1.5 text-sm border border-border rounded-lg hover:bg-muted transition" v-if="tasks.some(t => t.status === 'completed')">一键删除已完成</button>
        <button @click="showWizard = true" class="inline-flex items-center gap-2 px-4 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition"><span class="text-lg leading-none">+</span> 新建任务</button>
      </div>
    </div>

    <div class="flex gap-1 mb-4 bg-card border border-border rounded-lg p-1 w-fit">
      <button v-for="s in statuses" :key="s.value" @click="filter = s.value" class="px-3 py-1.5 rounded-md text-sm transition" :class="filter === s.value ? 'bg-accent text-black' : 'text-muted hover:text-fg'">{{ s.label }}</button>
    </div>

    <div v-if="error" class="text-destructive mb-4">{{ error }}</div>

    <div v-if="selectedIds.length > 0" class="flex gap-2 mb-4 p-3 bg-card border border-border rounded-lg items-center">
      <span class="text-sm text-muted">已选择 {{ selectedIds.length }} 项</span>
      <button @click="batchAction('pause')" class="px-3 py-1 text-sm border border-border rounded hover:bg-muted transition">批量暂停</button>
      <button @click="batchAction('delete')" class="px-3 py-1 text-sm border border-destructive text-destructive rounded hover:bg-destructive/10 transition">批量删除</button>
    </div>

    <div class="space-y-2">
      <div v-for="task in filteredTasks" :key="task.id" class="bg-card border border-border rounded-lg">
        <div class="flex items-center gap-3 p-3 cursor-pointer" @click="toggleExpand(task.id)">
          <input type="checkbox" :checked="selectedIds.includes(task.id)" @change="toggleSelect(task.id)" @click.stop class="w-4 h-4 rounded border-border bg-muted accent-accent" />
          <span class="flex-1 font-medium truncate">{{ task.name }}</span>
          <span class="px-2 py-0.5 rounded text-xs font-medium" :class="statusClass(task.status)">{{ statusLabel(task.status) }}</span>
          <div class="w-24 bg-muted rounded-full h-1.5 hidden sm:block">
            <div class="bg-accent h-1.5 rounded-full transition-all" :style="{ width: (task.progress ?? 0) + '%' }"></div>
          </div>
          <span class="text-xs text-muted w-16 text-right">{{ task.eta || '--' }}</span>
          <div class="flex gap-1">
            <button v-if="task.status === 'running'" @click.stop="action(task.id, 'pause')" class="p-1 rounded hover:bg-muted text-muted hover:text-fg" title="暂停"><span>&#9646;&#9646;</span></button>
            <button v-if="task.status === 'failed'" @click.stop="action(task.id, 'retry')" class="p-1 rounded hover:bg-muted text-muted hover:text-fg" title="重试"><span>&#8635;</span></button>
            <button @click.stop="action(task.id, 'delete')" class="p-1 rounded hover:bg-muted text-muted hover:text-destructive" title="删除"><span>&#10005;</span></button>
          </div>
        </div>
        <div v-if="expanded.has(task.id)" class="px-3 pb-3 border-t border-border pt-2">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-sm">
            <div><span class="text-muted">源文件:</span> {{ task.source_path || '--' }}</div>
            <div><span class="text-muted">目标路径:</span> {{ task.output_path || '--' }}</div>
            <div><span class="text-muted">配置:</span> {{ task.config_name || '--' }}</div>
            <div><span class="text-muted">PID:</span> {{ task.pid ?? '--' }}</div>
            <div v-if="task.error" class="sm:col-span-2"><span class="text-muted">错误:</span> <span class="text-destructive">{{ task.error }}</span></div>
          </div>
        </div>
      </div>

      <div v-if="filteredTasks.length === 0 && !error" class="text-center text-muted py-12">暂无任务，点击「新建任务」开始</div>
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
  pending: '等待',
  paused: '已暂停',
}

function statusLabel(status: string): string {
  return statusLabels[status] || status
}

function statusClass(status: string) {
  const map: Record<string, string> = {
    running: 'bg-accent/20 text-accent',
    completed: 'bg-blue-500/20 text-blue-400',
    failed: 'bg-destructive/20 text-destructive',
    pending: 'bg-yellow-500/20 text-yellow-400',
    paused: 'bg-yellow-500/20 text-yellow-400',
  }
  return map[status] || 'bg-muted text-muted'
}

function toggleExpand(id: number) {
  if (expanded.value.has(id)) expanded.value.delete(id)
  else expanded.value.add(id)
  expanded.value = new Set(expanded.value)
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

function closeWizard() {
  showWizard.value = false
  fetchTasks()
}

watch(filter, () => fetchTasks())

onMounted(() => {
  fetchTasks()
  if (route.query.wizard === 'open') showWizard.value = true
  wsOn('task_updated', fetchTasks)
})
</script>
