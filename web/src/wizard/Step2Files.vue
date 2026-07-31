<template>
  <div>
    <div class="mb-6">
      <h3 class="text-lg font-bold text-fg mb-1">选择内容</h3>
      <p class="text-sm text-muted font-mono truncate">{{ data.disc_name || data.source_path }}</p>
    </div>

    <div class="space-y-4">
      <div class="bg-muted/30 rounded-xl p-4">
        <div class="flex items-center justify-between mb-3">
          <span class="text-sm font-medium text-fg">主文件</span>
          <button @click="toggleAllMain" class="text-xs text-accent hover:underline">{{ allMainSelected ? '取消全选' : '全选' }}</button>
        </div>
        <div class="space-y-1 max-h-48 overflow-y-auto">
          <label v-for="(f, i) in data.files" :key="i" class="flex items-center gap-3 p-2.5 rounded-lg hover:bg-muted/50 cursor-pointer transition-all duration-150">
            <input type="checkbox" v-model="selectedFiles[i]" class="w-4 h-4 rounded border-border bg-muted accent-accent cursor-pointer" />
            <span class="flex-1 text-sm">{{ f.name || f }}</span>
            <span v-if="f.duration" class="text-xs text-muted">{{ f.duration }}</span>
            <span v-if="f.resolution" class="text-xs text-muted px-1.5 py-0.5 rounded bg-muted">{{ f.resolution }}</span>
          </label>
        </div>
      </div>

      <div v-if="data.sub_files?.length" class="bg-muted/30 rounded-xl p-4">
        <button @click="showSub = !showSub" class="flex items-center gap-2 text-sm font-medium text-fg hover:text-accent transition-colors duration-200">
          <svg class="w-4 h-4 transition-transform duration-200" :class="showSub ? 'rotate-90' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24"><polyline points="9,18 15,12 9,6"/></svg>
          子文件（{{ data.sub_files.length }}）
        </button>
        <div v-if="showSub" class="mt-3 space-y-1 max-h-32 overflow-y-auto">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-muted"></span>
            <button @click="toggleAllSub" class="text-xs text-accent hover:underline">{{ allSubSelected ? '取消全选' : '全选' }}</button>
          </div>
          <label v-for="(f, i) in data.sub_files" :key="i" class="flex items-center gap-3 p-2.5 rounded-lg hover:bg-muted/50 cursor-pointer transition-all duration-150">
            <input type="checkbox" v-model="selectedSub[i]" class="w-4 h-4 rounded border-border bg-muted accent-accent cursor-pointer" />
            <span class="text-sm">{{ f.name || f }}</span>
          </label>
        </div>
      </div>

      <div class="bg-muted/30 rounded-xl p-4">
        <label class="block text-sm font-medium text-fg mb-2">音频轨道</label>
        <select multiple v-model="selectedAudio" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg text-sm focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200 min-h-[80px]">
          <option v-for="(t, i) in data.audio_tracks" :key="i" :value="i">{{ t.lang || `Track ${i + 1}` }} {{ t.codec ? `(${t.codec})` : '' }}</option>
        </select>
      </div>

      <div class="bg-muted/30 rounded-xl p-4">
        <label class="block text-sm font-medium text-fg mb-2">字幕轨道</label>
        <select multiple v-model="selectedSubtitles" class="w-full bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg text-sm focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200 min-h-[80px]">
          <option v-for="(t, i) in data.subtitle_tracks" :key="i" :value="i">{{ t.lang || `Track ${i + 1}` }}</option>
        </select>
      </div>

      <div class="bg-muted/30 rounded-xl p-4">
        <label class="flex items-center gap-3 cursor-pointer">
          <input type="checkbox" v-model="includeChapters" class="w-4 h-4 rounded border-border bg-muted accent-accent cursor-pointer" />
          <span class="text-sm text-fg">包含章节信息</span>
        </label>
      </div>

      <div v-if="error" class="bg-destructive/10 border border-destructive/30 text-destructive rounded-lg px-4 py-3 text-sm">{{ error }}</div>

      <div class="flex gap-3 pt-2">
        <button @click="$emit('back')" class="px-5 py-2.5 border border-border/50 rounded-lg hover:bg-muted transition-all duration-200 text-sm font-medium">上一步</button>
        <button @click="next" class="px-6 py-2.5 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 text-sm ml-auto shadow-lg shadow-accent/20 disabled:opacity-50" :disabled="!hasSelection">下一步</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'

const props = defineProps<{ data: any }>()
const emit = defineEmits<{ next: [data: any], back: [] }>()

const selectedFiles = reactive<Record<number, boolean>>({})
const selectedSub = reactive<Record<number, boolean>>({})
const selectedAudio = ref<number[]>([])
const selectedSubtitles = ref<number[]>([])
const includeChapters = ref(true)
const showSub = ref(false)
const error = ref('')

props.data.files?.forEach((_: any, i: number) => (selectedFiles[i] = true))
props.data.sub_files?.forEach((_: any, i: number) => (selectedSub[i] = true))

const allMainSelected = computed(() => {
  const files = props.data.files || []
  return files.length > 0 && files.every((_: any, i: number) => selectedFiles[i])
})

const allSubSelected = computed(() => {
  const subs = props.data.sub_files || []
  return subs.length > 0 && subs.every((_: any, i: number) => selectedSub[i])
})

const hasSelection = computed(() => {
  return Object.values(selectedFiles).some(Boolean) || Object.values(selectedSub).some(Boolean)
})

function toggleAllMain() {
  const val = !allMainSelected.value
  ;(props.data.files || []).forEach((_: any, i: number) => (selectedFiles[i] = val))
}

function toggleAllSub() {
  const val = !allSubSelected.value
  ;(props.data.sub_files || []).forEach((_: any, i: number) => (selectedSub[i] = val))
}

function next() {
  if (!hasSelection.value) { error.value = '请至少选择一个文件'; return }
  emit('next', {
    files: props.data.files?.filter((_: any, i: number) => selectedFiles[i]) || [],
    sub_files: props.data.sub_files?.filter((_: any, i: number) => selectedSub[i]) || [],
    audio_tracks: selectedAudio.value,
    subtitle_tracks: selectedSubtitles.value,
    include_chapters: includeChapters.value,
  })
}
</script>
