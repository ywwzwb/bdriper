<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
      <h1 style="font-size:28px;font-weight:700;">任务管理</h1>
      <div style="display:flex;gap:8px;">
        <button class="btn-primary" @click="showWizard = true">+ 新建任务</button>
        <button class="btn-ghost" @click="deleteCompleted">清空已完成</button>
      </div>
    </div>

    <div style="display:flex;gap:8px;margin-bottom:24px;flex-wrap:wrap;">
      <button v-for="f in filters" :key="f.key" class="pill" :class="{ active: activeFilter === f.key }" @click="activeFilter = f.key">
        {{ f.label }}
        <span v-if="f.count" style="font-size:11px;opacity:0.6;">{{ f.count }}</span>
      </button>
    </div>

    <div v-if="filteredTasks.length" style="display:flex;flex-direction:column;gap:12px;">
      <div v-for="task in filteredTasks" :key="task.id" class="glass" style="overflow:hidden;">
        <div style="display:flex;align-items:center;gap:12px;padding:16px 20px;cursor:pointer;" @click="task.expanded = !task.expanded">
          <input type="checkbox" v-model="selectedIds" :value="task.id" style="accent-color:#5E6AD2;width:16px;height:16px;flex-shrink:0;" @click.stop />

          <div style="width:3px;height:40px;border-radius:2px;" :style="{background: statusColor(task.status)}" />

          <div style="flex:1;min-width:0;">
            <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
              <span style="font-weight:600;font-size:15px;">{{ task.name }}</span>
              <span :class="'badge badge-' + task.status">{{ statusLabel(task.status) }}</span>
            </div>
            <div style="font-size:13px;color:#8A8F98;">{{ task.source_path || task.file }}</div>
            <div v-if="task.status === 'running'" class="progress-track" style="margin-top:8px;">
              <div class="progress-fill" :style="{width: Math.round(task.progress) + '%'}" />
            </div>
          </div>
          <div style="text-align:right;flex-shrink:0;">
            <div v-if="task.status === 'running'" style="font-weight:600;font-size:15px;">{{ task.progress.toFixed(1) }}%</div>
            <div v-if="task.status === 'running'" style="font-size:12px;color:#8A8F98;">剩余 {{ formatETA(task.estimated_eta || task.eta) }}</div>
            <div v-if="task.status === 'completed'" style="font-size:13px;color:#8A8F98;">{{ task.size }}</div>
            <div v-if="task.status === 'failed'" style="font-size:12px;color:#EF4444;max-width:160px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ task.error_msg }}</div>
          </div>

          <div style="display:flex;gap:4px;flex-shrink:0;" @click.stop>
            <button v-if="task.status === 'running'" class="btn-ghost" @click="pauseTask(task)" style="padding:6px 12px;font-size:12px;">暂停</button>
            <button v-if="task.status === 'paused'" class="btn-ghost" @click="resumeTask(task)" style="padding:6px 12px;font-size:12px;">恢复</button>
            <button v-if="task.status === 'running' || task.status === 'pending'" class="btn-ghost" @click="cancelTask(task)" style="padding:6px 12px;font-size:12px;color:#EF4444;">取消</button>
            <button v-if="task.status === 'failed'" class="btn-ghost" @click="retryTask(task)" style="padding:6px 12px;font-size:12px;">重试</button>
            <button v-if="task.status === 'completed' || task.status === 'failed' || task.status === 'paused' || task.status === 'cancelled'" class="btn-ghost" @click="deleteTask(task)" style="padding:6px 12px;font-size:12px;color:#EF4444;">删除</button>
          </div>
        </div>

        <div v-if="task.expanded" style="border-top:1px solid rgba(255,255,255,0.06);padding:16px 20px 16px 39px;">
          <div style="display:grid;grid-template-columns:auto 1fr;gap:8px 20px;font-size:13px;">
            <span style="color:#8A8F98;">源文件</span><span style="color:#EDEDEF;">{{ task.source_path || task.sourcePath }}</span>
            <span style="color:#8A8F98;">目标路径</span><span style="color:#EDEDEF;">{{ task.output_path || task.outputPath }}</span>
            <span style="color:#8A8F98;">配置</span>
            <span style="color:#5E6AD2;cursor:pointer;text-decoration:underline;" @click.stop="viewConfig(task.config_name || task.config)">{{ task.config_name || task.config }}</span>
            <span style="color:#8A8F98;">PID</span><span style="color:#EDEDEF;">{{ task.pid || '—' }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-else style="text-align:center;padding:80px 20px;">
      <div style="font-size:48px;margin-bottom:16px;opacity:0.3;">
        <PhFolder size="48" color="#8A8F98" style="opacity:0.3;" />
      </div>
      <div style="font-size:16px;color:#8A8F98;">暂无任务</div>
      <div style="font-size:13px;color:#8A8F98;margin-top:4px;">点击「新建任务」开始转码</div>
    </div>

    <div v-if="selectedIds.length" style="position:fixed;bottom:24px;left:50%;transform:translateX(-50%);z-index:40;padding:12px 24px;display:flex;align-items:center;gap:12px;" class="glass">
      <span style="font-size:13px;color:#8A8F98;">已选 {{ selectedIds.length }} 项</span>
      <button class="btn-primary" @click="batchDelete">批量删除</button>
      <button class="btn-ghost" @click="batchPause" v-if="canBatchPause">批量暂停</button>
      <button class="btn-ghost" @click="selectedIds = []">取消选择</button>
    </div>

    <Teleport to="body">
      <div v-if="configDetail" class="cfg-modal-overlay" @click.self="configDetail = null">
        <div class="glass" style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:60;width:600px;max-height:80vh;overflow-y:auto;">
          <div style="display:flex;align-items:center;justify-content:space-between;padding:16px 20px;border-bottom:1px solid rgba(255,255,255,0.06);">
            <div style="display:flex;align-items:center;gap:10px;">
              <h3 style="font-weight:600;font-size:16px;">{{ configDetail.name }}</h3>
              <span v-if="configDetail.isPreset" class="badge badge-pending">内置预设</span>
              <span class="badge" :class="configDetail.encoder_type === 'gpu' || configDetail.mode === 'gpu' ? 'badge-running' : 'badge-pending'">{{ configDetail.encoder_type === 'gpu' || configDetail.mode === 'gpu' ? 'GPU' : 'CPU' }}</span>
            </div>
            <button @click="configDetail = null" style="color:#8A8F98;font-size:20px;cursor:pointer;background:none;border:none;">✕</button>
          </div>
          <div style="padding:20px;">
            <div style="margin-bottom:16px;">
              <div style="color:#8A8F98;font-size:12px;">编码器</div>
              <div style="font-weight:500;font-size:15px;">{{ configDetail.encoder }}</div>
            </div>
            <div style="margin-bottom:16px;">
              <div style="color:#8A8F98;font-size:12px;margin-bottom:8px;">视频参数</div>
              <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px 16px;">
                <div v-for="(v,k) in configDetail.params" :key="k" style="display:flex;justify-content:space-between;padding:4px 0;border-bottom:1px solid rgba(255,255,255,0.04);">
                  <span style="color:#8A8F98;font-size:13px;">{{ k }}</span>
                  <span style="font-size:13px;font-family:monospace;">{{ v }}</span>
                </div>
              </div>
            </div>
            <div v-if="configDetail.audio" style="margin-bottom:16px;">
              <div style="color:#8A8F98;font-size:12px;margin-bottom:8px;">音频参数</div>
              <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px 16px;">
                <div v-for="(v,k) in configDetail.audio" :key="k" style="display:flex;justify-content:space-between;padding:4px 0;border-bottom:1px solid rgba(255,255,255,0.04);">
                  <span style="color:#8A8F98;font-size:13px;">{{ k }}</span>
                  <span style="font-size:13px;">{{ v }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <WizardModal :visible="showWizard" @close="showWizard = false" />
  </div>
</template>
<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { PhFolder } from '@phosphor-icons/vue'
import WizardModal from '../wizard/WizardModal.vue'
import { api } from '@/api'
import { on as wsOn } from '@/ws'

const tasks = ref<any[]>([])
const showWizard = ref(false)
const selectedIds = ref<number[]>([])
const configDetail = ref<any>(null)

const activeFilter = ref('all')

const filters = computed(() => [
  { key: 'all', label: '全部', count: tasks.value.length },
  { key: 'running', label: '运行中', count: tasks.value.filter(t => t.status === 'running').length },
  { key: 'completed', label: '已完成', count: tasks.value.filter(t => t.status === 'completed').length },
  { key: 'failed', label: '失败', count: tasks.value.filter(t => t.status === 'failed').length },
  { key: 'pending', label: '等待', count: tasks.value.filter(t => t.status === 'pending').length },
  { key: 'paused', label: '已暂停', count: tasks.value.filter(t => t.status === 'paused').length },
])

const filteredTasks = computed(() =>
  activeFilter.value === 'all' ? tasks.value : tasks.value.filter(t => t.status === activeFilter.value)
)

const canBatchPause = computed(() => {
  const selected = tasks.value.filter(t => selectedIds.value.includes(t.id))
  return selected.some(t => t.status === 'running' || t.status === 'pending')
})

function formatETA(eta: string | number): string {
  if (!eta) return ''
  const sec = typeof eta === 'string' ? parseInt(eta) || 0 : eta
  if (sec <= 0) return ''
  if (sec < 60) return '即将完成'
  if (sec < 3600) return `${Math.floor(sec / 60)}分钟`
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  return `${h}小时${m}分钟`
}

let pollTimer: ReturnType<typeof setInterval> | null = null

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})

