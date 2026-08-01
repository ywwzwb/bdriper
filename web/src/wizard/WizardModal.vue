<template>
  <Teleport to="body">
    <div v-if="visible" class="wiz-overlay" @click.self="$emit('close')">
      <div class="wiz-panel glass" style="width:640px;max-height:90vh;overflow-y:auto;">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
          <h2 style="font-size:20px;font-weight:700;">新建转码任务</h2>
          <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="$emit('close')">取消</button>
        </div>

        <!-- Completion state -->
        <div v-if="wizardComplete" style="text-align:center;padding:60px 20px;">
          <div style="font-size:48px;margin-bottom:16px;">✓</div>
          <div style="font-size:20px;font-weight:600;margin-bottom:8px;">任务创建成功</div>
          <div style="color:#8A8F98;margin-bottom:24px;">{{ createdTaskName }} 已加入任务队列</div>
          <button class="btn-primary" @click="closeWizard">返回任务列表</button>
        </div>

        <template v-if="!wizardComplete">
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
              <button class="btn-ghost" style="padding:10px 18px;white-space:nowrap;" @click="openFilePicker('source')">浏览</button>
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

            <!-- Audio tracks — multi-select checkboxes -->
            <div style="margin-bottom:20px;">
              <div style="font-size:14px;color:#8A8F98;margin-bottom:8px;">音轨 (多选)</div>
              <div v-for="track in parsedAudioTracks" :key="track.id"
                   style="display:flex;align-items:center;gap:10px;padding:8px 12px;border-radius:8px;cursor:pointer;"
                   class="hover-span" @click="toggleTrack(track, 'audio')">
                <input type="checkbox" :checked="selectedAudio.includes(track.id)"
                       style="accent-color:#5E6AD2;width:16px;height:16px;" @click.stop />
                <span style="font-size:14px;">{{ track.label }}</span>
                <span style="font-size:12px;color:#8A8F98;">{{ track.lang }}</span>
              </div>
            </div>

            <!-- Subtitle tracks — multi-select checkboxes -->
            <div style="margin-bottom:20px;">
              <div style="font-size:14px;color:#8A8F98;margin-bottom:8px;">字幕 (多选)</div>
              <div v-for="track in parsedSubtitleTracks" :key="track.id"
                   style="display:flex;align-items:center;gap:10px;padding:8px 12px;border-radius:8px;cursor:pointer;"
                   class="hover-span" @click="toggleTrack(track, 'subtitle')">
                <input type="checkbox" :checked="selectedSubtitles.includes(track.id)"
                       style="accent-color:#5E6AD2;width:16px;height:16px;" @click.stop />
                <span style="font-size:14px;">{{ track.label }}</span>
                <span style="font-size:12px;color:#8A8F98;">{{ track.lang }}</span>
              </div>
            </div>

            <!-- Chapter toggle -->
            <div style="display:flex;align-items:center;justify-content:space-between;padding:8px 12px;border-radius:8px;">
              <span style="font-size:14px;">章节信息</span>
              <div class="toggle-track" :class="{ on: chaptersEnabled }" @click="chaptersEnabled = !chaptersEnabled" style="width:44px;height:24px;flex-shrink:0;">
                <div class="toggle-knob" style="width:20px;height:20px;" />
              </div>
            </div>

            <div style="display:flex;justify-content:space-between;margin-top:20px;">
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
                :style="{ borderColor: selectedConfig === cfg.id ? '#5E6AD2' : 'transparent' }"
                @click="selectedConfig = cfg.id">
                <input type="radio" :checked="selectedConfig === cfg.id" style="accent-color:#5E6AD2;width:16px;height:16px;" />
                <div style="flex:1;">
                  <div style="font-size:14px;font-weight:500;">{{ cfg.name }}
                    <span v-if="cfg.isPreset" class="badge" style="background:rgba(94,106,210,0.15);color:#5E6AD2;font-size:10px;padding:2px 6px;margin-left:6px;">内置</span>
                  </div>
                  <div style="font-size:12px;color:#8A8F98;margin-top:2px;">{{ cfg.encoder }} | {{ cfg.mode === 'gpu' ? 'GPU' : 'CPU' }}</div>
                </div>
              </label>
              <label style="display:flex;align-items:center;gap:12px;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);cursor:pointer;border:1px dashed rgba(255,255,255,0.1);"
                :style="{ borderColor: selectedConfig === -1 ? '#5E6AD2' : 'transparent' }"
                @click="selectedConfig = -1">
                <input type="radio" :checked="selectedConfig === -1" style="accent-color:#5E6AD2;width:16px;height:16px;" />
                <span style="font-size:14px;color:#8A8F98;">+ 新建配置</span>
              </label>
            </div>

            <!-- Inline config creation -->
            <div v-if="selectedConfig === -1" class="glass" style="padding:20px;margin-bottom:20px;">
              <div v-if="inlineConfigStep === 0" style="text-align:center;">
                <div style="font-size:14px;font-weight:600;margin-bottom:16px;">选择配置模式</div>
                <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
                  <div class="hover-span" @click="startSimpleConfig" style="padding:20px;cursor:pointer;text-align:center;border-radius:12px;background:rgba(255,255,255,0.02);border:1px solid rgba(255,255,255,0.06);">
                    <div style="font-size:28px;margin-bottom:6px;">⚡</div>
                    <div style="font-weight:600;margin-bottom:2px;font-size:13px;">简易模式</div>
                    <div style="font-size:11px;color:#8A8F98;">基础参数，快速配置</div>
                  </div>
                  <div class="hover-span" @click="startProConfig" style="padding:20px;cursor:pointer;text-align:center;border-radius:12px;background:rgba(255,255,255,0.02);border:1px solid rgba(255,255,255,0.06);">
                    <div style="font-size:28px;margin-bottom:6px;">🔧</div>
                    <div style="font-weight:600;margin-bottom:2px;font-size:13px;">专业模式</div>
                    <div style="font-size:11px;color:#8A8F98;">完整参数，精细调优</div>
                  </div>
                </div>
              </div>

              <!-- Simple mode form -->
              <div v-if="inlineConfigStep === 1 && inlineConfigMode === 'simple'" style="display:flex;flex-direction:column;gap:14px;">
                <div style="font-weight:600;font-size:14px;">简易配置</div>
                <div>
                  <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">视频编码器</div>
                  <select v-model="inlineNewConfig.encoder">
                    <option value="x264">x264 (H.264)</option>
                    <option value="x265">x265 (H.265)</option>
                    <option value="h264_nvenc">NVENC H.264</option>
                    <option value="hevc_nvenc">NVENC H.265</option>
                  </select>
                </div>
                <div>
                  <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">视频质量</div>
                  <select v-model="inlineNewConfig.quality">
                    <option value="lossless">无损</option>
                    <option value="high">高</option>
                    <option value="medium">中</option>
                    <option value="low">低</option>
                  </select>
                </div>
                <div>
                  <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">编码速度</div>
                  <select v-model="inlineNewConfig.speed">
                    <option value="slow">慢 (高质量)</option>
                    <option value="medium">平衡</option>
                    <option value="fast">快 (低质量)</option>
                  </select>
                </div>
                <div>
                  <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">位深</div>
                  <select v-model="inlineNewConfig.depth">
                    <option value="10">10-bit</option>
                    <option value="8">8-bit</option>
                  </select>
                </div>
                <div>
                  <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">配置名称</div>
                  <input v-model="inlineNewConfig.name" placeholder="输入配置名称" />
                </div>
                <div style="display:flex;justify-content:flex-end;gap:8px;">
                  <button class="btn-ghost" @click="cancelInlineConfig">取消</button>
                  <button class="btn-primary" @click="saveInlineConfig">保存配置</button>
                </div>
              </div>

              <!-- Pro mode form -->
              <div v-if="inlineConfigStep === 1 && inlineConfigMode === 'pro'" style="display:flex;flex-direction:column;gap:14px;">
                <div style="font-weight:600;font-size:14px;">专业配置</div>
                <div>
                  <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">配置名称</div>
                  <input v-model="inlineNewConfig.name" placeholder="输入配置名称" />
                </div>
                <div>
                  <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">视频编码器</div>
                  <select v-model="inlineNewConfig.encoder">
                    <option value="x264">x264 (H.264)</option>
                    <option value="x265">x265 (H.265)</option>
                    <option value="h264_nvenc">NVENC H.264</option>
                    <option value="hevc_nvenc">NVENC H.265</option>
                  </select>
                </div>
                <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
                  <div>
                    <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">CRF</div>
                    <input type="number" v-model.number="inlineNewConfig.crf" min="0" max="51" />
                  </div>
                  <div>
                    <div style="color:#8A8F98;font-size:12px;margin-bottom:4px;">Preset</div>
                    <select v-model="inlineNewConfig.preset">
                      <option value="ultrafast">ultrafast</option>
                      <option value="veryfast">veryfast</option>
                      <option value="fast">fast</option>
                      <option value="medium">medium</option>
                      <option value="slow">slow</option>
                      <option value="slower">slower</option>
                      <option value="veryslow">veryslow</option>
                    </select>
                  </div>
                </div>
                <div style="display:flex;justify-content:flex-end;gap:8px;">
                  <button class="btn-ghost" @click="cancelInlineConfig">取消</button>
                  <button class="btn-primary" @click="saveInlineConfig">保存配置</button>
                </div>
              </div>
            </div>

            <div style="display:flex;justify-content:space-between;">
              <button class="btn-ghost" @click="step = 1">上一步</button>
              <button class="btn-primary" :disabled="selectedConfig === null || selectedConfig === -1" @click="step = 3">下一步</button>
            </div>
          </div>

          <!-- Step 4: 目标路径 -->
          <div v-if="step === 3">
            <div>
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">输出目录</label>
              <div style="display:flex;gap:8px;">
                <input type="text" v-model="outputPath" placeholder="/output" style="flex:1;" />
                <button class="btn-ghost" style="padding:10px 18px;white-space:nowrap;" @click="openFilePicker('output')">浏览</button>
              </div>
            </div>
            <div style="margin-top:16px;">
              <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">输出文件名模板</label>
              <input type="text" v-model="outputNameTemplate" placeholder="{disc}_{track}.mkv" />
            </div>
            <div style="font-size:12px;color:#8A8F98;margin-top:4px;">可用变量: {'{disc}'}, {'{track}'}, {'{episode}'}</div>

            <!-- Preview panel -->
            <div v-if="previewState !== 'idle'" class="glass" style="margin-top:16px;padding:20px;">
              <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
                <span style="font-weight:600;">转码预览</span>
                <span class="badge" :class="'badge-' + previewStatusBadge">{{ previewStatus }}</span>
              </div>

              <div v-if="previewState === 'running'" class="progress-track" style="margin-bottom:12px;">
                <div class="progress-fill" :style="{width: previewProgress + '%'}" />
              </div>
              <div v-if="previewState === 'running'" style="display:flex;justify-content:space-between;font-size:13px;color:#8A8F98;margin-bottom:12px;">
                <span>帧: {{ previewFrame }}/{{ previewTotalFrames }}</span>
                <span>速度: {{ previewSpeed }}x</span>
                <span>剩余: {{ previewETA }}</span>
              </div>

              <div v-if="previewState === 'completed'" style="margin-bottom:12px;font-size:14px;color:#22C55E;">
                ✓ 预览完成 — 输出: {{ previewOutputSize }}
              </div>

              <div v-if="previewState === 'failed'" style="margin-bottom:12px;font-size:14px;color:#EF4444;">
                ✗ 预览已取消
              </div>

              <div style="display:flex;gap:8px;">
                <button v-if="previewState === 'running'" class="btn-ghost" style="color:#EF4444;" @click="cancelPreview">取消预览</button>
                <button v-if="previewState === 'completed'" class="btn-primary" style="font-size:13px;" @click="downloadPreview">下载预览文件</button>
              </div>
            </div>

            <!-- Preview config (shown when idle) -->
            <div v-if="previewState === 'idle'" style="margin-top:12px;">
              <div style="display:flex;gap:12px;align-items:center;">
                <select v-model="previewDuration" style="width:auto;">
                  <option value="30">30 秒</option>
                  <option value="60">60 秒</option>
                  <option value="90">90 秒</option>
                  <option value="120">120 秒</option>
                </select>
                <input v-model="previewStartTime" placeholder="开始时间 (如: 00:05:00)" style="width:auto;flex:1;" />
              </div>
              <button class="btn-primary" @click="startPreview" style="font-size:13px;margin-top:8px;">开始预览</button>
            </div>

            <div style="margin-top:24px;display:flex;justify-content:space-between;">
              <button class="btn-ghost" @click="step = 2">上一步</button>
              <button class="btn-primary" @click="finish">完成</button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- File picker modal -->
    <div v-if="showFilePicker" style="position:fixed;inset:0;z-index:160;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.6);" @click.self="showFilePicker = false">
      <div class="glass" style="width:500px;padding:24px;">
        <h3 style="font-size:16px;font-weight:600;margin-bottom:16px;">选择{{ filePickerTarget === 'source' ? '源文件 / 目录' : '输出目录' }}</h3>
        <div style="margin-bottom:12px;color:#8A8F98;font-size:13px;">{{ filePickerTarget === 'source' ? '输入 BDMV 目录路径或 ISO 文件路径：' : '输入输出目录路径：' }}</div>
        <input v-model="filePickerPath" :placeholder="filePickerTarget === 'source' ? '/input/BDROM' : '/output'" @keyup.enter="selectPath" style="margin-bottom:16px;" />
        <div style="margin-bottom:16px;">
          <div style="color:#8A8F98;font-size:12px;margin-bottom:8px;">常见路径：</div>
          <div v-for="p in mockPaths" :key="p" @click="selectPathDirect(p)"
               style="padding:8px 12px;border-radius:8px;cursor:pointer;font-size:13px;color:#5E6AD2;"
               class="hover-span">
            📁 {{ p }}
          </div>
        </div>
        <div style="display:flex;justify-content:flex-end;gap:8px;">
          <button class="btn-ghost" @click="showFilePicker = false">取消</button>
          <button class="btn-primary" @click="selectPath">确认</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ close: [] }>()

