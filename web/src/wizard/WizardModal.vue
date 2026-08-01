<template>
  <Teleport to="body">
    <div v-if="visible" class="wiz-overlay" @click.self="$emit('close')">
      <div class="wiz-panel glass" style="width:640px;max-height:90vh;overflow-y:auto;">
        <!-- Header -->
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
          <h2 style="font-size:20px;font-weight:700;">新建转码任务</h2>
          <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="$emit('close')">取消</button>
        </div>

        <!-- Steps indicator -->
        <div style="display:flex;gap:8px;margin-bottom:28px;">
          <div v-for="(s, i) in steps" :key="i"
            style="flex:1;text-align:center;cursor:pointer;"
            :style="{ cursor: i <= step ? 'pointer' : 'default' }"
            @click="i <= step && (step = i)">
            <div style="width:32px;height:32px;border-radius:50%;display:inline-flex;align-items:center;justify-content:center;font-size:13px;font-weight:600;margin-bottom:6px;"
              :style="{ background: i <= step ? '#5E6AD2' : 'rgba(255,255,255,0.06)', color: i <= step ? '#fff' : '#8A8F98' }">
              {{ i + 1 }}
            </div>
            <div style="font-size:12px;color:#8A8F98;">{{ s }}</div>
          </div>
        </div>

        <!-- Step 1: 源文件 -->
        <div v-if="step === 0">
          <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">选择蓝光光盘目录</label>
          <div style="display:flex;gap:8px;">
            <input type="text" v-model="sourcePath" placeholder="/input/BDROM" style="flex:1;" />
            <button class="btn-ghost" style="padding:8px 20px;white-space:nowrap;" @click="browseSource">浏览</button>
          </div>
          <div v-if="sourcePath" style="margin-top:8px;font-size:12px;color:#22C55E;">已选择: {{ sourcePath }}</div>
          <div style="margin-top:24px;display:flex;justify-content:flex-end;">
            <button class="btn-primary" :disabled="!sourcePath" @click="step = 1">下一步</button>
          </div>
        </div>

        <!-- Step 2: 文件选择 -->
        <div v-if="step === 1">
          <div style="font-size:13px;color:#8A8F98;margin-bottom:12px;">碟片名称: <span style="color:#EDEDEF;">{{ mockDiscName }}</span></div>
          <div style="font-size:13px;color:#8A8F98;margin-bottom:4px;">选择要转码的视频文件</div>
          <div style="display:flex;flex-direction:column;gap:4px;margin-bottom:20px;max-height:200px;overflow-y:auto;">
            <label v-for="f in mockFiles" :key="f.id"
              style="display:flex;align-items:center;gap:10px;padding:8px 12px;border-radius:8px;background:rgba(255,255,255,0.02);cursor:pointer;font-size:13px;">
              <input type="checkbox" v-model="selectedFiles" :value="f.id" style="accent-color:#5E6AD2;width:16px;height:16px;" />
              <span style="color:#EDEDEF;">{{ f.name }}</span>
              <span style="color:#8A8F98;margin-left:auto;">{{ f.size }}</span>
            </label>
          </div>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-bottom:20px;">
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">音频轨道</label>
              <select v-model="audioTrack">
                <option v-for="a in audioTracks" :key="a.id" :value="a.id">{{ a.label }}</option>
              </select>
            </div>
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">字幕轨道</label>
              <select v-model="subtitleTrack">
                <option v-for="s in subtitleTracks" :key="s.id" :value="s.id">{{ s.label }}</option>
              </select>
            </div>
          </div>
          <div style="display:flex;justify-content:space-between;">
            <button class="btn-ghost" @click="step = 0">上一步</button>
            <button class="btn-primary" :disabled="!selectedFiles.length" @click="step = 2">下一步</button>
          </div>
        </div>

        <!-- Step 3: 转码配置 -->
        <div v-if="step === 2">
          <div style="font-size:13px;color:#8A8F98;margin-bottom:12px;">选择编码配置</div>
          <div style="display:flex;flex-direction:column;gap:8px;margin-bottom:20px;">
            <label v-for="cfg in allConfigs" :key="cfg.id"
              style="display:flex;align-items:center;gap:12px;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);cursor:pointer;border:1px solid transparent;"
              :style="{ borderColor: selectedConfig === cfg.id ? '#5E6AD2' : 'transparent' }">
              <input type="radio" v-model="selectedConfig" :value="cfg.id" style="accent-color:#5E6AD2;width:16px;height:16px;" />
              <div style="flex:1;">
                <div style="font-size:14px;font-weight:500;">{{ cfg.name }}
                  <span v-if="cfg.isPreset" class="badge" style="background:rgba(94,106,210,0.15);color:#5E6AD2;font-size:10px;padding:2px 6px;margin-left:6px;">内置</span>
                </div>
                <div style="font-size:12px;color:#8A8F98;margin-top:2px;">{{ cfg.encoder }} | {{ cfg.mode === 'gpu' ? 'GPU' : 'CPU' }}</div>
              </div>
            </label>
            <label style="display:flex;align-items:center;gap:12px;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);cursor:pointer;border:1px dashed rgba(255,255,255,0.1);">
              <input type="radio" v-model="selectedConfig" :value="-1" style="accent-color:#5E6AD2;width:16px;height:16px;" />
              <span style="font-size:14px;color:#8A8F98;">+ 新建配置</span>
            </label>
          </div>
          <div style="display:flex;justify-content:space-between;">
            <button class="btn-ghost" @click="step = 1">上一步</button>
            <button class="btn-primary" :disabled="selectedConfig === null" @click="step = 3">下一步</button>
          </div>
        </div>

        <!-- Step 4: 目标路径 -->
        <div v-if="step === 3">
          <div>
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">输出目录</label>
            <input type="text" v-model="outputPath" placeholder="/output" />
          </div>
          <div style="margin-top:16px;">
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">输出文件名模板</label>
            <input type="text" v-model="outputNameTemplate" placeholder="{disc}_{track}.mkv" />
          </div>
          <div style="font-size:12px;color:#8A8F98;margin-top:4px;">可用变量: {'{disc}'}, {'{track}'}, {'{episode}'}</div>
          <div style="margin-top:24px;display:flex;justify-content:space-between;">
            <button class="btn-ghost" @click="step = 2">上一步</button>
            <div style="display:flex;gap:8px;">
              <button class="btn-ghost" @click="previewTranscode">预览转码</button>
              <button class="btn-primary" @click="complete">完成</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ close: [] }>()

