<template>
  <Teleport to="body">
    <div v-if="visible" class="wiz-overlay" @click.self="$emit('close')">
      <div class="wiz-panel glass" style="width:640px;max-height:90vh;overflow-y:auto;position:relative;">
        <div v-if="parsing" style="position:absolute;inset:0;display:flex;flex-direction:column;align-items:center;justify-content:center;background:rgba(2,2,3,0.92);backdrop-filter:blur(8px);-webkit-backdrop-filter:blur(8px);border-radius:16px;z-index:5;">
          <div style="width:36px;height:36px;border:3px solid rgba(255,255,255,0.1);border-top-color:#5E6AD2;border-radius:50%;animation:spin 0.7s linear infinite;margin-bottom:16px;" />
          <div style="font-size:15px;color:#EDEDEF;font-weight:500;">正在解析 BDMV 文件...</div>
          <div style="font-size:12px;color:#8A8F98;margin-top:4px;">{{ sourcePath }}</div>
        </div>
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
          <h2 style="font-size:20px;font-weight:700;">新建转码任务</h2>
          <button class="btn-ghost" style="padding:6px 12px;font-size:12px;" @click="$emit('close')">取消</button>
        </div>

        <div v-if="wizardComplete" style="text-align:center;padding:60px 20px;">
          <div style="font-size:48px;margin-bottom:16px;">✓</div>
          <div style="font-size:20px;font-weight:600;margin-bottom:8px;">任务创建成功</div>
          <div style="color:#8A8F98;margin-bottom:24px;">{{ createdTaskName }} 已加入任务队列</div>
          <button class="btn-primary" @click="closeWizard">返回任务列表</button>
        </div>

        <template v-if="!wizardComplete">
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

          <div v-if="error" style="color:#EF4444;font-size:13px;margin-bottom:16px;padding:8px 12px;border-radius:8px;background:rgba(239,68,68,0.1);">{{ error }}</div>

          <div v-if="step === 0">
            <label style="font-size:13px;color:#8A8F98;display:block;margin-bottom:6px;">选择蓝光光盘目录</label>
            <div style="display:flex;gap:8px;">
              <input type="text" v-model="sourcePath" placeholder="/input/BDROM" style="flex:1;" />
              <button class="btn-ghost" style="padding:10px 18px;white-space:nowrap;" @click="openFilePicker('source')">浏览</button>
            </div>
            <div v-if="sourcePath" style="margin-top:8px;font-size:12px;color:#22C55E;">已选择: {{ sourcePath }}</div>
            <div v-if="parseError" style="margin-top:12px;color:#EF4444;font-size:13px;">{{ parseError }}</div>
            <div style="margin-top:24px;display:flex;justify-content:flex-end;">
              <button class="btn-primary" @click="parseStep2" :disabled="!sourcePath || parsing" style="display:inline-flex;align-items:center;gap:8px;">
                <span v-if="parsing" class="spinner" /> {{ parsing ? '解析中...' : '下一步' }}
              </button>
            </div>
          </div>

          <div v-if="step === 1">
            <div style="font-size:13px;color:#8A8F98;margin-bottom:12px;">碟片名称: <span style="color:#EDEDEF;">{{ discName }}</span></div>
            <div style="font-size:13px;color:#8A8F98;margin-bottom:8px;">选择要转码的视频文件 ({{ selectedFiles.length }}/{{ parsedFiles.length }})</div>
            <div style="display:flex;flex-direction:column;gap:8px;margin-bottom:20px;max-height:380px;overflow-y:auto;">
              <div v-for="f in parsedFiles" :key="f.path" style="border-radius:8px;background:rgba(255,255,255,0.03);padding:10px 12px;">
                <div style="display:flex;align-items:center;gap:10px;font-size:13px;margin-bottom:6px;">
                  <input type="checkbox" v-model="selectedFiles" :value="f.path" style="accent-color:#5E6AD2;width:16px;height:16px;" />
                  <span style="color:#EDEDEF;flex:1;">{{ f.path.split('/').pop() }}</span>
                  <span style="color:#8A8F98;font-size:12px;">{{ f.duration }}  {{ f.resolution !== '?' ? f.resolution : '' }}</span>
                </div>
                <!-- Per-file tracks -->
                <div v-if="fileTracks[f.path]?.loading" style="padding-left:26px;font-size:12px;color:#8A8F98;">
                  <span class="spinner" style="display:inline-block;width:10px;height:10px;border:2px solid rgba(255,255,255,0.1);border-top-color:#5E6AD2;border-radius:50%;animation:spin 0.6s linear infinite;margin-right:6px;" /> 加载轨道...
                </div>
                <template v-else-if="fileTracks[f.path]">
                  <div v-if="fileTracks[f.path].audio.length" style="padding-left:26px;margin-top:4px;">
                    <div style="font-size:11px;color:#8A8F98;margin-bottom:2px;">音轨</div>
                    <div v-for="t in fileTracks[f.path].audio" :key="t.id"
                         style="display:flex;align-items:center;gap:8px;padding:3px 0;font-size:12px;cursor:pointer;"
                         class="hover-span" @click="t.selected = !t.selected">
                      <input type="checkbox" :checked="t.selected" style="accent-color:#5E6AD2;width:13px;height:13px;" @click.stop />
                      <span style="color:#EDEDEF;">{{ t.label }}</span>
                      <span style="color:#8A8F98;">{{ t.lang }}</span>
                    </div>
                  </div>
                  <div v-if="fileTracks[f.path].subtitle.length" style="padding-left:26px;margin-top:2px;">
                    <div style="font-size:11px;color:#8A8F98;margin-bottom:2px;">字幕</div>
                    <div v-for="t in fileTracks[f.path].subtitle" :key="t.id"
                         style="display:flex;align-items:center;gap:8px;padding:3px 0;font-size:12px;cursor:pointer;"
                         class="hover-span" @click="t.selected = !t.selected">
                      <input type="checkbox" :checked="t.selected" style="accent-color:#5E6AD2;width:13px;height:13px;" @click.stop />
                      <span style="color:#EDEDEF;">{{ t.label }}</span>
                      <span style="color:#8A8F98;">{{ t.lang }}</span>
                    </div>
                  </div>
                </template>
              </div>
            </div>

            <div style="display:flex;align-items:center;justify-content:space-between;padding:8px 12px;border-radius:8px;margin-bottom:16px;">
              <span style="font-size:14px;">章节信息</span>
              <div class="toggle-track" :class="{ on: chaptersEnabled }" @click="chaptersEnabled = !chaptersEnabled" style="width:44px;height:24px;flex-shrink:0;">
                <div class="toggle-knob" style="width:20px;height:20px;" />
              </div>
            </div>

            <div style="display:flex;justify-content:space-between;">
              <button class="btn-ghost" @click="step = 0">上一步</button>
              <button class="btn-primary" :disabled="!selectedFiles.length" @click="step = 2">下一步</button>
            </div>
          </div>

          <div v-if="step === 2">
            <div style="font-size:13px;color:#8A8F98;margin-bottom:12px;">选择编码配置</div>
            <div style="display:flex;flex-direction:column;gap:8px;margin-bottom:20px;">
              <label v-for="cfg in allConfigs" :key="cfgKey(cfg)"
                style="display:flex;align-items:center;gap:12px;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);cursor:pointer;border:1px solid transparent;"
                :style="{ borderColor: selectedConfigId === cfgKey(cfg) ? '#5E6AD2' : 'rgba(255,255,255,0.06)' }"
                @click="selectedConfigId = cfgKey(cfg)">
                <div style="width:18px;height:18px;border-radius:50%;border:2px solid;flex-shrink:0;display:flex;align-items:center;justify-content:center;"
                  :style="{ borderColor: selectedConfigId === cfgKey(cfg) ? '#5E6AD2' : '#475569' }">
                  <div v-if="selectedConfigId === cfgKey(cfg)" style="width:8px;height:8px;border-radius:50%;background:#5E6AD2;" />
                </div>
                <div style="flex:1;">
                  <div style="font-size:14px;font-weight:500;">{{ cfg.name }}
                    <span v-if="cfg.isPreset" class="badge" style="background:rgba(94,106,210,0.15);color:#5E6AD2;font-size:10px;padding:2px 6px;margin-left:6px;">内置</span>
                  </div>
                  <div style="font-size:12px;color:#8A8F98;margin-top:2px;">{{ cfg.encoder || cfg.video_encoder }} | {{ cfg.mode === 'gpu' ? 'GPU' : 'CPU' }}</div>
                </div>
              </label>
              <div style="display:flex;align-items:center;gap:12px;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);cursor:pointer;border:1px dashed rgba(255,255,255,0.1);"
                @click="showCreateConfig = true">
                <span style="font-size:14px;color:#8A8F98;">+ 新建配置</span>
              </div>
            </div>

            <div style="display:flex;justify-content:space-between;">
              <button class="btn-ghost" @click="step = 1">上一步</button>
              <button class="btn-primary" :disabled="selectedConfigId === null" @click="step = 3">下一步</button>
            </div>
          </div>

          <div v-if="step === 3">
            <div style="margin-bottom:24px;">
              <div style="color:#8A8F98;font-size:14px;margin-bottom:6px;">输出目录</div>
              <div style="display:flex;gap:8px;">
                <input type="text" v-model="outputPath" placeholder="/output" style="flex:1;" />
                <button class="btn-ghost" @click="openFilePicker('output')" style="padding:10px 16px;">浏览</button>
              </div>
            </div>

            <div style="margin-bottom:24px;">
              <div style="color:#8A8F98;font-size:14px;margin-bottom:6px;">输出文件名模板</div>
              <input type="text" v-model="outputNameTemplate" placeholder="{disc}_{track}.mkv" />
              <div style="font-size:12px;color:#8A8F98;margin-top:4px;">可用变量: {'{disc}'}, {'{track}'}, {'{episode}'}</div>
            </div>

            <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
              <div>
                <div style="font-size:14px;color:#8A8F98;">转码预览 (可选)</div>
                <div style="font-size:12px;color:#8A8F98;">转码一小段视频，检查效果</div>
              </div>
              <button class="btn-ghost" @click="openPreview" style="display:flex;align-items:center;gap:6px;">
                <PhPlay :size="16" /> 转码预览
              </button>
            </div>

            <div style="margin-top:24px;display:flex;flex-direction:column;gap:8px;">
              <div v-if="createError" style="color:#EF4444;font-size:13px;">{{ createError }}</div>
              <div style="display:flex;justify-content:space-between;">
                <button class="btn-ghost" @click="step = 2">上一步</button>
                <button class="btn-primary" @click="finishWizard" :disabled="!outputPath || creating" style="display:inline-flex;align-items:center;gap:8px;">
                  <span v-if="creating" class="spinner" /> {{ creating ? '创建中...' : '完成' }}
                </button>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div v-if="showFilePicker" style="position:fixed;inset:0;z-index:200;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.6);" @click.self="showFilePicker = false">
      <div class="glass" style="width:560px;max-height:70vh;display:flex;flex-direction:column;">
        <div style="display:flex;align-items:center;justify-content:space-between;padding:16px 20px;border-bottom:1px solid rgba(255,255,255,0.06);">
          <span style="font-weight:600;font-size:15px;">浏览文件</span>
          <button @click="showFilePicker = false" style="color:#8A8F98;font-size:20px;cursor:pointer;background:none;border:none;">✕</button>
        </div>

        <div style="padding:8px 20px;font-size:13px;color:#8A8F98;border-bottom:1px solid rgba(255,255,255,0.04);">
          📂 {{ browserPath }}
        </div>

        <div v-if="browserLoading" style="padding:40px;text-align:center;color:#8A8F98;">加载中...</div>

        <div v-else style="flex:1;overflow-y:auto;max-height:350px;">
          <div v-if="browserParent !== ''" @click="goUp" style="display:flex;align-items:center;gap:10px;padding:8px 20px;cursor:pointer;color:#5E6AD2;font-size:13px;border-bottom:1px solid rgba(255,255,255,0.03);" class="hover-span">
            📁 ..
          </div>
          <div v-for="entry in browserEntries" :key="entry.name"
               @click="entry.is_dir ? enterDir(entry.name) : selectFile(entry)"
               style="display:flex;align-items:center;gap:10px;padding:8px 20px;cursor:pointer;font-size:13px;border-bottom:1px solid rgba(255,255,255,0.03);"
               :style="{color: entry.is_dir ? '#5E6AD2' : '#EDEDEF'}" class="hover-span">
            <span>{{ entry.is_dir ? '📁' : '📄' }}</span>
            <span style="flex:1;">{{ entry.name }}</span>
            <span v-if="!entry.is_dir" style="font-size:11px;color:#8A8F98;">{{ formatSize(entry.size) }}</span>
          </div>
          <div v-if="browserError" style="padding:40px;text-align:center;color:#EF4444;font-size:13px;">
            {{ browserError }}
          </div>
          <div v-else-if="browserEntries.length === 0 && !browserLoading" style="padding:40px;text-align:center;color:#8A8F98;font-size:13px;">
            此目录为空
          </div>
        </div>

        <div style="padding:12px 20px;border-top:1px solid rgba(255,255,255,0.06);">
          <input type="text" v-model="filePickerPath" :placeholder="browserPath" @keyup.enter="selectPath" style="width:100%;" />
        </div>

        <div style="display:flex;justify-content:flex-end;gap:8px;padding:12px 20px;border-top:1px solid rgba(255,255,255,0.06);">
          <button class="btn-ghost" @click="showFilePicker = false">取消</button>
          <button class="btn-primary" @click="selectPath">确认</button>
        </div>
      </div>
    </div>

    <div v-if="showPreviewPanel" style="position:fixed;inset:0;z-index:160;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.6);" @click.self="closePreview">
      <div class="glass" style="width:500px;padding:24px;">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
          <h3 style="font-weight:600;font-size:16px;">转码预览</h3>
          <button @click="closePreview" style="color:#8A8F98;font-size:18px;cursor:pointer;background:none;border:none;">✕</button>
        </div>

        <div v-if="previewError" style="color:#EF4444;font-size:13px;margin-bottom:12px;">{{ previewError }}</div>

        <div v-if="previewState === 'idle'" style="display:flex;flex-direction:column;gap:16px;">
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">预览文件</div>
            <select v-model="previewFile" style="width:100%;">
              <option v-for="f in selectedFiles" :key="f" :value="f">{{ f.split('/').pop() }}</option>
            </select>
          </div>
          <div class="flex gap-4" style="display:flex;gap:12px;">
            <div style="flex:1;">
              <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">开始时间</div>
              <input type="text" v-model="previewStartTime" placeholder="00:05:00" />
            </div>
            <div style="flex:1;">
              <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">时长</div>
              <select v-model="previewDuration">
                <option value="30">30 秒</option>
                <option value="60">60 秒</option>
                <option value="90">90 秒</option>
                <option value="120">120 秒</option>
              </select>
            </div>
          </div>
          <button class="btn-primary" @click="startPreview" style="align-self:flex-start;">开始转码预览</button>
        </div>

        <div v-if="previewState === 'running'" style="display:flex;flex-direction:column;gap:12px;">
          <span class="badge badge-running">转码中...</span>
          <div class="progress-track"><div class="progress-fill" :style="{width: previewProgress + '%'}" /></div>
          <div style="font-size:13px;color:#EDEDEF;text-align:center;">{{ Math.floor(previewProgress) }}%</div>
          <button class="btn-ghost" style="color:#EF4444;align-self:center;" @click="cancelPreview">取消预览</button>
        </div>

        <div v-if="previewState === 'completed'">
          <div style="text-align:center;padding:24px;">
            <div style="font-size:32px;margin-bottom:8px;">✓</div>
            <div style="color:#22C55E;font-weight:600;margin-bottom:16px;">预览完成</div>
            <div style="display:flex;gap:8px;justify-content:center;">
              <button class="btn-primary" @click="downloadPreviewFile">下载预览文件</button>
              <button class="btn-ghost" @click="closePreview">关闭</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <ConfigCreateModal :visible="showCreateConfig" @close="showCreateConfig = false" @saved="onWizardConfigSaved" />
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { PhPlay } from '@phosphor-icons/vue'
import ConfigCreateModal from '../components/ConfigCreateModal.vue'
import { api } from '@/api'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ close: [] }>()

