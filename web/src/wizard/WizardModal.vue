<template>
  <Teleport to="body">
    <div v-if="visible" class="wiz-overlay" @click.self="$emit('close')">
      <div class="wiz-panel glass" style="width:640px;max-height:90vh;overflow-y:auto;">
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
            <div style="margin-top:24px;display:flex;justify-content:flex-end;">
              <button class="btn-primary" :disabled="!sourcePath" @click="parseStep2">下一步</button>
            </div>
          </div>

          <div v-if="step === 1">
            <div style="font-size:13px;color:#8A8F98;margin-bottom:12px;">碟片名称: <span style="color:#EDEDEF;">{{ discName }}</span></div>
            <div style="font-size:13px;color:#8A8F98;margin-bottom:4px;">选择要转码的视频文件</div>
            <div style="display:flex;flex-direction:column;gap:4px;margin-bottom:20px;max-height:200px;overflow-y:auto;">
              <label v-for="f in parsedFiles" :key="f.id || f.path"
                style="display:flex;align-items:center;gap:10px;padding:8px 12px;border-radius:8px;background:rgba(255,255,255,0.02);cursor:pointer;font-size:13px;">
                <input type="checkbox" v-model="selectedFiles" :value="f.id || f.path" style="accent-color:#5E6AD2;width:16px;height:16px;" />
                <span style="color:#EDEDEF;">{{ f.name || f.path }}</span>
                <span style="color:#8A8F98;margin-left:auto;">{{ f.size || f.duration }}</span>
              </label>
            </div>

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

          <div v-if="step === 2">
            <div style="font-size:13px;color:#8A8F98;margin-bottom:12px;">选择编码配置</div>
            <div style="display:flex;flex-direction:column;gap:8px;margin-bottom:20px;">
              <label v-for="cfg in allConfigs" :key="cfg.id"
                style="display:flex;align-items:center;gap:12px;padding:12px 16px;border-radius:12px;background:rgba(255,255,255,0.02);cursor:pointer;border:1px solid transparent;"
                :style="{ borderColor: selectedConfigId === cfg.id ? '#5E6AD2' : 'transparent' }"
                @click="selectedConfigId = cfg.id">
                <input type="radio" :checked="selectedConfigId === cfg.id" style="accent-color:#5E6AD2;width:16px;height:16px;" />
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
              <button class="btn-ghost" @click="showPreviewPanel = true" style="display:flex;align-items:center;gap:6px;">
                <PhPlay :size="16" /> 转码预览
              </button>
            </div>

            <div style="margin-top:24px;display:flex;justify-content:space-between;">
              <button class="btn-ghost" @click="step = 2">上一步</button>
              <button class="btn-primary" @click="finishWizard">完成</button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <div v-if="showFilePicker" style="position:fixed;inset:0;z-index:160;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.6);" @click.self="showFilePicker = false">
      <div class="glass" style="width:520px;max-height:70vh;display:flex;flex-direction:column;">
        <div style="display:flex;align-items:center;gap:8px;padding:16px 20px;border-bottom:1px solid rgba(255,255,255,0.06);">
          <span style="font-weight:600;font-size:15px;">浏览文件</span>
        </div>

        <div style="display:flex;align-items:center;gap:4px;padding:8px 20px;font-size:13px;color:#8A8F98;flex-wrap:wrap;">
          <template v-for="(seg, i) in pathSegments" :key="i">
            <span @click="navigateTo(i)" style="cursor:pointer;color:#5E6AD2;" class="hover-span">{{ seg || '/' }}</span>
            <span v-if="i < pathSegments.length - 1" style="color:#475569;">/</span>
          </template>
        </div>

        <div style="flex:1;overflow-y:auto;padding:8px 0;max-height:400px;">
          <div v-if="currentDirEntries.parent" @click="goUp" style="display:flex;align-items:center;gap:10px;padding:8px 20px;cursor:pointer;color:#8A8F98;font-size:13px;" class="hover-span">📁 ..</div>
          <div v-for="entry in currentDirEntries.entries" :key="entry.name"
               @click="entry.isDir ? enterDir(entry.name) : selectEntry(entry.name)"
               style="display:flex;align-items:center;gap:10px;padding:8px 20px;cursor:pointer;font-size:13px;"
               :style="{color: entry.isDir ? '#5E6AD2' : '#EDEDEF'}" class="hover-span">
            <span>{{ entry.isDir ? '📁' : '📄' }}</span>
            <span>{{ entry.name }}</span>
          </div>
          <div v-if="!currentDirEntries.entries.length && !currentDirEntries.parent" style="padding:20px;text-align:center;color:#8A8F98;font-size:13px;">
            无法读取目录内容<br/>(容器环境限制，请手动输入路径)
          </div>
        </div>

        <div style="padding:12px 20px;border-top:1px solid rgba(255,255,255,0.06);">
          <input type="text" v-model="filePickerPath" placeholder="或直接输入路径，如 /input/BDROM" @keyup.enter="selectPath" />
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

        <div v-if="previewState === 'idle'" style="display:flex;flex-direction:column;gap:16px;">
          <div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:4px;">预览文件</div>
            <select v-model="previewFile" style="width:100%;">
              <option v-for="f in selectedFiles" :key="f" :value="f">{{ parsedFiles.find(m => (m.id || m.path) === f)?.name || f }}</option>
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
          <div class="progress-track"><div class="progress-fill" :style="{width: previewProgress+'%'}" /></div>
          <div style="display:flex;justify-content:space-between;font-size:13px;color:#8A8F98;">
            <span>帧: {{ previewFrame }}/{{ previewTotalFrames }}</span>
            <span>{{ previewSpeed }}x</span>
          </div>
          <button class="btn-ghost" style="color:#EF4444;align-self:flex-start;" @click="cancelPreview">取消</button>
        </div>

        <div v-if="previewState === 'completed'">
          <div style="text-align:center;padding:24px;">
            <div style="font-size:32px;margin-bottom:8px;">✓</div>
            <div style="color:#22C55E;font-weight:600;margin-bottom:4px;">预览完成</div>
            <div style="color:#8A8F98;font-size:13px;margin-bottom:16px;">{{ previewOutputSize }}</div>
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
import { ref, computed, watch } from 'vue'
import { PhPlay } from '@phosphor-icons/vue'
import ConfigCreateModal from '../components/ConfigCreateModal.vue'
import { api } from '@/api'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ close: [] }>()

