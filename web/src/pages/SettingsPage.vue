<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:32px;">
      <div>
        <h1 style="font-size:28px;font-weight:700;">设置</h1>
        <p style="color:#8A8F98;margin-top:4px;font-size:14px;">系统配置与编码预设</p>
      </div>
      <button class="btn-primary" @click="save">保存设置</button>
    </div>

    <div style="display:flex;flex-direction:column;gap:24px;">
      <div class="glass" style="padding:24px;">
        <h2 style="font-size:16px;font-weight:600;margin-bottom:20px;display:flex;align-items:center;gap:8px;">
          <PhGear size="18" color="#5E6AD2" /> 系统设置
        </h2>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:20px;">
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">最大并发任务</label>
            <input type="number" v-model.number="form.max_concurrent" min="1" max="16" />
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">预览缓存时长（分钟）</label>
            <input type="number" v-model.number="form.preview_ttl" min="1" max="1440" />
          </div>
        </div>
        <div style="margin-top:20px;display:flex;align-items:center;justify-content:space-between;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);">
          <div>
            <div style="font-size:14px;font-weight:500;">启用 GPU 加速</div>
            <div style="font-size:12px;color:#8A8F98;margin-top:2px;">使用 NVENC / AMF 硬件编码</div>
          </div>
          <div class="toggle-track" :class="{ on: form.gpu_enabled === 'true' }" @click="form.gpu_enabled = form.gpu_enabled === 'true' ? 'false' : 'true'">
            <div class="toggle-knob" />
          </div>
        </div>
      </div>

      <div class="glass" style="padding:24px;">
        <div style="display:flex;align-items:center;justify-content:space-between;">
          <div>
            <h2 style="font-size:16px;font-weight:600;display:flex;align-items:center;gap:8px;">
              <PhMonitor size="18" color="#5E6AD2" /> GPU 硬件加速
            </h2>
            <div style="font-size:12px;color:#8A8F98;margin-top:2px;">检测可用的硬件编码器</div>
          </div>
          <button class="btn-ghost" @click="testGpu" :disabled="testingGpu">
            {{ testingGpu ? '检测中...' : '检测 GPU' }}
          </button>
        </div>
        <div v-if="gpuCards.length" class="glass" style="margin-top:12px;padding:16px;border:1px solid rgba(255,255,255,0.06);">
          <div style="font-weight:600;margin-bottom:8px;font-size:15px;">{{ gpuCards[0].vendor || gpuCards[0].name }}</div>
          <div style="display:flex;gap:8px;flex-wrap:wrap;">
            <span v-for="enc in (gpuCards[0].encoders || [])" :key="enc.name" class="badge" :class="enc.available ? 'badge-completed' : 'badge-failed'">
              {{ enc.name }}
            </span>
          </div>
        </div>
        <div v-if="gpuError" style="margin-top:12px;color:#EF4444;font-size:13px;">{{ gpuError }}</div>
      </div>

      <div class="glass" style="padding:24px;">
        <h2 style="font-size:16px;font-weight:600;margin-bottom:20px;display:flex;align-items:center;gap:8px;">
          <PhScroll size="18" color="#5E6AD2" /> 日志配置
        </h2>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:20px;">
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">日志级别</label>
            <select v-model="form.log_level">
              <option value="debug">DEBUG</option>
              <option value="info">INFO</option>
              <option value="warn">WARN</option>
              <option value="error">ERROR</option>
            </select>
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">最大文件数</label>
            <input type="number" v-model.number="form.log_max_files" min="1" max="100" />
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">单文件最大大小（MB）</label>
            <input type="number" v-model.number="form.log_max_size" min="1" max="1024" />
          </div>
        </div>
      </div>

      <div class="glass" style="padding:24px;">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
          <h2 style="font-size:16px;font-weight:600;display:flex;align-items:center;gap:8px;">
            <PhFloppyDisk size="18" color="#5E6AD2" /> 已保存的编码配置
          </h2>
          <button class="btn-primary" style="padding:8px 18px;font-size:13px;" @click="showCreateConfig = true">+ 新建配置</button>
        </div>
        <div style="display:flex;flex-direction:column;gap:8px;">
          <div v-for="cfg in configs" :key="cfg.id || cfg.name" style="display:flex;align-items:center;justify-content:space-between;padding:14px 16px;border-radius:12px;background:rgba(255,255,255,0.02);">
            <div>
              <div style="display:flex;align-items:center;gap:8px;">
                <span style="font-size:14px;font-weight:500;">{{ cfg.name }}</span>
                <span v-if="cfg.isPreset" class="badge" style="background:rgba(94,106,210,0.15);color:#5E6AD2;font-size:11px;padding:2px 8px;">内置</span>
              </div>
              <div style="font-size:12px;color:#8A8F98;margin-top:2px;">{{ cfg.encoder || cfg.video_encoder }} | {{ cfg.mode === 'gpu' ? 'GPU' : 'CPU' }}</div>
            </div>
            <div style="display:flex;gap:4px;">
              <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="viewConfigDetail(cfg)">查看</button>
              <button v-if="!cfg.isPreset" class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="editConfig(cfg)">编辑</button>
              <button v-if="!cfg.isPreset" class="btn-ghost" style="padding:6px 12px;font-size:12px;color:#EF4444;" @click="deleteConfig(cfg)">删除</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="saved" class="glass" style="position:fixed;bottom:24px;right:24px;padding:12px 20px;display:flex;align-items:center;gap:8px;z-index:100;animation:fadeIn 300ms ease;">
      <PhCheckCircle size="18" color="#22C55E" />
      <span style="font-size:14px;">设置已保存</span>
    </div>

    <div v-if="error" class="glass" style="position:fixed;bottom:24px;right:24px;padding:12px 20px;display:flex;align-items:center;gap:8px;z-index:100;border:1px solid rgba(239,68,68,0.3);">
      <span style="font-size:14px;color:#EF4444;">{{ error }}</span>
    </div>

    <Teleport to="body">
      <div v-if="configDetail" class="set-modal-overlay" @click.self="configDetail = null" style="z-index:80;">
        <div class="glass" style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:80;width:600px;max-height:80vh;overflow-y:auto;">
          <div style="display:flex;align-items:center;justify-content:space-between;padding:16px 20px;border-bottom:1px solid rgba(255,255,255,0.06);">
            <div style="display:flex;align-items:center;gap:10px;">
              <h3 style="font-weight:600;font-size:16px;">{{ configDetail.name }}</h3>
              <span v-if="configDetail.isPreset" class="badge badge-pending">内置预设</span>
              <span class="badge" :class="configDetail.mode === 'gpu' ? 'badge-running' : 'badge-pending'">{{ configDetail.mode === 'gpu' ? 'GPU' : 'CPU' }}</span>
            </div>
            <button @click="configDetail = null" style="color:#8A8F98;font-size:20px;cursor:pointer;background:none;border:none;">✕</button>
          </div>
          <div style="padding:20px;">
            <div style="margin-bottom:16px;">
              <div style="color:#8A8F98;font-size:12px;">编码器</div>
              <div style="font-weight:500;font-size:15px;">{{ configDetail.encoder || configDetail.video_encoder }}</div>
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

    <ConfigCreateModal :visible="showCreateConfig" :edit-config="editingConfig" @close="showCreateConfig = false; editingConfig = null" @saved="onConfigSaved" />
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { PhGear, PhScroll, PhFloppyDisk, PhCheckCircle, PhMonitor } from '@phosphor-icons/vue'
import ConfigCreateModal from '../components/ConfigCreateModal.vue'
import { api } from '@/api'

