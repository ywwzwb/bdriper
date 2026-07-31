<template>
  <div>
    <h3 class="text-lg font-medium mb-4">Step 3: Encoding Config</h3>

    <div class="space-y-2 mb-4 max-h-48 overflow-y-auto">
      <div v-if="configs.length === 0" class="text-muted text-sm">No saved configs</div>
      <label v-for="c in configs" :key="c.id" class="flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors" :class="selectedConfigId === c.id ? 'border-accent bg-accent/5' : 'border-border hover:border-muted'">
        <input type="radio" :value="c.id" v-model="selectedConfigId" class="mt-0.5 accent-accent" />
        <div>
          <div class="font-medium text-sm">{{ c.name }}</div>
          <div class="text-xs text-muted">{{ c.encoder }} &middot; {{ c.mode || 'simple' }}</div>
        </div>
      </label>
      <label class="flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors" :class="selectedConfigId === 0 ? 'border-accent bg-accent/5' : 'border-border hover:border-muted'">
        <input type="radio" :value="0" v-model="selectedConfigId" class="mt-0.5 accent-accent" />
        <div>
          <div class="font-medium text-sm">New config</div>
          <div class="text-xs text-muted">Create a new encoding configuration</div>
        </div>
      </label>
    </div>

    <div v-if="selectedConfigId === 0" class="border-t border-border pt-4">
      <ConfigEditor :config="null" @saved="onNewConfigSaved" />
    </div>

    <div v-if="error" class="text-destructive text-sm mb-4">{{ error }}</div>

    <div class="flex gap-3 pt-4">
      <button @click="$emit('back')" class="px-4 py-2 border border-border rounded-lg hover:bg-muted transition text-sm">Back</button>
      <button @click="next" class="px-6 py-2 bg-accent text-black font-medium rounded-lg hover:brightness-110 transition text-sm ml-auto" :disabled="!canProceed">Next</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api } from '@/api'
import ConfigEditor from '@/config/ConfigEditor.vue'

defineProps<{ data: any }>()
const emit = defineEmits<{ next: [data: any], back: [] }>()

const configs = ref<any[]>([])
const selectedConfigId = ref<number | null>(null)
const newConfig = ref<any>(null)
const error = ref('')

const canProceed = computed(() => {
  if (selectedConfigId.value === 0) return !!newConfig.value
  return !!selectedConfigId.value
})

onMounted(async () => {
  try {
    configs.value = await api.configs.list()
    if (configs.value.length > 0) selectedConfigId.value = configs.value[0].id
  } catch { /* ignore */ }
})

function onNewConfigSaved(config: any) {
  newConfig.value = config
}

function next() {
  if (!canProceed.value) {
    error.value = 'Select or create a config'
    return
  }
  emit('next', {
    config_id: selectedConfigId.value === 0 ? null : selectedConfigId.value,
    new_config: selectedConfigId.value === 0 ? newConfig.value : null,
  })
}
</script>