const steps = ['源文件', '文件选择', '转码配置', '目标路径']
const step = ref(0)
const sourcePath = ref(localStorage.getItem('bdriper_source_dir') || '')
const selectedFiles = ref<any[]>([])
const selectedConfigId = ref<string | null>(null)
const outputPath = ref(localStorage.getItem('bdriper_output_dir') || '')

watch(sourcePath, (v) => { if (v) localStorage.setItem('bdriper_source_dir', v) })
watch(outputPath, (v) => { if (v) localStorage.setItem('bdriper_output_dir', v) })
const outputNameTemplate = ref('{disc}_{track}.mkv')
const error = ref('')

const discName = ref('')
const parsedFiles = ref<any[]>([])
const chaptersEnabled = ref(true)
const parsing = ref(false)
const parseError = ref('')
const loadingStreams = ref(false)
const fileTracks = ref<Record<string, { audio: Track[], subtitle: Track[], loading: boolean }>>({})

type Track = { id: string; label: string; lang: string; selected: boolean }

async function loadFileTracks(filePath: string) {
  if (fileTracks.value[filePath]?.audio.length) return
  fileTracks.value[filePath] = { audio: [], subtitle: [], loading: true }
  try {
    const data = await api.wizard.fileStreams(filePath)
    fileTracks.value[filePath] = {
      audio: (data.audio || []).map((s: any) => ({
        id: String(s.index), label: `${s.codec.toUpperCase()} ${s.channels || ''}ch`, lang: s.language || '?', selected: true,
      })),
      subtitle: (data.subtitle || []).map((s: any) => ({
        id: String(s.index), label: s.codec.replace('hdmv_pgs_', '').toUpperCase(), lang: s.language || '?', selected: true,
      })),
      loading: false,
    }
  } catch {
    fileTracks.value[filePath] = { audio: [], subtitle: [], loading: false }
  }
}

