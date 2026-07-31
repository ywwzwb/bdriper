<template>
  <div>
    <h3 class="text-lg font-medium mb-1">Step 2: Select Content</h3>
    <p class="text-sm text-muted mb-4">{{ data.disc_name || data.source_path }}</p>

    <div class="space-y-4">
      <div>
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium">Main Files</span>
          <button @click="toggleAllMain" class="text-xs text-accent hover:underline">{{ allMainSelected ? 'Deselect all' : 'Select all' }}</button>
        </div>
        <div class="space-y-1 max-h-48 overflow-y-auto">
          <label v-for="(f, i) in data.files" :key="i" class="flex items-center gap-2 p-2 rounded hover:bg-muted/50 cursor-pointer">
            <input type="checkbox" v-model="selectedFiles[i]" class="w-4 h-4 rounded border-border bg-muted accent-accent" />
            <span class="flex-1 text-sm">{{ f.name || f }}</span>
            <span v-if="f.duration" class="text-xs text-muted">{{ f.duration }}</span>
            <span v-if="f.resolution" class="text-xs text-muted ml-2">{{ f.resolution }}</span>
          </label>
        </div>
      </div>

      <div v-if="data.sub_files?.length" class="border-t border-border pt-3">
        <button @click="showSub = !showSub" class="flex items-center gap-1 text-sm text-muted hover:text-fg">
          <span class="transition-transform" :class="showSub ? 'rotate-90' : ''">&#9654;</span> Sub Files ({{ data.sub_files.length }})
        </button>
        <div v-if="showSub" class="space-y-1 mt-2 max-h-32 overflow-y-auto">
          <div class="flex items-center justify-between mb-1">
            <span class="text-xs text-muted"></span>
            <button @click="toggleAllSub" class="text-xs text-accent hover:underline">{{ allSubSelected ? 'Deselect all' : 'Select all' }}</button>
          </div>
          <label v-for="(f, i) in data.sub_files" :key="i" class="flex items-center gap-2 p-2 rounded hover:bg-muted/50 cursor-pointer">
            <input type="checkbox" v-model="selectedSub[i]" class="w-4 h-4 rounded border-border bg-muted accent-accent" />
            <span class="text-sm">{{ f.name || f }}</span>
          </label>
        </div>
      </div>

      <div class="border-t border-border pt-3">
        <label class="block text-sm text-muted mb-1">Audio Tracks</label>
        <select multiple v-model="selectedAudio" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent min-h-[80px]">
          <option v-for="(t, i) in data.audio_tracks" :key="i" :value="i">{{ t.lang || `Track ${i + 1}` }} {{ t.codec ? `(${t.codec})` : '' }}</option>
        </select>
      </div>

      <div class="border-t border-border pt-3">
        <label class="block text-sm text-muted mb-1">Subtitle Tracks</label>
        <select multiple v-model="selectedSubtitles" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent min-h-[80px]">
          <option v-for="(t, i) in data.subtitle_tracks" :key="i" :value="i">{{ t.lang || `Track ${i + 1}` }}</option>
        </select>
      </div>

      <div class="border-t border-border pt-3">
        <label class="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" v-model="includeChapters" class="w-4 h-4 rounded border-border bg-muted accent-accent" />
          <span class="text-sm">Include Chapters</span>
        </label>
      </div>

      <div v-if="error" class="text-destructive text-sm">{{ error }}</div>

      <div class="flex gap-3 pt-4">
        <button @click="$emit('back')" class="px-4 py-2 border border-border rounded-lg hover:bg-muted transition text-sm">Back</button>
        <button @click="next" class="px-6 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition text-sm ml-auto" :disabled="!hasSelection">Next</button>
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
  if (!hasSelection.value) {
    error.value = 'Select at least one file'
    return
  }
  emit('next', {
    files: props.data.files?.filter((_: any, i: number) => selectedFiles[i]) || [],
    sub_files: props.data.sub_files?.filter((_: any, i: number) => selectedSub[i]) || [],
    audio_tracks: selectedAudio.value,
    subtitle_tracks: selectedSubtitles.value,
    include_chapters: includeChapters.value,
  })
}
</script>
