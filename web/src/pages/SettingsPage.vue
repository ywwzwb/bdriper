<template>
  <div class="max-w-2xl">
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-fg">系统设置</h1>
      <p class="text-muted mt-1">配置转码引擎与日志参数</p>
    </div>

    <transition name="fade">
      <div v-if="error" class="bg-destructive/10 border border-destructive/30 text-destructive rounded-lg px-4 py-3 mb-4 text-sm">{{ error }}</div>
    </transition>
    <transition name="fade">
      <div v-if="saved" class="bg-accent/10 border border-accent/30 text-accent rounded-lg px-4 py-3 mb-4 text-sm flex items-center gap-2">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>
        设置已保存
      </div>
    </transition>

    <div class="space-y-4">
      <!-- Transcoding -->
      <div class="bg-card border border-border/50 rounded-xl overflow-hidden">
        <div class="px-5 py-4 border-b border-border/30">
          <h2 class="font-semibold text-fg">转码设置</h2>
        </div>
        <div class="p-5 space-y-4">
          <div>
            <label class="block text-sm font-medium text-fg mb-2">最大同时任务数</label>
            <input type="number" v-model="form.max_concurrent" min="1" max="64" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200" />
          </div>
          <div>
            <label class="block text-sm font-medium text-fg mb-2">预览缓存时间</label>
            <select v-model="form.preview_ttl" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200">
              <option value="15m">15 分钟</option>
              <option value="30m">30 分钟</option>
              <option value="1h">1 小时</option>
              <option value="2h">2 小时</option>
            </select>
          </div>
        </div>
      </div>

      <!-- GPU -->
      <div class="bg-card border border-border/50 rounded-xl overflow-hidden">
        <div class="px-5 py-4 border-b border-border/30">
          <h2 class="font-semibold text-fg">GPU 加速</h2>
        </div>
        <div class="p-5 space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <span class="text-sm font-medium text-fg">启用 GPU 编码</span>
              <p class="text-xs text-muted mt-0.5">使用硬件加速编码器</p>
            </div>
            <button @click="form.gpu_enabled = !form.gpu_enabled" class="relative w-11 h-6 rounded-full transition-colors duration-200" :class="form.gpu_enabled ? 'bg-accent' : 'bg-muted'" role="switch">
              <span class="absolute top-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform duration-200" :class="form.gpu_enabled ? 'translate-x-5.5 left-0.5' : 'left-0.5'" />
            </button>
          </div>

          <div class="flex items-center gap-3">
            <button @click="testGpu" class="px-4 py-2 text-sm border border-border/50 rounded-lg hover:bg-muted transition-all duration-200 flex items-center gap-2" :disabled="testingGpu">
              <svg v-if="testingGpu" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
              <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
              {{ testingGpu ? '检测中...' : '检测 GPU' }}
            </button>
          </div>

          <div v-if="gpuCards.length" class="space-y-2">
            <div v-for="g in gpuCards" :key="g.model" class="bg-muted rounded-lg p-3 flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg flex items-center justify-center" :class="g.vendor === 'NVIDIA' ? 'bg-green-500/10' : g.vendor === 'Intel' ? 'bg-blue-500/10' : 'bg-red-500/10'">
                <span class="text-xs font-bold" :class="g.vendor === 'NVIDIA' ? 'text-green-400' : g.vendor === 'Intel' ? 'text-blue-400' : 'text-red-400'">{{ g.vendor[0] }}</span>
              </div>
              <div class="flex-1">
                <span class="text-sm font-medium">{{ g.model }}</span>
                <div class="flex gap-2 mt-0.5">
                  <span v-for="enc in g.encoders" :key="enc.name" class="text-xs" :class="enc.supported ? 'text-accent' : 'text-muted'">{{ enc.codec }}</span>
                </div>
              </div>
            </div>
          </div>
          <div v-if="gpuError" class="text-destructive text-sm">{{ gpuError }}</div>
        </div>
      </div>

      <!-- Logging -->
      <div class="bg-card border border-border/50 rounded-xl overflow-hidden">
        <div class="px-5 py-4 border-b border-border/30">
          <h2 class="font-semibold text-fg">日志设置</h2>
        </div>
        <div class="p-5 space-y-4">
          <div>
            <label class="block text-sm font-medium text-fg mb-2">日志级别</label>
            <select v-model="form.log_level" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200">
              <option value="debug">调试</option>
              <option value="info">信息</option>
              <option value="warn">警告</option>
              <option value="error">错误</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-fg mb-2">最大日志文件数</label>
              <input type="number" v-model="form.log_max_files" min="1" max="100" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200" />
            </div>
            <div>
              <label class="block text-sm font-medium text-fg mb-2">单文件最大 (MB)</label>
              <input type="number" v-model="form.log_max_size" min="1" max="1024" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200" />
            </div>
          </div>
        </div>
      </div>

      <button @click="save" class="px-6 py-2.5 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 shadow-lg shadow-accent/20" :disabled="saving">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <!-- Saved Configs -->
    <div class="mt-12">
      <div class="mb-4">
        <h2 class="text-xl font-bold text-fg">已保存的编码配置</h2>
        <p class="text-muted text-sm mt-1">可复用的转码预设方案</p>
      </div>
      <div v-if="configs.length === 0" class="bg-card border border-border/50 rounded-xl p-8 text-center text-muted text-sm">暂无保存的配置</div>
      <div v-else class="space-y-2">
        <div v-for="c in configs" :key="c.id" class="bg-card border border-border/50 rounded-xl p-4 flex items-center justify-between hover:border-accent/20 transition-all duration-200">
          <div>
            <div class="font-medium text-fg">{{ c.name }}</div>
            <div class="text-xs text-muted mt-0.5">{{ c.encoder }} / {{ c.mode || '简易' }}</div>
          </div>
          <div class="flex gap-2">
            <button @click="viewConfig(c)" class="px-3 py-1.5 text-sm border border-border/50 rounded-lg hover:bg-muted transition-all duration-200">查看</button>
            <button @click="deleteConfig(c.id)" class="px-3 py-1.5 text-sm border border-destructive/30 text-destructive rounded-lg hover:bg-destructive/10 transition-all duration-200">删除</button>
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
  saving.value = true; error.value = ''; saved.value = false
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
  testingGpu.value = true; gpuError.value = ''; gpuCards.value = []
  try { gpuCards.value = await api.settings.gpuInfo() } catch (e: any) { gpuError.value = e.message || 'GPU 检测失败' } finally { testingGpu.value = false }
}

function viewConfig(config: any) { editingConfig.value = config }
function onConfigSaved() { editingConfig.value = null; api.configs.list().then(c => configs.value = c).catch(() => {}) }
async function deleteConfig(id: number) {
  try { await api.configs.delete(id); configs.value = configs.value.filter(c => c.id !== id) } catch (e: any) { error.value = e.message || '删除失败' }
}
</script>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
