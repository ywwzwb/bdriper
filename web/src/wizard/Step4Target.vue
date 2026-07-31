<template>
  <div>
    <h3 class="text-lg font-medium mb-4">Step 4: Output & Preview</h3>

    <div class="mb-4">
      <label class="block text-sm text-muted mb-1">Output Directory</label>
      <div class="flex gap-2">
        <input v-model="outputPath" type="text" placeholder="/path/to/output" class="flex-1 bg-muted border border-border rounded-lg px-3 py-2 text-fg focus:outline-none focus:border-accent font-mono text-sm" />
        <button class="px-4 py-2 border border-border rounded-lg hover:bg-muted transition text-sm">Browse</button>
      </div>
    </div>

    <div class="border-t border-border pt-4">
      <div class="flex items-center justify-between mb-3">
        <h4 class="text-sm font-medium">Preview Transcode</h4>
        <button v-if="!showPreview" @click="showPreview = true" class="px-3 py-1.5 text-sm border border-border rounded-lg hover:bg-muted transition">Preview</button>
      </div>

      <div v-if="showPreview" class="bg-muted/30 rounded-lg p-4 space-y-3">
        <div>
          <label class="block text-xs text-muted mb-1">File</label>
          <select v-model="previewFile" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent">
            <option v-for="(f, i) in allFiles" :key="i" :value="i">{{ f.name || `File ${i + 1}` }}</option>
          </select>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-xs text-muted mb-1">Start Time</label>
            <input v-model="startTime" type="text" placeholder="00:00:00" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent font-mono" />
          </div>
          <div>
            <label class="block text-xs text-muted mb-1">Duration</label>
            <select v-model="duration" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent">
              <option value="30">30s</option>
              <option value="60">60s</option>
              <option value="90">90s</option>
              <option value="120">120s</option>
            </select>
          </div>
        </div>

        <div class="flex gap-2">
          <button v-if="!previewRunning && !previewDone" @click="startPreview" class="px-4 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition text-sm" :disabled="startingPreview">
            {{ startingPreview ? 'Starting...' : 'Start Preview' }}
          </button>
          <button v-if="previewRunning" @click="cancelPreview" class="px-4 py-2 border border-destructive text-destructive rounded-lg hover:bg-destructive/10 transition text-sm">Cancel</button>
          <a v-if="previewDone && downloadLink" :href="downloadLink" class="px-4 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition text-sm">Download Preview</a>
        </div>

        <div v-if="previewStatus" class="text-sm" :class="previewError ? 'text-destructive' : 'text-muted'">
          {{ previewError || previewStatus }}
        </div>
      </div>
    </div>

    <div v-if="error" class="text-destructive text-sm mt-4">{{ error }}</div>

    <div class="flex gap-3 pt-6">
      <button @click="$emit('back')" class="px-4 py-2 border border-border rounded-lg hover:bg-muted transition text-sm">Back</button>
      <button @click="submit" class="px-6 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition text-sm ml-auto" :disabled="submitting || !outputPath">
        {{ submitting ? 'Creating...' : 'Create Task' }}
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
const downloadLink = ref('')
const previewId = ref<number | null>(null)
const error = ref('')
const submitting = ref(false)

const allFiles = computed(() => {
  return [...(props.data.files || []), ...(props.data.sub_files || [])]
})

async function startPreview() {
  startingPreview.value = true
  previewError.value = ''
  previewStatus.value = ''
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
    previewError.value = e.message || 'Preview failed'
  } finally {
    startingPreview.value = false
  }
}

async function pollPreview() {
  if (!previewId.value) return
  try {
    const status = await api.preview.status(previewId.value)
    previewStatus.value = status.status || `Progress: ${status.progress || 0}%`
    if (status.status === 'completed' || status.progress >= 100) {
      previewRunning.value = false
      previewDone.value = true
      downloadLink.value = api.preview.downloadUrl(previewId.value!, status.token || '')
    } else if (status.status === 'failed') {
      previewRunning.value = false
      previewError.value = status.error || 'Preview failed'
    } else {
      setTimeout(pollPreview, 2000)
    }
  } catch {
    setTimeout(pollPreview, 2000)
  }
}

async function cancelPreview() {
  if (previewId.value) {
    try { await api.preview.delete(previewId.value) } catch { /* ignore */ }
  }
  previewRunning.value = false
  previewId.value = null
}

async function submit() {
  if (!outputPath.value) return
  submitting.value = true
  error.value = ''
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
    error.value = e.message || 'Failed to create task'
  } finally {
    submitting.value = false
  }
}
</script>