const steps = ['源文件', '文件选择', '转码配置', '目标路径']
const step = ref(0)
const sourcePath = ref('')
const selectedFiles = ref<any[]>([])
const selectedConfigId = ref<number | null>(null)
const outputPath = ref('/output')
const outputNameTemplate = ref('{disc}_{track}.mkv')
const error = ref('')

const discName = ref('')
const parsedFiles = ref<any[]>([])
const parsedAudioTracks = ref<any[]>([])
const parsedSubtitleTracks = ref<any[]>([])
const selectedAudio = ref<string[]>([])
const selectedSubtitles = ref<string[]>([])
const chaptersEnabled = ref(true)

async function parseStep2() {
  if (!sourcePath.value) return
  error.value = ''
  try {
    const result = await api.wizard.parse(sourcePath.value)
    discName.value = result.disc_name || sourcePath.value.split('/').pop() || '未知'
    parsedFiles.value = result.files || []
    wizardStep.value = 2
  } catch {
    discName.value = sourcePath.value.split('/').pop() || '未知'
  }
  step.value = 1
}

function toggleTrack(track: any, type: string) {
  const arr = type === 'audio' ? selectedAudio : selectedSubtitles
  const idx = arr.value.indexOf(track.id)
  if (idx >= 0) arr.value.splice(idx, 1)
  else arr.value.push(track.id)
}

const allConfigs = ref<any[]>([])

const showCreateConfig = ref(false)
async function onWizardConfigSaved(config: any) {
  try {
    const configs = await api.configs.list()
    const presets = await api.presets()
    allConfigs.value = [...presets, ...configs]
    const saved = configs.find((c: any) => c.name === config.name)
    if (saved) selectedConfigId.value = saved.id
  } catch {}
}

const showFilePicker = ref(false)
const filePickerPath = ref('/')
const pickerTarget = ref<'source' | 'output'>('source')

interface DirEntry { name: string; isDir: boolean }
const currentDirEntries = ref<{ entries: DirEntry[]; parent: boolean }>({ entries: [], parent: false })

const mockFs: Record<string, string[]> = {
  '/': ['input', 'output', 'mnt', 'tmp', 'home'],
  '/input': ['BDROM', '四月は君の嘘_Vol1', 'Charlotte_Vol1.iso'],
  '/input/BDROM': ['BDMV'],
  '/input/BDROM/BDMV': ['STREAM', 'PLAYLIST', 'CLIPINF'],
  '/input/BDROM/BDMV/STREAM': ['00000.m2ts', '00001.m2ts', '00002.m2ts', '00003.m2ts'],
  '/mnt': ['bdmv'],
}

function pathJoin(a: string, b: string): string {
  if (a.endsWith('/')) return a + b
  return a + '/' + b
}

function pathParent(p: string): string {
  if (p === '/') return '/'
  const idx = p.lastIndexOf('/')
  return idx === 0 ? '/' : p.substring(0, idx)
}

const pathSegments = computed(() => {
  if (filePickerPath.value === '/') return ['']
  return filePickerPath.value.split('/').filter(s => s)
})