const form = reactive({
  max_concurrent: '2',
  preview_ttl: '30m',
  gpu_enabled: 'true',
  log_level: 'info',
  log_max_files: '5',
  log_max_size: '10',
})

const configs = ref<any[]>([])
const saved = ref(false)
const error = ref('')
const saving = ref(false)

async function loadConfigs() {
  try {
    const [presets, userCfgs] = await Promise.all([api.presets(), api.configs.list()])
    const presetNames = new Set(presets.map((p: any) => p.name))
    const userOnly = (userCfgs || []).filter((c: any) => !presetNames.has(c.name))
    configs.value = [...presets.map((p: any) => ({ ...p, isPreset: true })), ...userOnly]
  } catch {}
}

onMounted(async () => {
  try {
    const settings = await api.settings.list()
    Object.assign(form, settings)
  } catch {}
  loadConfigs()
})

async function save() {
  saving.value = true
  try {
    await api.settings.update({ ...form })
    saved.value = true
    setTimeout(() => saved.value = false, 3000)
  } catch (e: any) {
    error.value = e.message
    setTimeout(() => error.value = '', 5000)
  } finally {
    saving.value = false
  }
}

const testingGpu = ref(false)
const gpuCards = ref<any[]>([])
const gpuError = ref('')

async function testGpu() {
  testingGpu.value = true
  gpuError.value = ''
  try {
    gpuCards.value = await api.settings.gpuInfo()
  } catch {
    gpuError.value = 'GPU 检测失败'
  } finally {
    testingGpu.value = false
  }
}

const configDetail = ref<any>(null)
function viewConfigDetail(cfg: any) { configDetail.value = cfg }

const showCreateConfig = ref(false)
const editingConfig = ref<any>(null)

async function onConfigSaved(config: any) {
  loadConfigs()
}

function editConfig(cfg: any) {
  editingConfig.value = cfg
  showCreateConfig.value = true
}

async function deleteConfig(cfg: any) {
  try {
    await api.configs.delete(cfg.id)
    configs.value = configs.value.filter(c => c.id !== cfg.id)
  } catch (e: any) {
    error.value = e.message
    setTimeout(() => error.value = '', 5000)
  }
}
</script>
<style scoped>
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
.set-modal-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  z-index: 59;
}
</style>
