<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:32px;">
      <div>
        <h1 style="font-size:28px;font-weight:700;color:#EDEDEF;">概览</h1>
        <p style="color:#8A8F98;margin-top:4px;font-size:14px;">系统资源与任务状态</p>
      </div>
      <button class="btn-primary" style="display:inline-flex;align-items:center;gap:6px;" @click="showWizard = true">
        <span style="font-size:18px;">+</span> 新建任务
      </button>
    </div>

    <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:16px;margin-bottom:32px;">
      <div class="glass" style="padding:20px;">
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;">
          <div class="stat-icon" style="background:rgba(94,106,210,0.12);">
            <PhCpu :size="18" color="#5E6AD2" />
          </div>
          <span style="font-size:13px;color:#8A8F98;">CPU 使用率</span>
        </div>
        <div style="display:flex;align-items:baseline;gap:6px;margin-bottom:8px;">
          <span style="font-size:32px;font-weight:700;color:#EDEDEF;">{{ stats.cpu_usage.toFixed(1) }}</span>
          <span style="font-size:13px;color:#8A8F98;">%</span>
        </div>
        <div class="progress-track">
          <div class="progress-fill" :style="{width: stats.cpu_usage + '%'}" />
        </div>
      </div>

      <div class="glass" style="padding:20px;">
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;">
          <div class="stat-icon" style="background:rgba(34,197,94,0.12);">
            <PhMonitor :size="18" color="#22C55E" />
          </div>
          <span style="font-size:13px;color:#8A8F98;">GPU</span>
        </div>
        <div style="display:flex;align-items:baseline;gap:6px;margin-bottom:8px;">
          <span style="font-size:32px;font-weight:700;color:#EDEDEF;">{{ stats.gpu_available ? stats.gpu_usage.toFixed(1) : 'N/A' }}</span>
          <span v-if="stats.gpu_available" style="font-size:13px;color:#8A8F98;">%</span>
        </div>
        <div v-if="stats.gpu_available" class="progress-track">
          <div class="progress-fill" style="background:rgba(34,197,94,0.8);" :style="{width: stats.gpu_usage + '%'}" />
        </div>
        <div style="font-size:12px;color:#8A8F98;margin-top:6px;">{{ stats.gpu_available ? stats.gpu_vendor : '无可用 GPU' }}</div>
      </div>

      <div class="glass" style="padding:20px;">
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;">
          <div class="stat-icon" style="background:rgba(245,158,11,0.12);">
            <PhPlay :size="18" color="#F59E0B" />
          </div>
          <span style="font-size:13px;color:#8A8F98;">运行中</span>
        </div>
        <div style="display:flex;align-items:baseline;gap:6px;margin-bottom:8px;">
          <span style="font-size:32px;font-weight:700;color:#EDEDEF;">{{ stats.running }}</span>
          <span style="font-size:13px;color:#8A8F98;">/ {{ stats.total }} 个任务</span>
        </div>
      </div>

      <div class="glass" style="padding:20px;">
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;">
          <div class="stat-icon" style="background:rgba(168,85,247,0.12);">
            <PhMemory :size="18" color="#A855F7" />
          </div>
          <span style="font-size:13px;color:#8A8F98;">内存</span>
        </div>
        <div style="display:flex;align-items:baseline;gap:6px;margin-bottom:8px;">
          <span style="font-size:32px;font-weight:700;color:#EDEDEF;">{{ stats.mem_mb }}</span>
          <span style="font-size:13px;color:#8A8F98;">MB</span>
        </div>
        <div style="font-size:12px;color:#8A8F98;margin-top:6px;">{{ stats.goroutines }} 个协程</div>
      </div>
    </div>

    <div class="glass" style="padding:24px;">
      <h2 style="font-size:18px;font-weight:600;margin-bottom:16px;">最近任务</h2>
      <div v-if="recentTasks.length" style="display:flex;flex-direction:column;gap:8px;">
        <div v-for="t in recentTasks" :key="t.id" style="display:flex;align-items:center;gap:12px;padding:10px 14px;border-radius:12px;background:rgba(255,255,255,0.02);">
          <div :class="'badge badge-' + t.status">{{ statusLabel(t.status) }}</div>
          <span style="font-size:14px;font-weight:500;">{{ t.name }}</span>
          <span style="font-size:12px;color:#8A8F98;margin-left:auto;">{{ t.time }}</span>
        </div>
      </div>
      <div v-else style="text-align:center;padding:40px;color:#8A8F98;">暂无最近任务</div>
    </div>

    <WizardModal :visible="showWizard" @close="showWizard = false" />
  </div>
</template>
<script setup lang="ts">
import { PhCpu, PhMonitor, PhPlay, PhMemory } from '@phosphor-icons/vue'
import { ref, onMounted, onUnmounted } from 'vue'
import WizardModal from '../wizard/WizardModal.vue'
import { api } from '@/api'

const showWizard = ref(false)

const stats = ref({ cpu_usage: 0, gpu_available: false, gpu_vendor: '', gpu_usage: 0, running: 0, total: 0, mem_mb: 0, goroutines: 0 })
let timer: any

onMounted(async () => {
  await fetchStatus()
  timer = setInterval(fetchStatus, 5000)
  try {
    recentTasks.value = await api.tasks.list()
  } catch {}
})
onUnmounted(() => clearInterval(timer))

async function fetchStatus() {
  try {
    const data = await api.overview()
    stats.value = data
  } catch {}
}

const recentTasks = ref<any[]>([])

function statusLabel(s: string) {
  const m: Record<string, string> = { running: '运行中', completed: '已完成', failed: '失败', pending: '等待', paused: '已暂停' }
  return m[s] || s
}
</script>
