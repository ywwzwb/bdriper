<template>
  <div>
    <h3 class="text-lg font-medium mb-4">Step 1: Source</h3>

    <div class="mb-4">
      <label class="block text-sm text-muted mb-1">Source Path</label>
      <div class="flex gap-2">
        <input v-model="path" type="text" placeholder="/path/to/BDMV or .iso" class="flex-1 bg-muted border border-border rounded-lg px-3 py-2 text-fg focus:outline-none focus:border-accent font-mono text-sm" @keyup.enter="parse" />
        <button @click="parse" class="px-4 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition disabled:opacity-50" :disabled="!path || parsing">
          {{ parsing ? 'Parsing...' : 'Browse' }}
        </button>
      </div>
      <p class="text-xs text-muted mt-1">Supports BDMV folders and ISO files</p>
    </div>

    <div v-if="error" class="text-destructive text-sm mb-4">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/api'

const emit = defineEmits<{ next: [data: any] }>()

const path = ref('')
const parsing = ref(false)
const error = ref('')

async function parse() {
  parsing.value = true
  error.value = ''
  try {
    const result = await api.wizard.parse(path.value)
    emit('next', { source_path: path.value, ...result })
  } catch (e: any) {
    error.value = e.message || 'Failed to parse source'
  } finally {
    parsing.value = false
  }
}
</script>
