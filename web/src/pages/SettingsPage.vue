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
        <h2 style="font-size:16px;font-weight:600;margin-bottom:20px;display:flex;align-items:center;gap:8px;">
          <PhFloppyDisk size="18" color="#5E6AD2" /> 已保存的编码配置
        </h2>
        <div style="display:flex;flex-direction:column;gap:8px;">
          <div v-for="cfg in savedConfigs" :key="cfg.name" style="display:flex;align-items:center;justify-content:space-between;padding:14px 16px;border-radius:12px;background:rgba(255,255,255,0.02);">
            <div>
              <div style="font-size:14px;font-weight:500;">{{ cfg.name }}</div>
              <div style="font-size:12px;color:#8A8F98;margin-top:2px;">{{ cfg.encoder }} | CRF {{ cfg.crf }} | {{ cfg.preset }}</div>
            </div>
            <div style="display:flex;gap:4px;">
              <button class="btn-ghost" style="padding:6px 12px;font-size:12px;">加载</button>
              <button class="btn-ghost" style="padding:6px 12px;font-size:12px;color:#EF4444;">删除</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="saved" class="glass" style="position:fixed;bottom:24px;right:24px;padding:12px 20px;display:flex;align-items:center;gap:8px;z-index:100;animation:fadeIn 300ms ease;">
      <PhCheckCircle size="18" color="#22C55E" />
      <span style="font-size:14px;">设置已保存</span>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref } from 'vue'
import { PhGear, PhScroll, PhFloppyDisk, PhCheckCircle } from '@phosphor-icons/vue'

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

const savedConfigs = ref([
  { name: 'x265 HQ Anime', encoder: 'libx265', crf: 18, preset: 'slow' },
  { name: 'x264 High Quality', encoder: 'libx264', crf: 20, preset: 'medium' },
  { name: 'NVENC Fast', encoder: 'h264_nvenc', crf: 23, preset: 'fast' },
])
</script>
<style scoped>
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