const steps = ['源文件', '文件选择', '转码配置', '目标路径']
const step = ref(0)
const sourcePath = ref('')
const selectedFiles = ref<number[]>([])
const selectedConfig = ref<number | null>(null)
const outputPath = ref('/output')
const outputNameTemplate = ref('{disc}_{track}.mkv')

const mockDiscName = 'AMAZING_ANIME_BD_1'
const mockFiles = [
  { id: 0, name: 'BDMV/STREAM/00000.m2ts', size: '24.2 GB' },
  { id: 1, name: 'BDMV/STREAM/00001.m2ts', size: '19.8 GB' },
  { id: 2, name: 'BDMV/STREAM/00002.m2ts', size: '3.1 GB' },
  { id: 3, name: 'BDMV/STREAM/00003.m2ts', size: '6.4 GB' },
]

// ── Step 2: multi-select audio/subtitle + chapter toggle ──
const parsedAudioTracks = ref([
  { id: 'a1', label: 'FLAC 2.0', lang: 'Japanese' },
  { id: 'a2', label: 'FLAC 2.0', lang: 'Commentary' },
  { id: 'a3', label: 'AAC 2.0', lang: 'Japanese' },
])
const parsedSubtitleTracks = ref([
  { id: 's1', label: 'PGS', lang: 'Japanese' },
  { id: 's2', label: 'PGS', lang: 'English' },
  { id: 's3', label: 'PGS', lang: 'Chinese' },
])
const selectedAudio = ref<string[]>(['a1'])
const selectedSubtitles = ref<string[]>(['s1'])
const chaptersEnabled = ref(true)

