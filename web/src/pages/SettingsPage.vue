<template>
  <div>
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:32px;">
      <div>
        <h1 style="font-size:28px;font-weight:700;">设置</h1>
        <p style="color:#8A8F98;margin-top:4px;font-size:14px;">系统配置与编码预设</p>
      </div>
      <button class="btn-primary" @click="saveSettings">保存设置</button>
    </div>

    <div style="display:flex;flex-direction:column;gap:24px;">
      <!-- 系统设置 -->
      <div class="glass" style="padding:24px;">
        <h2 style="font-size:16px;font-weight:600;margin-bottom:20px;display:flex;align-items:center;gap:8px;">
          <PhGear size="18" color="#5E6AD2" /> 系统设置
        </h2>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:20px;">
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">最大并发任务</label>
            <input type="number" v-model.number="settings.maxConcurrent" min="1" max="16" />
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">预览缓存时长（分钟）</label>
            <input type="number" v-model.number="settings.previewTTL" min="1" max="1440" />
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">临时目录</label>
            <input type="text" v-model="settings.tempDir" />
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">输出目录</label>
            <input type="text" v-model="settings.outputDir" />
          </div>
        </div>
        <div style="margin-top:20px;display:flex;align-items:center;justify-content:space-between;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);">
          <div>
            <div style="font-size:14px;font-weight:500;">启用 GPU 加速</div>
            <div style="font-size:12px;color:#8A8F98;margin-top:2px;">使用 NVENC / AMF 硬件编码</div>
          </div>
          <div class="toggle-track" :class="{ on: settings.gpuEnabled }" @click="settings.gpuEnabled = !settings.gpuEnabled">
            <div class="toggle-knob" />
          </div>
        </div>
      </div>

      <!-- GPU 检测 -->
      <div class="glass" style="padding:24px;">
        <div style="display:flex;align-items:center;justify-content:space-between;">
          <div>
            <h2 style="font-size:16px;font-weight:600;display:flex;align-items:center;gap:8px;">
              <PhMonitor size="18" color="#5E6AD2" /> GPU 硬件加速
            </h2>
            <div style="font-size:12px;color:#8A8F98;margin-top:2px;">检测可用的硬件编码器</div>
          </div>
          <button class="btn-ghost" @click="detectGPU" :disabled="detecting">
            {{ detecting ? '检测中...' : '检测 GPU' }}
          </button>
        </div>
        <div v-if="gpuResult" class="glass" style="margin-top:12px;padding:16px;border:1px solid rgba(255,255,255,0.06);">
          <div style="font-weight:600;margin-bottom:8px;font-size:15px;">{{ gpuResult.vendor }}</div>
          <div style="display:flex;gap:8px;flex-wrap:wrap;">
            <span v-for="enc in gpuResult.encoders" :key="enc.name" class="badge" :class="enc.available ? 'badge-completed' : 'badge-failed'">
              {{ enc.name }}
            </span>
          </div>
        </div>
      </div>

      <!-- 日志配置 -->
      <div class="glass" style="padding:24px;">
        <h2 style="font-size:16px;font-weight:600;margin-bottom:20px;display:flex;align-items:center;gap:8px;">
          <PhScroll size="18" color="#5E6AD2" /> 日志配置
        </h2>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:20px;">
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">日志级别</label>
            <select v-model="settings.logLevel">
              <option value="debug">DEBUG</option>
              <option value="info">INFO</option>
              <option value="warn">WARN</option>
              <option value="error">ERROR</option>
            </select>
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">最大文件数</label>
            <input type="number" v-model.number="settings.maxLogFiles" min="1" max="100" />
          </div>
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">单文件最大大小（MB）</label>
            <input type="number" v-model.number="settings.maxLogSize" min="1" max="1024" />
          </div>
        </div>
      </div>

      <!-- 已保存的配置 -->
      <div class="glass" style="padding:24px;">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
          <h2 style="font-size:16px;font-weight:600;display:flex;align-items:center;gap:8px;">
            <PhFloppyDisk size="18" color="#5E6AD2" /> 已保存的编码配置
          </h2>
          <button class="btn-primary" style="padding:8px 18px;font-size:13px;" @click="showCreateConfig = true">+ 新建配置</button>
        </div>
        <div style="display:flex;flex-direction:column;gap:8px;">
          <div v-for="cfg in configs" :key="cfg.id" style="display:flex;align-items:center;justify-content:space-between;padding:14px 16px;border-radius:12px;background:rgba(255,255,255,0.02);">
            <div>
              <div style="display:flex;align-items:center;gap:8px;">
                <span style="font-size:14px;font-weight:500;">{{ cfg.name }}</span>
                <span v-if="cfg.isPreset" class="badge" style="background:rgba(94,106,210,0.15);color:#5E6AD2;font-size:11px;padding:2px 8px;">内置</span>
              </div>
              <div style="font-size:12px;color:#8A8F98;margin-top:2px;">{{ cfg.encoder }} | {{ cfg.mode === 'gpu' ? 'GPU' : 'CPU' }}</div>
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

    <!-- Save toast -->
    <div v-if="saved" class="glass" style="position:fixed;bottom:24px;right:24px;padding:12px 20px;display:flex;align-items:center;gap:8px;z-index:100;animation:fadeIn 300ms ease;">
      <PhCheckCircle size="18" color="#22C55E" />
      <span style="font-size:14px;">设置已保存</span>
    </div>

    <!-- Config detail modal -->
    <Teleport to="body">
      <div v-if="configDetail" class="set-modal-overlay" @click.self="configDetail = null">
        <div class="glass" style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:50;width:560px;max-height:80vh;overflow:auto;padding:24px;">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
            <h3 style="font-size:18px;font-weight:700;">{{ configDetail.name }}</h3>
            <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="configDetail = null">关闭</button>
          </div>
          <div style="font-size:13px;color:#8A8F98;margin-bottom:4px;">{{ configDetail.encoder }} | {{ configDetail.mode === 'gpu' ? 'GPU 加速' : 'CPU 编码' }}</div>
          <div v-if="configDetail.isPreset" style="font-size:11px;color:#5E6AD2;margin-bottom:16px;">内置预设配置</div>
          <div style="display:flex;flex-direction:column;gap:6px;">
            <div v-for="(v, k) in configDetail.params" :key="k" style="display:flex;justify-content:space-between;padding:8px 12px;border-radius:8px;background:rgba(255,255,255,0.02);">
              <span style="color:#8A8F98;font-family:'JetBrains Mono',monospace;font-size:13px;">{{ k }}</span>
              <span style="color:#EDEDEF;font-weight:500;font-family:'JetBrains Mono',monospace;font-size:13px;">{{ v }}</span>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create config modal -->
    <Teleport to="body">
      <div v-if="showCreateConfig" class="set-modal-overlay" @click.self="showCreateConfig = false">
        <div class="glass" style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:50;width:480px;max-height:80vh;overflow:auto;padding:24px;">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
            <h3 style="font-size:18px;font-weight:700;">新建编码配置</h3>
            <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="showCreateConfig = false">取消</button>
          </div>
          <div style="display:flex;flex-direction:column;gap:14px;">
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">配置名称</label>
              <input type="text" v-model="newConfig.name" placeholder="我的配置" />
            </div>
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">编码器</label>
              <select v-model="newConfig.encoder">
                <option value="x265">x265</option>
                <option value="x264">x264</option>
                <option value="h264_nvenc">h264_nvenc</option>
                <option value="hevc_nvenc">hevc_nvenc</option>
              </select>
            </div>
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">编码模式</label>
              <select v-model="newConfig.mode">
                <option value="cpu">CPU</option>
                <option value="gpu">GPU</option>
              </select>
            </div>
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">CRF</label>
              <input type="number" v-model.number="newConfig.crf" min="0" max="51" />
            </div>
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">Preset</label>
              <select v-model="newConfig.preset">
                <option value="ultrafast">ultrafast</option>
                <option value="superfast">superfast</option>
                <option value="veryfast">veryfast</option>
                <option value="faster">faster</option>
                <option value="fast">fast</option>
                <option value="medium">medium</option>
                <option value="slow">slow</option>
                <option value="slower">slower</option>
                <option value="veryslow">veryslow</option>
              </select>
            </div>
          </div>
          <div style="margin-top:20px;display:flex;justify-content:flex-end;">
            <button class="btn-primary" @click="createConfig">创建配置</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Edit config modal -->
    <Teleport to="body">
      <div v-if="editingConfig" class="set-modal-overlay" @click.self="editingConfig = null">
        <div class="glass" style="position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);z-index:50;width:480px;max-height:80vh;overflow:auto;padding:24px;">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
            <h3 style="font-size:18px;font-weight:700;">编辑配置</h3>
            <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="editingConfig = null">取消</button>
          </div>
          <div style="display:flex;flex-direction:column;gap:14px;">
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">配置名称</label>
              <input type="text" v-model="editingConfig.name" />
            </div>
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">编码器</label>
              <select v-model="editingConfig.encoder">
                <option value="x265">x265</option>
                <option value="x264">x264</option>
                <option value="h264_nvenc">h264_nvenc</option>
                <option value="hevc_nvenc">hevc_nvenc</option>
              </select>
            </div>
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">编码模式</label>
              <select v-model="editingConfig.mode">
                <option value="cpu">CPU</option>
                <option value="gpu">GPU</option>
              </select>
            </div>
          </div>
          <div style="margin-top:20px;display:flex;justify-content:flex-end;">
            <button class="btn-primary" @click="saveEditedConfig">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { PhGear, PhScroll, PhFloppyDisk, PhCheckCircle, PhMonitor } from '@phosphor-icons/vue'

