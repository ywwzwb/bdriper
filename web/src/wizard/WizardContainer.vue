<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" @click.self="$emit('close')">
    <div class="bg-card border border-border/50 rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-auto shadow-2xl">
      <div class="flex items-center justify-between px-6 py-4 border-b border-border/30">
        <div>
          <h2 class="text-lg font-bold text-fg">新建转码任务</h2>
          <p class="text-xs text-muted mt-0.5">步骤 {{ step }} / 4</p>
        </div>
        <button @click="$emit('close')" class="p-1.5 rounded-lg hover:bg-muted text-muted hover:text-fg transition-all duration-200">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <div class="flex items-center justify-center gap-0 px-6 py-4">
        <template v-for="i in 4" :key="i">
          <div class="flex items-center">
            <div class="flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold transition-all duration-300" :class="i < step ? 'bg-accent text-black' : i === step ? 'bg-accent text-black ring-2 ring-accent/30' : 'bg-muted text-muted'">
              <svg v-if="i < step" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><polyline points="20,6 9,17 4,12"/></svg>
              <span v-else>{{ i }}</span>
            </div>
          </div>
          <div v-if="i < 4" class="w-12 h-0.5 rounded-full transition-all duration-300" :class="i < step ? 'bg-accent' : 'bg-muted'" />
        </template>
      </div>

      <div class="px-6 pb-6">
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

function onFilesSelected(data: any) { Object.assign(wizardData, data); step.value = 3 }
function onConfigSelected(data: any) { wizardData.config_id = data.config_id; wizardData.new_config = data.new_config; step.value = 4 }
function onDone() { emit('close') }

onUnmounted(() => {
  if (wizardData.preview_id) {
    import('@/api').then(m => m.api.preview.delete(wizardData.preview_id).catch(() => {}))
  }
})
</script>
