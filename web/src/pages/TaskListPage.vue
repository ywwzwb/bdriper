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
            <span style="color:#8A8F98;">配置</span>
            <span style="color:#5E6AD2;cursor:pointer;text-decoration:underline;" @click.stop="viewConfig(task.config)">{{ task.config }}</span>
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

    <!-- Floating batch action bar -->
    <div v-if="selectedIds.length" style="position:fixed;bottom:24px;left:50%;transform:translateX(-50%);z-index:40;padding:12px 24px;display:flex;align-items:center;gap:12px;" class="glass">
      <span style="font-size:13px;color:#8A8F98;">已选 {{ selectedIds.length }} 项</span>
      <button class="btn-primary" @click="batchDelete">批量删除</button>
      <button class="btn-ghost" @click="batchPause" v-if="canBatchPause">批量暂停</button>
      <button class="btn-ghost" @click="selectedIds = []">取消选择</button>
    </div>

    <!-- Config detail modal -->
    <Teleport to="body">
      <div v-if="configDetail" class="cfg-modal-overlay" @click.self="configDetail = null">
        <div class="glass" style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:50;width:560px;max-height:80vh;overflow:auto;padding:24px;">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
            <h3 style="font-size:18px;font-weight:700;">{{ configDetail.name }}</h3>
            <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="configDetail = null">关闭</button>
          </div>
          <div style="font-size:13px;color:#8A8F98;margin-bottom:16px;">{{ configDetail.encoder }} | {{ configDetail.mode === 'gpu' ? 'GPU 加速' : 'CPU 编码' }}</div>
          <div style="display:flex;flex-direction:column;gap:6px;">
            <div v-for="(v, k) in configDetail.params" :key="k" style="display:flex;justify-content:space-between;padding:8px 12px;border-radius:8px;background:rgba(255,255,255,0.02);">
              <span style="color:#8A8F98;font-family:'JetBrains Mono',monospace;font-size:13px;">{{ k }}</span>
              <span style="color:#EDEDEF;font-weight:500;font-family:'JetBrains Mono',monospace;font-size:13px;">{{ v }}</span>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <WizardModal :visible="showWizard" @close="showWizard = false" />
  </div>
</template>
<script setup lang="ts">
import { ref, computed } from 'vue'
import { PhFolder } from '@phosphor-icons/vue'
import WizardModal from '../wizard/WizardModal.vue'

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
    outputPath: '/output/Bakemonogatari_EP02.mkv', config: 'x264 高画质 (mbtree on)', pid: null,
  },
  {
    id: 3, name: 'Charlotte EP01', status: 'completed', file: '00000.m2ts',
    progress: 100, eta: '', size: '2.3 GB', sourcePath: '/input/BDROM/BDMV/STREAM/00000.m2ts',
    outputPath: '/output/Charlotte_EP01.mkv', config: 'x265 HQ Anime', pid: null,
  },
  {
    id: 4, name: 'Kill la Kill EP01', status: 'failed', file: '00002.m2ts',
    progress: 23, eta: '', size: '', sourcePath: '/input/BDROM/BDMV/STREAM/00002.m2ts',
    outputPath: '/output/KLK_EP01.mkv', config: 'NVENC 快速编码', pid: null, error_msg: '编码器错误: 未知流类型',
  },
  {
    id: 5, name: 'Steins;Gate EP01', status: 'completed', file: '00003.m2ts',
    progress: 100, eta: '', size: '1.8 GB', sourcePath: '/input/BDROM/BDMV/STREAM/00003.m2ts',
    outputPath: '/output/SG_EP01.mkv', config: 'x264 高画质 (mbtree off)', pid: null,
  },
  {
    id: 6, name: '魔法少女小圆 EP01', status: 'paused', file: '00004.m2ts',
    progress: 68, eta: '', size: '', sourcePath: '/input/BDROM/BDMV/STREAM/00004.m2ts',
    outputPath: '/output/Madoka_EP01.mkv', config: 'x265 均衡', pid: null,
  },
])

const showWizard = ref(false)
const selectedIds = ref<number[]>([])

const configsMap: Record<string, any> = {
  'x265 HQ Anime': { id: 0, name: 'x265 HQ Anime', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 15, preset: 'slower', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', 'no-sao': true, bframes: 16, 'rc-lookahead': 60, subme: 7, merange: 57 } },
  'x264 高画质 (mbtree on)': { id: 1, name: 'x264 高画质 (mbtree on)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 18, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 1, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11, merange: 48 } },
  'x265 均衡': { id: 2, name: 'x265 均衡', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 20, preset: 'medium', 'aq-mode': 3, 'aq-strength': 0.7, deblock: '0:0', bframes: 8, 'rc-lookahead': 40 } },
  'x264 高画质 (mbtree off)': { id: 3, name: 'x264 高画质 (mbtree off)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 16, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 0, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11 } },
  'NVENC 快速编码': { id: 4, name: 'NVENC 快速编码', encoder: 'h264_nvenc', mode: 'gpu', isPreset: false, params: { crf: 20, preset: 'p7', tune: 'hq', rc: 'vbr', b_ref_mode: 'middle', multipass: 'qres', 'aq-strength': 8, lookahead: 32, bframes: 4 } },
  '我的自定义配置': { id: 5, name: '我的自定义配置', encoder: 'x265', mode: 'cpu', isPreset: false, params: { crf: 17, preset: 'slow', 'aq-mode': 3, 'aq-strength': 0.9, deblock: '-1:-1', 'no-sao': true, bframes: 12 } },
}

const configDetail = ref<any>(null)
function viewConfig(name: string) {
  configDetail.value = configsMap[name] || null
}

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

function batchDelete() {
  tasks.value = tasks.value.filter(t => !selectedIds.value.includes(t.id))
  selectedIds.value = []
}
function batchPause() {
  tasks.value.forEach(t => {
    if (selectedIds.value.includes(t.id) && (t.status === 'running' || t.status === 'pending')) {
      t.status = 'paused'
    }
  })
  selectedIds.value = []
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
