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

    <ConfigCreateModal :visible="showCreateConfig" :edit-config="editingConfig" @close="showCreateConfig = false; editingConfig = null" @saved="onConfigSaved" />
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { PhGear, PhScroll, PhFloppyDisk, PhCheckCircle, PhMonitor } from '@phosphor-icons/vue'
import ConfigCreateModal from '../components/ConfigCreateModal.vue'

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
  { id: 0, name: 'x265 HQ Anime', encoder: 'x265', mode: 'cpu', isPreset: true,
    audio: { codec: 'FLAC', sampleRate: '48000Hz' },
    params: { crf: 15, preset: 'slower', deblock: '-1:-1', ctu: 32, 'qg-size': 8, me: 'star', subme: 5, merange: 38, bframes: 6, ref: 4, qcomp: 0.65, 'aq-mode': 1, 'aq-strength': 0.8, 'no-sao': true, 'psy-rd': 2.0, 'psy-rdoq': 1.0, 'rdoq-level': 2, rd: 5, pbratio: 1.2, cbqpoffs: -2, crqpoffs: -2, keyint: 360 } },
  { id: 1, name: 'x264 高画质 (mbtree on)', encoder: 'x264', mode: 'cpu', isPreset: true,
    audio: { codec: 'FLAC', sampleRate: '48000Hz' },
    params: { crf: 18, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 1, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11, merange: 48 } },
  { id: 2, name: 'x265 均衡', encoder: 'x265', mode: 'cpu', isPreset: true,
    audio: { codec: 'AAC', bitrate: '192kbps' },
    params: { crf: 20, preset: 'medium', 'aq-mode': 3, 'aq-strength': 0.7, deblock: '0:0', bframes: 8, 'rc-lookahead': 40 } },
  { id: 3, name: 'x264 高画质 (mbtree off)', encoder: 'x264', mode: 'cpu', isPreset: true,
    audio: { codec: 'FLAC', sampleRate: '48000Hz' },
    params: { crf: 16, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 0, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11 } },
  { id: 4, name: 'NVENC 快速编码', encoder: 'h264_nvenc', mode: 'gpu', isPreset: false,
    audio: { codec: 'OPUS', bitrate: '192kbps' },
    params: { crf: 20, preset: 'p7', tune: 'hq', rc: 'vbr', b_ref_mode: 'middle', multipass: 'qres', 'aq-strength': 8, lookahead: 32, bframes: 4 } },
  { id: 5, name: '我的自定义配置', encoder: 'x265', mode: 'cpu', isPreset: false,
    audio: { codec: 'FLAC', sampleRate: '48000Hz' },
    params: { crf: 17, preset: 'slow', 'aq-mode': 3, 'aq-strength': 0.9, deblock: '-1:-1', 'no-sao': true, bframes: 12 } },
])

const configDetail = ref<any>(null)
function viewConfigDetail(cfg: any) { configDetail.value = cfg }

const showCreateConfig = ref(false)

let nextId = 6
function onConfigSaved(config: any) {
  if (config.id !== undefined) {
    const idx = configs.value.findIndex(c => c.id === config.id)
    if (idx !== -1) configs.value[idx] = config
  } else {
    configs.value.push({ id: nextId++, ...config })
  }
}

const editingConfig = ref<any>(null)
function editConfig(cfg: any) {
  editingConfig.value = cfg
  showCreateConfig.value = true
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
  z-index: 59;
}
</style>