onMounted(() => {
  fetchTasks()
  wsOn('progress', (data: any) => {
    const task = tasks.value.find(t => t.id === data.task_id)
    if (task) {
      task.progress = data.progress * 100
    }
  })
  // Poll every 3 seconds for status changes / new tasks
  pollTimer = setInterval(fetchTasks, 3000)
})

watch(activeFilter, () => fetchTasks())

async function fetchTasks() {
  try {
    const data = await api.tasks.list(activeFilter.value)
    const expandedIds = new Set(tasks.value.filter(t => t.expanded).map(t => t.id))
    tasks.value = (data || []).map(t => ({ ...t, progress: (t.progress || 0) * 100, expanded: expandedIds.has(t.id) }))
  } catch {}
}

function statusColor(s: string) {
  const m: Record<string,string> = { running: '#5E6AD2', completed: '#22C55E', failed: '#EF4444', pending: '#EAB308', paused: '#94A3B8' }
  return m[s] || '#475569'
}
function statusLabel(s: string) {
  const m: Record<string,string> = { running: '运行中', completed: '已完成', failed: '失败', pending: '等待', paused: '已暂停' }
  return m[s] || s
}

async function pauseTask(task: any) {
  try { await api.tasks.update(task.id, { status: 'paused' }); fetchTasks() } catch {}
}