async function parseStep2() {
  if (!sourcePath.value) return
  parsing.value = true
  parseError.value = ''
  try {
    const result = await api.wizard.parse(sourcePath.value)
    discName.value = result.disc_name || sourcePath.value.split('/').pop() || '未知'
    parsedFiles.value = result.files || []
    step.value = 1
    // Auto-select all main files
    selectedFiles.value = parsedFiles.value.filter((f: any) => f.is_main).map((f: any) => f.path)
    // Load tracks for all selected files
    selectedFiles.value.forEach((fp: string) => loadFileTracks(fp))
  } catch (e: any) {
    parseError.value = '解析失败: ' + (e.message || '未知错误')
  } finally {
    parsing.value = false
  }
}
const allConfigs = ref<any[]>([])

function cfgKey(cfg: any) {
  return cfg.id ? String(cfg.id) : cfg.name
}

const showCreateConfig = ref(false)
async function onWizardConfigSaved(config: any) {
  try {
    const configs = await api.configs.list()
    const presets = await api.presets()
    const presetNames = new Set(presets.map((p: any) => p.name))
    const userConfigs = configs.filter((c: any) => !presetNames.has(c.name))
    allConfigs.value = [...presets, ...userConfigs]
    const saved = configs.find((c: any) => c.name === config.name)
    if (saved) selectedConfigId.value = cfgKey(saved)
  } catch {}
}