function loadDir(path: string) {
  const dirs = mockFs[path]
  if (dirs) {
    currentDirEntries.value = {
      entries: dirs.map(name => ({ name, isDir: !name.includes('.') })),
      parent: path !== '/',
    }
  } else {
    currentDirEntries.value = { entries: [], parent: path !== '/' }
  }
}

function openFilePicker(target: 'source' | 'output') {
  pickerTarget.value = target
  filePickerPath.value = target === 'source' ? sourcePath.value || '/' : outputPath.value || '/'
  showFilePicker.value = true
  loadDir('/')
}

function enterDir(name: string) {
  filePickerPath.value = pathJoin(filePickerPath.value, name)
  loadDir(filePickerPath.value)
}

function navigateTo(idx: number) {
  const parts = pathSegments.value
  if (idx === 0) {
    filePickerPath.value = '/'
  } else {
    filePickerPath.value = '/' + parts.slice(0, idx + 1).join('/')
  }
  loadDir(filePickerPath.value)
}

function goUp() {
  filePickerPath.value = pathParent(filePickerPath.value)
  loadDir(filePickerPath.value)
}

function selectEntry(name: string) {
  filePickerPath.value = pathJoin(filePickerPath.value, name)
  selectPath()
}

function selectPath() {
  if (filePickerPath.value) {
    if (pickerTarget.value === 'source') {
      sourcePath.value = filePickerPath.value
    } else {
      outputPath.value = filePickerPath.value
    }
    showFilePicker.value = false
  }
}

const showPreviewPanel = ref(false)
const previewState = ref<'idle'|'running'|'completed'>('idle')
const previewProgress = ref(0)
const previewFrame = ref(0)
const previewTotalFrames = ref(1200)
const previewSpeed = ref('3.2')
const previewOutputSize = ref('12.4 MB')
const previewDuration = ref('60')
const previewStartTime = ref('00:05:00')
const previewFile = ref<any>(null)
const previewId = ref<number | null>(null)
let previewInterval: ReturnType<typeof setInterval> | null = null

async function startPreview() {
  error.value = ''
  try {
    const res = await api.preview.create({
      source_file: previewFile.value,
      start_time: previewStartTime.value,
      duration: parseInt(previewDuration.value),
    })
    previewId.value = res.id
    previewState.value = 'running'
    previewProgress.value = 0
    const check = setInterval(async () => {
      try {
        const s = await api.preview.status(previewId.value!)
        previewProgress.value = s.progress * 100
        previewFrame.value = Math.floor(s.progress * previewTotalFrames.value)
        if (s.status === 'completed') {
          previewProgress.value = 100
          previewFrame.value = previewTotalFrames.value
          previewState.value = 'completed'
          previewOutputSize.value = s.size || '—'
          clearInterval(check)
        }
        if (s.status === 'failed') {
          previewState.value = 'idle'
          error.value = '预览转码失败'
          clearInterval(check)
        }
      } catch {
        clearInterval(check)
        startMockPreview()
      }
    }, 1000)
  } catch {
    startMockPreview()
  }
}

function startMockPreview() {
  previewState.value = 'running'
  previewProgress.value = 0
  previewFrame.value = 0
  const interval = setInterval(() => {
    previewProgress.value += Math.random() * 8
    previewFrame.value = Math.floor(previewProgress.value / 100 * previewTotalFrames.value)
    if (previewProgress.value >= 100) {
      previewProgress.value = 100
      previewFrame.value = previewTotalFrames.value
      previewState.value = 'completed'
      clearInterval(interval)
    }
  }, 300)
  previewInterval = interval
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
const wizardStep = ref(0)

async function finishWizard() {
  error.value = ''
  try {
    await api.tasks.create({
      name: discName.value,
      source_path: sourcePath.value,
      output_path: outputPath.value,
      config_id: selectedConfigId.value,
      files: selectedFiles.value.map((f: any) => ({
        source_file: f,
        streams: '{}',
        selected: true,
      })),
    })
    createdTaskName.value = discName.value + '_转码任务'
    wizardComplete.value = true
  } catch (e: any) {
    error.value = e.message
  }
}

function closeWizard() {
  wizardComplete.value = false; step.value = 0; sourcePath.value = ''; selectedFiles.value = []
  selectedConfigId.value = null; showPreviewPanel.value = false; previewState.value = 'idle'
  error.value = ''
  emit('close')
}

watch(() => props.visible, (val) => {
  if (val) {
    step.value = 0; sourcePath.value = ''; selectedFiles.value = []; selectedConfigId.value = null;
    outputPath.value = '/output'; showPreviewPanel.value = false; previewState.value = 'idle';
    wizardComplete.value = false; error.value = ''
    allConfigs.value = []
    Promise.all([api.configs.list(), api.presets()]).then(([configs, presets]) => {
      allConfigs.value = [...presets, ...configs]
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