const steps = ['源文件', '文件选择', '转码配置', '目标路径']
const step = ref(0)
const sourcePath = ref('')
const selectedFiles = ref<number[]>([])
const selectedConfig = ref<number | null>(null)
const outputPath = ref('/output')
const outputNameTemplate = ref('{disc}_{track}.mkv')
const audioTrack = ref(0)
const subtitleTrack = ref(0)

const mockDiscName = 'AMAZING_ANIME_BD_1'
const mockFiles = [
  { id: 0, name: 'BDMV/STREAM/00000.m2ts', size: '24.2 GB' },
  { id: 1, name: 'BDMV/STREAM/00001.m2ts', size: '19.8 GB' },
  { id: 2, name: 'BDMV/STREAM/00002.m2ts', size: '3.1 GB' },
  { id: 3, name: 'BDMV/STREAM/00003.m2ts', size: '6.4 GB' },
]

const audioTracks = [
  { id: 0, label: '日语 (FLAC 2.0)' },
  { id: 1, label: '日语 (FLAC 5.1)' },
  { id: 2, label: '评论音轨 (AC3 2.0)' },
]

const subtitleTracks = [
  { id: 0, label: '简体中文 (PGS)' },
  { id: 1, label: '繁体中文 (PGS)' },
  { id: 2, label: '无字幕' },
]

const allConfigs = [
  { id: 0, name: 'x265 HQ Anime', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 15, preset: 'slower', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', 'no-sao': true, 'bframes': 16, 'rc-lookahead': 60, 'subme': 7, 'merange': 57 } },
  { id: 1, name: 'x264 高画质 (mbtree on)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 18, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', 'mbtree': 1, 'rc-lookahead': 250, ref: 16, 'bframes': 16, 'subme': 11, 'merange': 48 } },
  { id: 2, name: 'x265 均衡', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 20, preset: 'medium', 'aq-mode': 3, 'aq-strength': 0.7, deblock: '0:0', 'bframes': 8, 'rc-lookahead': 40 } },
  { id: 3, name: 'x264 高画质 (mbtree off)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 16, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', 'mbtree': 0, 'rc-lookahead': 250, ref: 16, 'bframes': 16, 'subme': 11 } },
  { id: 4, name: 'NVENC 快速编码', encoder: 'h264_nvenc', mode: 'gpu', isPreset: false, params: { crf: 20, preset: 'p7', tune: 'hq', 'rc': 'vbr', 'b_ref_mode': 'middle', 'multipass': 'qres', 'aq-strength': 8, 'lookahead': 32, 'bframes': 4 } },
  { id: 5, name: '我的自定义配置', encoder: 'x265', mode: 'cpu', isPreset: false, params: { crf: 17, preset: 'slow', 'aq-mode': 3, 'aq-strength': 0.9, deblock: '-1:-1', 'no-sao': true, 'bframes': 12 } },
]

function browseSource() { sourcePath.value = '/input/BDROM/' + mockDiscName }

function previewTranscode() {
  alert('预览转码命令: ffmpeg -i ... -c:v libx265 -preset slow ...')
}

function complete() {
  alert('任务已创建！（Mock）')
  step.value = 0
  sourcePath.value = ''
  selectedFiles.value = []
  selectedConfig.value = null
  emit('close')
}
</script>

<style scoped>
.wiz-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.7);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex; align-items: center; justify-content: center;
  z-index: 100;
}
.wiz-panel {
  padding: 32px;
  animation: wizIn 250ms cubic-bezier(0.16,1,0.3,1);
}
@keyframes wizIn {
  from { opacity: 0; transform: scale(0.95) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
</style>