const showFilePicker = ref(false)
const filePickerPath = ref('/')
const pickerTarget = ref<'source' | 'output'>('source')

interface DirEntry { name: string; is_dir: boolean; size: number }

const browserPath = ref('/')
const browserEntries = ref<DirEntry[]>([])
const browserLoading = ref(false)
const browserParent = ref('')

const browserError = ref('')

async function navigateTo(path: string) {
  browserPath.value = path
  browserLoading.value = true
  browserError.value = ''
  try {
    const res = await fetch(`/api/fs/list?path=${encodeURIComponent(path)}`)
    if (!res.ok) throw new Error('无法访问')
    const data = await res.json()
    browserEntries.value = data.entries || []
    browserParent.value = data.parent || ''
  } catch (e: any) {
    browserError.value = '无法读取该目录'
    browserEntries.value = []
    browserParent.value = ''
  } finally {
    browserLoading.value = false
  }
}

function enterDir(name: string) {
  const newPath = browserPath.value === '/' ? '/' + name : browserPath.value + '/' + name
  filePickerPath.value = newPath
  navigateTo(newPath)
}

function goUp() {
  if (browserParent.value) {
    filePickerPath.value = browserParent.value
    navigateTo(browserParent.value)
  }
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024*1024) return (bytes/1024).toFixed(1) + ' KB'
  if (bytes < 1024*1024*1024) return (bytes/(1024*1024)).toFixed(1) + ' MB'
  return (bytes/(1024*1024*1024)).toFixed(1) + ' GB'
}