const settings = ref({
  maxConcurrent: 3,
  previewTTL: 30,
  tempDir: '/tmp/bdriper',
  outputDir: '/output',
  gpuEnabled: false,
  logLevel: 'info',
  maxLogFiles: 10,
  maxLogSize: 100,
})

const saved = ref(false)

function saveSettings() {
  saved.value = true
  setTimeout(() => { saved.value = false }, 2000)
}

const configs = ref([
  { id: 0, name: 'x265 HQ Anime', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 15, preset: 'slower', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', 'no-sao': true, bframes: 16, 'rc-lookahead': 60, subme: 7, merange: 57 } },
  { id: 1, name: 'x264 高画质 (mbtree on)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 18, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 1, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11, merange: 48 } },
  { id: 2, name: 'x265 均衡', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 20, preset: 'medium', 'aq-mode': 3, 'aq-strength': 0.7, deblock: '0:0', bframes: 8, 'rc-lookahead': 40 } },
  { id: 3, name: 'x264 高画质 (mbtree off)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 16, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 0, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11 } },
  { id: 4, name: 'NVENC 快速编码', encoder: 'h264_nvenc', mode: 'gpu', isPreset: false, params: { crf: 20, preset: 'p7', tune: 'hq', rc: 'vbr', b_ref_mode: 'middle', multipass: 'qres', 'aq-strength': 8, lookahead: 32, bframes: 4 } },
  { id: 5, name: '我的自定义配置', encoder: 'x265', mode: 'cpu', isPreset: false, params: { crf: 17, preset: 'slow', 'aq-mode': 3, 'aq-strength': 0.9, deblock: '-1:-1', 'no-sao': true, bframes: 12 } },
])

const configDetail = ref<any>(null)
function viewConfigDetail(cfg: any) { configDetail.value = cfg }

const showCreateConfig = ref(false)
const newConfig = ref({ name: '', encoder: 'x265', mode: 'cpu', crf: 18, preset: 'medium' })
let nextId = 6
function createConfig() {
  const name = newConfig.value.name || '未命名配置'
  configs.value.push({
    id: nextId++,
    name,
    encoder: newConfig.value.encoder,
    mode: newConfig.value.mode,
    isPreset: false,
    params: { crf: newConfig.value.crf, preset: newConfig.value.preset },
  })
  showCreateConfig.value = false
  newConfig.value = { name: '', encoder: 'x265', mode: 'cpu', crf: 18, preset: 'medium' }
}

const editingConfig = ref<any>(null)
function editConfig(cfg: any) {
  editingConfig.value = { ...cfg }
}
function saveEditedConfig() {
  const idx = configs.value.findIndex(c => c.id === editingConfig.value.id)
  if (idx !== -1) {
    configs.value[idx] = editingConfig.value
  }
  editingConfig.value = null
}

function deleteConfig(cfg: any) {
  configs.value = configs.value.filter(c => c.id !== cfg.id)
}

const gpuResult = ref<any>(null)
const detecting = ref(false)
async function detectGPU() {
  detecting.value = true
  await new Promise(r => setTimeout(r, 1500))
  gpuResult.value = {
    vendor: 'NVIDIA GeForce RTX 4060',
    encoders: [
      { name: 'h264_nvenc', available: true },
      { name: 'hevc_nvenc', available: true },
      { name: 'av1_nvenc', available: false },
    ]
  }
  detecting.value = false
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
  z-index: 49;
}
</style>
