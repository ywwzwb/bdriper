<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:32px;">
      <div>
        <h1 style="font-size:28px;font-weight:700;color:#EDEDEF;">概览</h1>
        <p style="color:#8A8F98;margin-top:4px;font-size:14px;">系统资源与任务状态</p>
      </div>
      <router-link to="/tasks?wizard=open" class="btn-primary" style="text-decoration:none;display:inline-flex;align-items:center;gap:6px;">
        <span style="font-size:18px;">+</span> 新建任务
      </router-link>
    </div>

    <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:16px;margin-bottom:32px;">
      <div v-for="s in stats" :key="s.label" class="glass" style="padding:20px;">
        <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;">
          <div class="stat-icon" :style="{background: s.bg}">
            <component :is="s.icon" :size="18" :color="s.color" />
          </div>
          <span style="font-size:13px;color:#8A8F98;">{{ s.label }}</span>
        </div>
        <div style="display:flex;align-items:baseline;gap:6px;margin-bottom:8px;">
          <span style="font-size:32px;font-weight:700;color:#EDEDEF;">{{ s.value }}</span>
          <span v-if="s.unit" style="font-size:13px;color:#8A8F98;">{{ s.unit }}</span>
        </div>
        <div v-if="s.progress !== null" class="progress-track">
          <div class="progress-fill" :style="{width: s.progress + '%'}" />
        </div>
        <div v-if="s.detail" style="font-size:12px;color:#8A8F98;margin-top:6px;">{{ s.detail }}</div>
      </div>
    </div>

    <div class="glass" style="padding:24px;">
      <h2 style="font-size:18px;font-weight:600;margin-bottom:16px;">最近任务</h2>
      <div v-if="recentTasks.length" style="display:flex;flex-direction:column;gap:8px;">
        <div v-for="t in recentTasks" :key="t.id" style="display:flex;align-items:center;gap:12px;padding:10px 14px;border-radius:12px;background:rgba(255,255,255,0.02);">
          <div :class="'badge badge-' + t.status">{{ t.statusLabel }}</div>
          <span style="font-size:14px;font-weight:500;">{{ t.name }}</span>
          <span style="font-size:12px;color:#8A8F98;margin-left:auto;">{{ t.time }}</span>
        </div>
      </div>
      <div v-else style="text-align:center;padding:40px;color:#8A8F98;">暂无最近任务</div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { PhCpu, PhMonitor, PhPlay, PhMemory } from '@phosphor-icons/vue'
import { ref } from 'vue'

const stats = ref([
  { label: 'CPU 使用率', value: 13, unit: '%', progress: 13, bg: 'rgba(94,106,210,0.12)', color: '#5E6AD2', icon: PhCpu, detail: null },
  { label: 'GPU', value: 'N/A', unit: '', progress: null, bg: 'rgba(34,197,94,0.12)', color: '#22C55E', icon: PhMonitor, detail: '无可用 GPU' },
  { label: '运行中', value: 2, unit: '/ 5 个任务', progress: null, bg: 'rgba(245,158,11,0.12)', color: '#F59E0B', icon: PhPlay, detail: null },
  { label: '内存', value: 24, unit: 'MB', progress: null, bg: 'rgba(168,85,247,0.12)', color: '#A855F7', icon: PhMemory, detail: '12 个协程' },
])

const recentTasks = ref([
  { id: 1, name: '四月は君の嘘 EP01', status: 'running', statusLabel: '运行中', time: '2 分钟前' },
  { id: 2, name: 'Charlotte EP01', status: 'completed', statusLabel: '已完成', time: '15 分钟前' },
  { id: 3, name: 'Kill la Kill EP01', status: 'failed', statusLabel: '失败', time: '32 分钟前' },
])
</script>
