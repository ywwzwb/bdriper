<template>
  <div>
    <div class="mb-6">
      <h3 class="text-lg font-bold text-fg mb-1">选择源文件</h3>
      <p class="text-sm text-muted">指定 BDMV 文件夹或 ISO 映像文件路径</p>
    </div>

    <div class="mb-4">
      <label class="block text-sm font-medium text-fg mb-2">源文件路径</label>
      <div class="flex gap-2">
        <input v-model="path" type="text" placeholder="/path/to/BDMV or .iso" class="flex-1 bg-muted border border-border/50 rounded-lg px-4 py-2.5 text-fg font-mono text-sm focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/30 transition-all duration-200" @keyup.enter="parse" />
        <button @click="parse" class="px-5 py-2.5 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 disabled:opacity-50 shadow-lg shadow-accent/20 flex items-center gap-2" :disabled="!path || parsing">
          <svg v-if="parsing" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/></svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
          {{ parsing ? '解析中...' : '浏览' }}
        </button>
      </div>
      <p class="text-xs text-muted mt-2">支持 BDMV 文件夹和 ISO 文件</p>
    </div>

    <div v-if="error" class="bg-destructive/10 border border-destructive/30 text-destructive rounded-lg px-4 py-3 text-sm">{{ error }}</div>
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
  parsing.value = true; error.value = ''
  try {
    const result = await api.wizard.parse(path.value)
    emit('next', { source_path: path.value, ...result })
  } catch (e: any) {
    error.value = e.message || '解析源文件失败'
  } finally {
    parsing.value = false
  }
}
</script>
