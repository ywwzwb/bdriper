<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60" @click.self="$emit('close')">
    <div class="bg-bg border border-border rounded-xl w-full max-w-2xl max-h-[90vh] overflow-auto shadow-2xl">
      <div class="flex items-center justify-between p-4 border-b border-border">
        <h2 class="text-lg font-semibold">New Task</h2>
        <button @click="$emit('close')" class="p-1 rounded hover:bg-muted text-muted hover:text-fg">
          <span class="text-xl leading-none">&times;</span>
        </button>
      </div>

      <div class="flex gap-1 p-4 pb-0 justify-center">
        <button v-for="i in 4" :key="i" class="w-2.5 h-2.5 rounded-full transition-colors" :class="i <= step ? 'bg-accent' : 'bg-muted'" />
      </div>

      <div class="p-6">
        <Step1Source v-if="step === 1" @next="onSourceSelected" />
        <Step2Files v-else-if="step === 2" :data="wizardData" @next="onFilesSelected" @back="step = 1" />
        <Step3Config v-else-if="step === 3" :data="wizardData" @next="onConfigSelected" @back="step = 2" />
        <Step4Target v-else-if="step === 4" :data="wizardData" @back="step = 3" @done="onDone" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onUnmounted } from 'vue'
import Step1Source from './Step1Source.vue'
import Step2Files from './Step2Files.vue'
import Step3Config from './Step3Config.vue'
import Step4Target from './Step4Target.vue'

const emit = defineEmits<{ close: [] }>()

const step = ref(1)
const wizardData = reactive<any>({
  source_path: '',
  disc_name: '',
  files: [],
  sub_files: [],
  audio_tracks: [],
  subtitle_tracks: [],
  include_chapters: true,
  config_id: null,
  new_config: null,
  output_path: '',
  preview_id: null,
})

function onSourceSelected(data: any) {
  wizardData.source_path = data.source_path
  wizardData.disc_name = data.disc_name
  wizardData.files = data.files || []
  wizardData.sub_files = data.sub_files || []
  wizardData.audio_tracks = data.audio_tracks || []
  wizardData.subtitle_tracks = data.subtitle_tracks || []
  step.value = 2
}

function onFilesSelected(data: any) {
  Object.assign(wizardData, data)
  step.value = 3
}

function onConfigSelected(data: any) {
  wizardData.config_id = data.config_id
  wizardData.new_config = data.new_config
  step.value = 4
}

function onDone() {
  emit('close')
}

onUnmounted(() => {
  if (wizardData.preview_id) {
    import('@/api').then(m => m.api.preview.delete(wizardData.preview_id).catch(() => {}))
  }
})
</script>
