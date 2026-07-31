<template>
  <div class="bg-card border border-border rounded-lg p-4 space-y-4 mt-2">
    <div class="flex items-center justify-between">
      <h4 class="font-medium text-sm">{{ config ? 'Edit Config' : 'New Config' }}</h4>
      <button v-if="!embedded" @click="$emit('close')" class="text-muted hover:text-fg text-lg leading-none">&times;</button>
    </div>

    <div>
      <label class="block text-xs text-muted mb-1">Name</label>
      <input v-model="name" type="text" placeholder="My Config" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent" />
    </div>

    <div>
      <label class="block text-xs text-muted mb-1">Encoder</label>
      <select v-model="encoder" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent">
        <option value="x264">x264</option>
        <option value="x265">x265</option>
        <option value="nvenc">NVENC</option>
        <option value="qsv">QSV</option>
        <option value="amf">AMF</option>
      </select>
    </div>

    <div class="flex gap-2">
      <button @click="mode = 'simple'" class="flex-1 px-3 py-1.5 rounded text-sm transition" :class="mode === 'simple' ? 'bg-accent text-black' : 'bg-muted text-muted'">Simple</button>
      <button @click="mode = 'professional'" class="flex-1 px-3 py-1.5 rounded text-sm transition" :class="mode === 'professional' ? 'bg-accent text-black' : 'bg-muted text-muted'">Professional</button>
    </div>

    <div v-if="mode === 'simple'" class="space-y-4">
      <div>
        <label class="block text-xs text-muted mb-1">Quality</label>
        <div class="flex gap-2">
          <button v-for="q in qualities" :key="q.value" @click="quality = q.value" class="flex-1 px-2 py-1.5 rounded text-xs transition" :class="quality === q.value ? 'bg-accent text-black' : 'bg-muted text-muted'">{{ q.label }}</button>
        </div>
      </div>

      <div>
        <label class="block text-xs text-muted mb-1">Speed</label>
        <select v-model="speed" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent">
          <option value="ultrafast">Ultra Fast</option>
          <option value="veryfast">Very Fast</option>
          <option value="fast">Fast</option>
          <option value="medium">Medium</option>
          <option value="slow">Slow</option>
          <option value="veryslow">Very Slow</option>
          <option value="placebo">Placebo</option>
        </select>
      </div>

      <div>
        <label class="block text-xs text-muted mb-1">Bit Depth</label>
        <select v-model="bitDepth" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent">
          <option value="8">8-bit</option>
          <option value="10">10-bit</option>
          <option value="12">12-bit</option>
        </select>
      </div>
    </div>

    <div v-else class="space-y-4">
      <div v-for="group in professionalGroups" :key="group.name" class="border border-border rounded-lg p-3">
        <div class="text-xs font-medium text-muted mb-2">{{ group.name }}</div>
        <div class="grid grid-cols-2 gap-3">
          <div v-for="param in group.params" :key="param.key">
            <label class="block text-xs text-muted mb-0.5">{{ param.label }}</label>
            <div v-if="param.options" class="w-full">
              <select v-model="professionalParams[param.key]" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent">
                <option v-for="opt in param.options" :key="opt" :value="opt">{{ opt }}</option>
              </select>
            </div>
            <div v-else class="w-full">
              <input v-model="professionalParams[param.key]" type="text" :placeholder="param.placeholder || ''" class="w-full bg-muted border border-border rounded-lg px-3 py-2 text-fg text-sm focus:outline-none focus:border-accent" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="flex gap-2 pt-2">
      <button v-if="embedded" @click="save" class="px-4 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition text-sm" :disabled="saving">{{ saving ? 'Saving...' : 'Save Config' }}</button>
      <button v-else @click="save" class="px-4 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition text-sm" :disabled="saving">{{ saving ? 'Saving...' : 'Save' }}</button>
      <button v-if="!embedded" class="px-4 py-2 border border-border rounded-lg hover:bg-muted transition text-sm" @click="$emit('close')">Cancel</button>
    </div>

    <div v-if="saveError" class="text-destructive text-xs">{{ saveError }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { api } from '@/api'

const props = defineProps<{ config: any }>()
const emit = defineEmits<{ close: [], saved: [config: any] }>()

const embedded = ref(false)

const name = ref('')
const encoder = ref('x264')
const mode = ref('simple')
const quality = ref('medium')
const speed = ref('medium')
const bitDepth = ref('8')
const saving = ref(false)
const saveError = ref('')

const qualities = [
  { label: 'Low', value: 'low' },
  { label: 'Med', value: 'medium' },
  { label: 'High', value: 'high' },
  { label: 'Lossless', value: 'lossless' },
]

const x264Params: Record<string, string> = {
  crf: '23',
  preset: 'medium',
  tune: '',
  profile: '',
  ref: '',
  bframes: '',
  qcomp: '',
}

const x265Params: Record<string, string> = {
  crf: '28',
  preset: 'medium',
  tune: '',
  profile: '',
  ref: '',
  bframes: '',
  qcomp: '',
  'no-sao': '',
  'aq-mode': '',
}

const professionalParams = reactive<Record<string, string>>({})

const professionalGroups = computed(() => {
  const params = encoder.value === 'x265' ? x265Params : x264Params
  const groups: { name: string; params: { key: string; label: string; placeholder?: string; options?: string[] }[] }[] = [
    {
      name: 'Rate Control',
      params: [
        { key: 'crf', label: 'CRF', placeholder: '0-51' },
      ],
    },
    {
      name: 'Preset & Tuning',
      params: [
        { key: 'preset', label: 'Preset', options: ['ultrafast', 'veryfast', 'fast', 'medium', 'slow', 'veryslow', 'placebo'] },
        { key: 'tune', label: 'Tune', placeholder: 'film/animation/grain' },
      ],
    },
    {
      name: 'Profile & Level',
      params: [
        { key: 'profile', label: 'Profile', placeholder: 'high/main/baseline' },
      ],
    },
    {
      name: 'Advanced',
      params: [
        { key: 'ref', label: 'Reference Frames', placeholder: '1-16' },
        { key: 'bframes', label: 'B-Frames', placeholder: '0-16' },
        { key: 'qcomp', label: 'qComp', placeholder: '0-1.0' },
      ],
    },
  ]
  if (encoder.value === 'x265') {
    groups[3].params.push(
      { key: 'no-sao', label: 'No SAO', placeholder: '0 or 1' },
      { key: 'aq-mode', label: 'AQ Mode', placeholder: '0-3' },
    )
  }
  return groups
})

onMounted(() => {
  if (props.config) {
    name.value = props.config.name || ''
    encoder.value = props.config.encoder || 'x264'
    mode.value = props.config.mode || 'simple'
    quality.value = props.config.quality || 'medium'
    speed.value = props.config.preset || 'medium'
    bitDepth.value = props.config.bit_depth?.toString() || '8'
    if (props.config.params) {
      Object.assign(professionalParams, props.config.params)
    }
  } else {
    embedded.value = true
  }
})

function buildData() {
  return {
    name: name.value,
    encoder: encoder.value,
    mode: mode.value,
    quality: mode.value === 'simple' ? quality.value : undefined,
    preset: mode.value === 'simple' ? speed.value : undefined,
    bit_depth: mode.value === 'simple' ? parseInt(bitDepth.value) : undefined,
    params: mode.value === 'professional' ? { ...professionalParams } : undefined,
  }
}

async function save() {
  saving.value = true
  saveError.value = ''
  const data = buildData()
  try {
    let result: any
    if (props.config?.id) {
      result = await api.configs.update(props.config.id, data)
    } else {
      result = await api.configs.create(data)
    }
    emit('saved', result)
  } catch (e: any) {
    saveError.value = e.message || 'Failed to save'
  } finally {
    saving.value = false
  }
}
</script>