function openFilePicker(target: 'source' | 'output') {
  pickerTarget.value = target
  const initialPath = target === 'source' ? sourcePath.value : outputPath.value
  filePickerPath.value = initialPath || '/'
  const navPath = initialPath || '/'
  navigateTo(navPath)
  showFilePicker.value = true
}

function selectFile(entry: DirEntry) {
  const newPath = browserPath.value === '/' ? '/' + entry.name : browserPath.value + '/' + entry.name
  filePickerPath.value = newPath
  selectPath()
}

function selectPath() {
  const p = filePickerPath.value
  if (!p) return
  if (pickerTarget.value === 'source') sourcePath.value = p
  else outputPath.value = p
  showFilePicker.value = false

  // If this is the source picker and we're on Step 1, auto-parse
  if (pickerTarget.value === 'source') {
    parseStep2()
  }
}

const showPreviewPanel = ref(false)
const previewError = ref('')
const previewFile = ref('')

function openPreview() {
  previewFile.value = selectedFiles.value[0] || ''
  previewError.value = ''
  showPreviewPanel.value = true
}
const previewState = ref<'idle'|'running'|'completed'>('idle')
const previewDuration = ref('60')
const previewStartTime = ref('00:00:00')
const previewProgress = ref(0)
const previewId = ref<number | null>(null)
let previewInterval: ReturnType<typeof setInterval> | null = null

