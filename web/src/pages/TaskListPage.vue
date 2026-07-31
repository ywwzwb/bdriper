<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
      <h1 style="font-size:28px;font-weight:700;">任务管理</h1>
      <div style="display:flex;gap:8px;">
        <router-link to="/tasks?wizard=open" class="btn-primary" style="text-decoration:none;">+ 新建任务</router-link>
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
        <div style="display:flex;align-items:center;gap:16px;padding:16px 20px;cursor:pointer;" @click="task.expanded = !task.expanded">
          <div style="width:3px;height:40px;border-radius:2px;" :style="{background: statusColor(task.status)}" />

          <div style="flex:1;min-width:0;">
            <div style="display:flex;align-items:center;gap:8px;margin-bottom:4px;">
              <span style="font-weight:600;font-size:15px;">{{ task.name }}</span>
              <span :class="'badge badge-' + task.status">{{ statusLabel(task.status) }}</span>
            </div>
            <div style="font-size:13px;color:#8A8F98;">{{ task.file }}</div>
            <div v-if="task.status === 'running'" class="progress-track" style="margin-top:8px;">
              <div class="progress-fill" :style="{width: task.progress + '%'}" />
            </div>
          </div>

          <div style="text-align:right;flex-shrink:0;">
            <div v-if="task.status === 'running'" style="font-weight:600;font-size:15px;">{{ task.progress }}%</div>
            <div v-if="task.status === 'running'" style="font-size:12px;color:#8A8F98;">剩余 {{ task.eta }}</div>
            <div v-if="task.status === 'completed'" style="font-size:13px;color:#8A8F98;">{{ task.size }}</div>
            <div v-if="task.status === 'failed'" style="font-size:12px;color:#EF4444;max-width:160px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ task.error_msg }}</div>
          </div>

          <div style="display:flex;gap:4px;flex-shrink:0;" @click.stop>
            <button v-if="task.status === 'running'" class="btn-ghost" @click="pauseTask(task)" style="padding:6px 12px;font-size:12px;">暂停</button>
            <button v-if="task.status === 'running' || task.status === 'pending'" class="btn-ghost" @click="cancelTask(task)" style="padding:6px 12px;font-size:12px;color:#EF4444;">取消</button>
            <button v-if="task.status === 'failed'" class="btn-ghost" @click="retryTask(task)" style="padding:6px 12px;font-size:12px;">重试</button>
            <button v-if="task.status === 'completed' || task.status === 'failed' || task.status === 'paused'" class="btn-ghost" @click="deleteTask(task)" style="padding:6px 12px;font-size:12px;color:#EF4444;">删除</button>
          </div>
        </div>

        <div v-if="task.expanded" style="border-top:1px solid rgba(255,255,255,0.06);padding:16px 20px 16px 39px;">
          <div style="display:grid;grid-template-columns:auto 1fr;gap:8px 20px;font-size:13px;">
            <span style="color:#8A8F98;">源文件</span><span style="color:#EDEDEF;">{{ task.sourcePath }}</span>
            <span style="color:#8A8F98;">目标路径</span><span style="color:#EDEDEF;">{{ task.outputPath }}</span>
            <span style="color:#8A8F98;">配置</span><span style="color:#5E6AD2;">{{ task.config }}</span>
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
  </div>
</template>
<script setup lang="ts">
import { ref, computed } from 'vue'
import { PhFolder } from '@phosphor-icons/vue'

interface Task {
  id: number; name: string; status: string; file: string; progress: number
  eta: string; size: string; sourcePath: string; outputPath: string; config: string; pid: number | null
  error_msg?: string; expanded?: boolean
}

const tasks = ref<Task[]>([
  {
    id: 1, name: '四月は君の嘘 EP01', status: 'running', file: '00000.m2ts',
    progress: 45, eta: '12分钟', size: '', sourcePath: '/input/BDROM/BDMV/STREAM/00000.m2ts',
    outputPath: '/output/Shigatsu_EP01.mkv', config: 'x265 HQ Anime', pid: 12345,
  },
  {
    id: 2, name: '化物語 EP02', status: 'pending', file: '00001.m2ts',
    progress: 0, eta: '', size: '', sourcePath: '/input/BDROM/BDMV/STREAM/00001.m2ts',
    outputPath: '/output/Bakemonogatari_EP02.mkv', config: 'x264 High Quality', pid: null,
  },
  {
    id: 3, name: 'Charlotte EP01', status: 'completed', file: '00000.m2ts',
    progress: 100, eta: '', size: '2.3 GB', sourcePath: '/input/BDROM/BDMV/STREAM/00000.m2ts',
    outputPath: '/output/Charlotte_EP01.mkv', config: 'x265 HQ Anime', pid: null,
  },
  {
    id: 4, name: 'Kill la Kill EP01', status: 'failed', file: '00002.m2ts',
    progress: 23, eta: '', size: '', sourcePath: '/input/BDROM/BDMV/STREAM/00002.m2ts',
    outputPath: '/output/KLK_EP01.mkv', config: 'NVENC Fast', pid: null, error_msg: '编码器错误: 未知流类型',
  },
  {
    id: 5, name: 'Steins;Gate EP01', status: 'completed', file: '00003.m2ts',
    progress: 100, eta: '', size: '1.8 GB', sourcePath: '/input/BDROM/BDMV/STREAM/00003.m2ts',
    outputPath: '/output/SG_EP01.mkv', config: 'x264 High Quality', pid: null,
  },
  {
    id: 6, name: '魔法少女小圆 EP01', status: 'paused', file: '00004.m2ts',
    progress: 68, eta: '', size: '', sourcePath: '/input/BDROM/BDMV/STREAM/00004.m2ts',
    outputPath: '/output/Madoka_EP01.mkv', config: 'x265 HQ Anime', pid: null,
  },
])

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

function statusColor(s: string) {
  const m: Record<string,string> = { running: '#5E6AD2', completed: '#22C55E', failed: '#EF4444', pending: '#EAB308', paused: '#94A3B8' }
  return m[s] || '#475569'
}
function statusLabel(s: string) {
  const m: Record<string,string> = { running: '运行中', completed: '已完成', failed: '失败', pending: '等待', paused: '已暂停' }
  return m[s] || s
}
function pauseTask(t: Task) { t.status = 'paused' }
function cancelTask(t: Task) { t.status = 'failed'; t.error_msg = '用户取消' }
function retryTask(t: Task) { t.status = 'pending'; t.progress = 0; t.error_msg = '' }
function deleteTask(t: Task) { tasks.value = tasks.value.filter(x => x.id !== t.id) }
function deleteCompleted() { tasks.value = tasks.value.filter(t => t.status !== 'completed') }
</script>