async function resumeTask(task: any) {
  try { await api.tasks.update(task.id, { status: 'pending' }); fetchTasks() } catch {}
}
async function cancelTask(task: any) {
  try { await api.tasks.update(task.id, { status: 'cancelled' }); fetchTasks() } catch {}
}
async function retryTask(task: any) {
  try { await api.tasks.retry(task.id); fetchTasks() } catch {}
}
async function deleteTask(task: any) {
  try { await api.tasks.delete(task.id); fetchTasks() } catch {}
}
async function deleteCompleted() {
  try { await api.tasks.deleteCompleted(); fetchTasks() } catch {}
}

async function batchDelete() {
  try { await api.tasks.batch(selectedIds.value, 'delete'); selectedIds.value = []; fetchTasks() } catch {}
}
async function batchPause() {
  try { await api.tasks.batch(selectedIds.value, 'pause'); selectedIds.value = []; fetchTasks() } catch {}
}

async function viewConfig(name: string) {
  try {
    const configs = await api.configs.list()
    const cfg = configs.find((c: any) => c.name === name)
    if (cfg) {
      configDetail.value = {
        name: cfg.name,
        encoder: cfg.video_encoder || cfg.encoder,
        mode: cfg.encoder_type || cfg.mode || 'cpu',
        isPreset: cfg.is_builtin || false,
        params: typeof cfg.video_params === 'string' ? JSON.parse(cfg.video_params) : (cfg.video_params || {}),
        audio: typeof cfg.audio_tracks === 'string' ? JSON.parse(cfg.audio_tracks) : (cfg.audio_tracks || {}),
      }
    } else {
      configDetail.value = null
    }
  } catch {}
}
</script>
<style scoped>
.cfg-modal-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  z-index: 49;
}
</style>