async function startPreview() {
  previewError.value = ''
  try {
    // Find selected config's encoder and params
    const cfg = allConfigs.value.find(c => cfgKey(c) === selectedConfigId.value)
    const encoder = cfg?.encoder || cfg?.video_encoder || 'x264'
    const params = cfg?.params || {}

    const res = await api.preview.create({
      source_file: previewFile.value,
      start_time: previewStartTime.value,
      duration: parseInt(previewDuration.value),
      encoder,
      video_params: params,
    })
    previewId.value = res.id
    previewState.value = 'running'
    previewProgress.value = 0
    const check = setInterval(async () => {
      try {
        const s = await api.preview.status(previewId.value!)
        previewProgress.value = (s.progress || 0) * 100
        if (s.status === 'completed') {
          previewProgress.value = 100
          previewState.value = 'completed'
          clearInterval(check)
        }
        if (s.status === 'failed') {
          previewState.value = 'idle'
          error.value = '预览转码失败'
          clearInterval(check)
        }
      } catch {
        // keep polling on transient errors
      }
    }, 1000)
  } catch (e: any) {
    error.value = '创建预览失败: ' + (e.message || '')
  }
}

function cancelPreview() {
  if (previewInterval) clearInterval(previewInterval)
  if (previewId.value) {
    api.preview.delete(previewId.value).catch(() => {})
    previewId.value = null
  }
  previewState.value = 'idle'
}

