<template>
  <div class="max-w-2xl">
    <h1 class="text-2xl font-semibold mb-6">系统设置</h1>

    <div v-if="error" class="text-destructive mb-4">{{ error }}</div>
    <div v-if="saved" class="text-accent mb-4">设置已保存</div>

    <div class="space-y-6">
      <div class="bg-card border border-border rounded-lg p-4">
        <label class="block text-sm text-muted mb-1">最大同时任务数</label>
        <input type="number" v-model="form.max_concurrent" min="1" max="64" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg focus:outline-none focus:border-accent" />
      </div>

      <div class="bg-card border border-border rounded-lg p-4">
        <label class="block text-sm text-muted mb-1">预览缓存时间</label>
        <select v-model="form.preview_ttl" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg focus:outline-none focus:border-accent">
          <option value="15m">15 分钟</option>
          <option value="30m">30 分钟</option>
          <option value="1h">1 小时</option>
          <option value="2h">2 小时</option>
        </select>
      </div>

      <div class="bg-card border border-border rounded-lg p-4">
        <div class="flex items-center justify-between mb-3">
          <label class="text-sm text-muted">GPU 加速</label>
          <div class="flex items-center gap-3">
            <button @click="form.gpu_enabled = !form.gpu_enabled" class="relative w-10 h-5 rounded-full transition-colors" :class="form.gpu_enabled ? 'bg-accent' : 'bg-muted'" role="switch">
              <span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform" :class="form.gpu_enabled ? 'translate-x-5' : ''" />
            </button>
            <button @click="testGpu" class="px-3 py-1 text-xs border border-border rounded hover:bg-muted transition" :disabled="testingGpu">{{ testingGpu ? '检测中...' : '检测 GPU' }}</button>
          </div>
        </div>
        <div v-if="gpuCards.length" class="mt-3 space-y-2">
          <div v-for="g in gpuCards" :key="g.name" class="bg-muted rounded-lg p-3 flex items-center justify-between">
            <span class="text-sm font-medium">{{ g.name }}</span>
            <span class="text-xs text-muted">{{ g.memory_mb }} MB</span>
          </div>
        </div>
        <div v-if="gpuError" class="text-destructive text-sm mt-2">{{ gpuError }}</div>
      </div>

      <div class="bg-card border border-border rounded-lg p-4">
        <label class="block text-sm text-muted mb-1">日志级别</label>
        <select v-model="form.log_level" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg focus:outline-none focus:border-accent">
          <option value="debug">调试</option>
          <option value="info">信息</option>
          <option value="warn">警告</option>
          <option value="error">错误</option>
        </select>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="bg-card border border-border rounded-lg p-4">
          <label class="block text-sm text-muted mb-1">最大日志文件数</label>
          <input type="number" v-model="form.log_max_files" min="1" max="100" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg focus:outline-none focus:border-accent" />
        </div>
        <div class="bg-card border border-border rounded-lg p-4">
          <label class="block text-sm text-muted mb-1">单日志最大大小 (MB)</label>
          <input type="number" v-model="form.log_max_size" min="1" max="1024" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg focus:outline-none focus:border-accent" />
        </div>
      </div>

      <button @click="save" class="px-6 py-2.5 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition" :disabled="saving">{{ saving ? '保存中...' : '保存设置' }}</button>
    </div>

    <div class="mt-10">
      <h2 class="text-xl font-semibold mb-4">已保存的配置</h2>
      <div v-if="configs.length === 0" class="text-muted text-sm">暂无保存的配置</div>
      <div class="space-y-2">
        <div v-for="c in configs" :key="c.id" class="bg-card border border-border rounded-lg p-3 flex items-center justify-between">
          <div>
            <div class="font-medium">{{ c.name }}</div>
            <div class="text-xs text-muted">{{ c.encoder }} / {{ c.mode || '简易' }}</div>
          </div>
          <div class="flex gap-2">
            <button @click="viewConfig(c)" class="px-2 py-1 text-xs border border-border rounded hover:bg-muted transition">查看</button>
            <button @click="deleteConfig(c.id)" class="px-2 py-1 text-xs border border-destructive text-destructive rounded hover:bg-destructive/10 transition">删除</button>
          </div>
        </div>
      </div>
    </div>

    <ConfigEditor v-if="editingConfig" :config="editingConfig" @close="editingConfig = null" @saved="onConfigSaved" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { api } from '@/api'
import ConfigEditor from '@/config/ConfigEditor.vue'

const form = reactive({
  max_concurrent: 4,
  preview_ttl: '1h',
  gpu_enabled: false,
  log_level: 'info',
  log_max_files: 10,
  log_max_size: 100,
})

const configs = ref<any[]>([])
const gpuCards = ref<any[]>([])
const testingGpu = ref(false)
const gpuError = ref('')
const error = ref('')
const saved = ref(false)
const saving = ref(false)
const editingConfig = ref<any>(null)

onMounted(async () => {
  try {
    const settings = await api.settings.list()
    if (settings) Object.assign(form, settings)
  } catch { /* ignore */ }

  try {
    configs.value = await api.configs.list()
  } catch { /* ignore */ }
})

async function save() {
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    await api.settings.update({ ...form })
    saved.value = true
    setTimeout(() => saved.value = false, 3000)
  } catch (e: any) {
    error.value = e.message || '保存失败'
  } finally {
    saving.value = false
  }
}

async function testGpu() {
  testingGpu.value = true
  gpuError.value = ''
  gpuCards.value = []
  try {
    gpuCards.value = await api.settings.gpuInfo()
  } catch (e: any) {
    gpuError.value = e.message || 'GPU 检测失败'
  } finally {
    testingGpu.value = false
  }
}

function viewConfig(config: any) {
  editingConfig.value = config
}

function onConfigSaved() {
  editingConfig.value = null
  api.configs.list().then(c => configs.value = c).catch(() => {})
}

async function deleteConfig(id: number) {
  try {
    await api.configs.delete(id)
    configs.value = configs.value.filter(c => c.id !== id)
  } catch (e: any) {
    error.value = e.message || '删除失败'
  }
}
</script>