function toggleTrack(track: any, type: string) {
  const arr = type === 'audio' ? selectedAudio : selectedSubtitles
  const idx = arr.value.indexOf(track.id)
  if (idx >= 0) arr.value.splice(idx, 1)
  else arr.value.push(track.id)
}

// ── Step 3: config list + inline creation ──
const allConfigs = ref([
  { id: 0, name: 'x265 HQ Anime', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 15, preset: 'slower', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', 'no-sao': true, bframes: 16, 'rc-lookahead': 60, subme: 7, merange: 57 } },
  { id: 1, name: 'x264 高画质 (mbtree on)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 18, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 1, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11, merange: 48 } },
  { id: 2, name: 'x265 均衡', encoder: 'x265', mode: 'cpu', isPreset: true, params: { crf: 20, preset: 'medium', 'aq-mode': 3, 'aq-strength': 0.7, deblock: '0:0', bframes: 8, 'rc-lookahead': 40 } },
  { id: 3, name: 'x264 高画质 (mbtree off)', encoder: 'x264', mode: 'cpu', isPreset: true, params: { crf: 16, preset: 'veryslow', tune: 'animation', 'aq-mode': 3, 'aq-strength': 0.8, deblock: '1:1', mbtree: 0, 'rc-lookahead': 250, ref: 16, bframes: 16, subme: 11 } },
  { id: 4, name: 'NVENC 快速编码', encoder: 'h264_nvenc', mode: 'gpu', isPreset: false, params: { crf: 20, preset: 'p7', tune: 'hq', rc: 'vbr', b_ref_mode: 'middle', multipass: 'qres', 'aq-strength': 8, lookahead: 32, bframes: 4 } },
  { id: 5, name: '我的自定义配置', encoder: 'x265', mode: 'cpu', isPreset: false, params: { crf: 17, preset: 'slow', 'aq-mode': 3, 'aq-strength': 0.9, deblock: '-1:-1', 'no-sao': true, bframes: 12 } },
])

const inlineConfigStep = ref(0)
const inlineConfigMode = ref<'simple' | 'pro'>('simple')
let nextInlineId = 6
const inlineNewConfig = ref({ name: '', encoder: 'x265', quality: 'high', speed: 'medium', depth: '10', crf: 18, preset: 'medium' })

function startSimpleConfig() { inlineConfigMode.value = 'simple'; inlineConfigStep.value = 1 }
function startProConfig() { inlineConfigMode.value = 'pro'; inlineConfigStep.value = 1 }
function cancelInlineConfig() {
  selectedConfig.value = null
  inlineConfigStep.value = 0
  inlineNewConfig.value = { name: '', encoder: 'x265', quality: 'high', speed: 'medium', depth: '10', crf: 18, preset: 'medium' }
}
function saveInlineConfig() {
  const name = inlineNewConfig.value.name || '未命名配置'
  const mode = inlineNewConfig.value.encoder.includes('nvenc') ? 'gpu' : 'cpu'
  const qualityMap: Record<string, number> = { lossless: 0, high: 15, medium: 20, low: 23 }
  const speedMap: Record<string, string> = { slow: 'slower', medium: 'medium', fast: 'veryfast' }
  const newId = nextInlineId++
  allConfigs.value.push({
    id: newId, name, encoder: inlineNewConfig.value.encoder, mode, isPreset: false,
    params: inlineConfigMode.value === 'pro'
      ? { crf: inlineNewConfig.value.crf, preset: inlineNewConfig.value.preset }
      : { crf: qualityMap[inlineNewConfig.value.quality] || 18, preset: speedMap[inlineNewConfig.value.speed] || 'medium' },
  })
  selectedConfig.value = newId
  cancelInlineConfig()
}

// ── File picker (shared between Step 1 and Step 4) ──
const showFilePicker = ref(false)
const filePickerPath = ref('')
const filePickerTarget = ref<'source' | 'output'>('source')
const mockPaths = ['/input/BDROM', '/input/四月は君の嘘_Vol1', '/input/Charlotte_Vol1.iso', '/mnt/bdmv']

function openFilePicker(target: 'source' | 'output') { filePickerTarget.value = target; showFilePicker.value = true; filePickerPath.value = '' }
function selectPathDirect(p: string) { filePickerPath.value = p; selectPath() }
function selectPath() {
  if (filePickerPath.value) {
    if (filePickerTarget.value === 'source') sourcePath.value = filePickerPath.value
    else outputPath.value = filePickerPath.value
    showFilePicker.value = false
  }
}

// ── Step 4: preview simulation ──
const previewState = ref<'idle'|'running'|'completed'|'failed'>('idle')
const previewProgress = ref(0)
const previewFrame = ref(0)
const previewTotalFrames = ref(1200)
const previewSpeed = ref('3.2')
const previewETA = ref('18s')
const previewOutputSize = ref('12.4 MB')
const previewDuration = ref('60')
const previewStartTime = ref('00:05:00')
const previewStatus = ref('')
const previewStatusBadge = ref('')
let previewInterval: ReturnType<typeof setInterval> | null = null

function startPreview() {
  previewState.value = 'running'
  previewStatus.value = '转码中...'
  previewStatusBadge.value = 'running'
  previewProgress.value = 0
  previewFrame.value = 0
  const interval = setInterval(() => {
    previewProgress.value += Math.random() * 8
    previewFrame.value = Math.floor(previewProgress.value / 100 * previewTotalFrames.value)
    if (previewProgress.value >= 100) {
      previewProgress.value = 100
      previewFrame.value = previewTotalFrames.value
      previewState.value = 'completed'
      previewStatus.value = '已完成'
      previewStatusBadge.value = 'completed'
      clearInterval(interval)
    }
  }, 300)
  previewInterval = interval
}
function cancelPreview() {
  if (previewInterval) clearInterval(previewInterval)
  previewState.value = 'failed'
  previewStatus.value = '已取消'
  previewStatusBadge.value = 'failed'
}
function downloadPreview() { /* mock download */ }

// ── Completion ──
const wizardComplete = ref(false)
const createdTaskName = ref('')

function finish() { createdTaskName.value = mockDiscName + '_转码任务'; wizardComplete.value = true }

function closeWizard() {
  wizardComplete.value = false; step.value = 0; sourcePath.value = ''; selectedFiles.value = []
  selectedConfig.value = null; previewState.value = 'idle'
  emit('close')
}

// Reset state when wizard opens
watch(() => props.visible, (val) => {
  if (val) { step.value = 0; sourcePath.value = ''; selectedFiles.value = []; selectedConfig.value = null; outputPath.value = '/output'; previewState.value = 'idle'; wizardComplete.value = false }
})
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
.wiz-panel { padding: 32px; animation: wizIn 250ms cubic-bezier(0.16,1,0.3,1); }
@keyframes wizIn {
  from { opacity: 0; transform: scale(0.95) translateY(10px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
.hover-span:hover { background: rgba(94,106,210,0.1); }
</style>