function downloadPreviewFile() {
  if (previewId.value) {
    const url = api.preview.downloadUrl(previewId.value, 'preview')
    window.open(url, '_blank')
  }
}

function closePreview() {
  if (previewInterval) clearInterval(previewInterval)
  showPreviewPanel.value = false
  previewState.value = 'idle'
}

const wizardComplete = ref(false)
const createdTaskName = ref('')
const creating = ref(false)
const createError = ref('')

async function finishWizard() {
  if (!outputPath.value) return
  creating.value = true
  createError.value = ''
  try {
    const cfg = allConfigs.value.find(c => cfgKey(c) === selectedConfigId.value)
    let configId = cfg && typeof cfg.id === 'number' ? cfg.id : 0
    // For presets without DB id, check if a config with same name exists, else create one
    if (!configId && cfg) {
      const allCfgs = await api.configs.list()
      const existing = allCfgs.find((c: any) => c.name === cfg.name)
      if (existing) {
        configId = existing.id
      } else {
        const created = await api.configs.create({
          name: cfg.name,
          encoder_type: cfg.mode || 'cpu',
          video_encoder: cfg.encoder || cfg.video_encoder || 'x264',
          video_params: JSON.stringify(cfg.params || {}),
          audio_tracks: '[]',
          subtitle_tracks: '[]',
          chapters_enabled: true,
          output_muxer: 'mkvmerge',
        })
        configId = created.id
      }
    }
    await api.tasks.create({
      name: discName.value,
      source_path: sourcePath.value,
      output_path: outputPath.value,
      config_id: configId,
      files: selectedFiles.value.map((f: any) => ({
        source_file: f,
        streams: '{}',
        selected: true,
      })),
    })
    createdTaskName.value = discName.value + '_转码任务'
    wizardComplete.value = true
  } catch (e: any) {
    createError.value = '创建任务失败: ' + (e.message || '未知错误')
  } finally {
    creating.value = false
  }
}

function closeWizard() {
  wizardComplete.value = false; step.value = 0; selectedFiles.value = []
  selectedConfigId.value = null; showPreviewPanel.value = false; previewState.value = 'idle'
  error.value = ''; parsing.value = false; parseError.value = ''; creating.value = false; createError.value = ''
  loadingStreams.value = false
  emit('close')
}

watch(() => props.visible, (val) => {
  if (val) {
    step.value = 0; selectedFiles.value = []; selectedConfigId.value = null;
    showPreviewPanel.value = false; previewState.value = 'idle';
    wizardComplete.value = false; error.value = ''; parsing.value = false; parseError.value = ''
    creating.value = false; createError.value = ''
    loadingStreams.value = false
    loadingStreams.value = false
    allConfigs.value = []
    Promise.all([api.configs.list(), api.presets()]).then(([configs, presets]) => {
      const presetNames = new Set(presets.map((p: any) => p.name))
      const userConfigs = configs.filter((c: any) => !presetNames.has(c.name))
      allConfigs.value = [...presets, ...userConfigs]
    }).catch(() => {})
  }
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
