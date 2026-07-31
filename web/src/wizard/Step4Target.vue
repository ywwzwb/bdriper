<template>
  <div>
    <div class="mb-6">
      <h3 class="text-lg font-bold text-fg mb-1">输出与预览</h3>
      <p class="text-sm text-muted">设置输出目录并预览转码效果</p>
    </div>

    <div class="mb-4">
      <label class="block text-sm font-medium text-fg mb-2">输出目录</label>
      <div class="flex gap-2">
        <input v-model="outputPath" type="text" placeholder="/path/to/output" class="flex-1 bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg font-mono text-sm focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200" />
        <button class="px-4 py-2.5 border border-border/50 rounded-lg hover:bg-muted transition-all duration-200 text-sm">浏览</button>
      </div>
    </div>

    <div class="bg-muted/30 rounded-xl p-4">
      <div class="flex items-center justify-between mb-3">
        <h4 class="text-sm font-medium text-fg">预览转码</h4>
        <button v-if="!showPreview" @click="showPreview = true" class="px-4 py-2 text-sm border border-border/50 rounded-lg hover:bg-muted transition-all duration-200">预览</button>
      </div>

      <div v-if="showPreview" class="space-y-4">
        <div>
          <label class="block text-xs font-medium text-fg mb-1.5">测试文件</label>
          <select v-model="previewFile" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg text-sm focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200">
            <option v-for="(f, i) in allFiles" :key="i" :value="i">{{ f.name || `File ${i + 1}` }}</option>
          </select>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-medium text-fg mb-1.5">起始时间</label>
            <input v-model="startTime" type="text" placeholder="00:00:00" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg text-sm font-mono focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200" />
          </div>
          <div>
            <label class="block text-xs font-medium text-fg mb-1.5">时长</label>
            <select v-model="duration" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg text-sm focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200">
              <option value="30">30s</option>
              <option value="60">60s</option>
              <option value="90">90s</option>
              <option value="120">120s</option>
            </select>
          </div>
        </div>

        <div class="flex gap-2">
          <button v-if="!previewRunning && !previewDone" @click="startPreview" class="px-4 py-2 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 text-sm shadow-lg shadow-accent/20 flex items-center gap-2" :disabled="startingPreview">
            <svg v-if="startingPreview" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
            {{ startingPreview ? '启动中...' : '开始预览' }}
          </button>
          <button v-if="previewRunning" @click="cancelPreview" class="px-4 py-2 border border-destructive/50 text-destructive rounded-lg hover:bg-destructive/10 transition-all duration-200 text-sm">取消</button>
          <a v-if="previewDone && downloadLink" :href="downloadLink" class="px-4 py-2 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 text-sm">下载预览</a>
        </div>

        <div v-if="previewRunning" class="w-full bg-muted rounded-full h-1.5">
          <div class="bg-accent h-1.5 rounded-full transition-all duration-500" :style="{ width: (previewProgress ?? 0) + '%' }" />
        </div>

        <div v-if="previewStatus" class="text-sm" :class="previewError ? 'text-destructive' : 'text-muted'">{{ previewError || previewStatus }}</div>
      </div>
    </div>

    <div v-if="error" class="bg-destructive/10 border border-destructive/30 text-destructive rounded-lg px-4 py-3 text-sm mt-4">{{ error }}</div>

    <div class="flex gap-3 pt-6">
      <button @click="$emit('back')" class="px-5 py-2.5 border border-border/50 rounded-lg hover:bg-muted transition-all duration-200 text-sm font-medium">上一步</button>
      <button @click="submit" class="px-6 py-2.5 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 text-sm ml-auto shadow-lg shadow-accent/20 disabled:opacity-50 flex items-center gap-2" :disabled="submitting || !outputPath">
        <svg v-if="submitting" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
        {{ submitting ? '创建中...' : '创建任务' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { api } from '@/api'

const props = defineProps<{ data: any }>()
const emit = defineEmits<{ back: [], done: [] }>()

const outputPath = ref('')
const showPreview = ref(false)
const previewFile = ref(0)
const startTime = ref('00:00:00')
const duration = ref('60')
const previewRunning = ref(false)
const previewDone = ref(false)
const startingPreview = ref(false)
const previewStatus = ref('')
const previewError = ref('')
const previewProgress = ref(0)
const downloadLink = ref('')
const previewId = ref<number | null>(null)
const error = ref('')
const submitting = ref(false)

const allFiles = computed(() => [...(props.data.files || []), ...(props.data.sub_files || [])])

async function startPreview() {
  startingPreview.value = true; previewError.value = ''; previewStatus.value = ''; previewProgress.value = 0
  try {
    const res = await api.preview.create({
      source_path: props.data.source_path,
      file_index: previewFile.value,
      start_time: startTime.value,
      duration: parseInt(duration.value),
      config_id: props.data.config_id,
      new_config: props.data.new_config,
    })
    previewId.value = res.id
    props.data.preview_id = res.id
    previewRunning.value = true
    pollPreview()
  } catch (e: any) {
    previewError.value = e.message || '预览启动失败'
  } finally {
    startingPreview.value = false
  }
}

async function pollPreview() {
  if (!previewId.value) return
  try {
    const status = await api.preview.status(previewId.value)
    previewProgress.value = status.progress || 0
    previewStatus.value = status.status || `进度: ${status.progress || 0}%`
    if (status.status === 'completed' || status.progress >= 100) {
      previewRunning.value = false; previewDone.value = true
      downloadLink.value = api.preview.downloadUrl(previewId.value!, status.token || '')
    } else if (status.status === 'failed') {
      previewRunning.value = false; previewError.value = status.error || '预览失败'
    } else {
      setTimeout(pollPreview, 2000)
    }
  } catch {
    setTimeout(pollPreview, 2000)
  }
}

async function cancelPreview() {
  if (previewId.value) { try { await api.preview.delete(previewId.value) } catch { /* ignore */ } }
  previewRunning.value = false; previewId.value = null
}

async function submit() {
  if (!outputPath.value) return
  submitting.value = true; error.value = ''
  try {
    await api.tasks.create({
      source_path: props.data.source_path,
      output_path: outputPath.value,
      files: props.data.files,
      sub_files: props.data.sub_files,
      audio_tracks: props.data.audio_tracks,
      subtitle_tracks: props.data.subtitle_tracks,
      include_chapters: props.data.include_chapters,
      config_id: props.data.config_id,
      new_config: props.data.new_config,
    })
    emit('done')
  } catch (e: any) {
    error.value = e.message || '创建任务失败'
  } finally {
    submitting.value = false
  }
}
</script>
