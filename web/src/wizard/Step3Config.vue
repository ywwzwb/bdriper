<template>
  <div>
    <div class="mb-6">
      <h3 class="text-lg font-bold text-fg mb-1">编码配置</h3>
      <p class="text-sm text-muted">选择已有配置或创建新配置</p>
    </div>

    <div class="space-y-2 mb-4 max-h-56 overflow-y-auto">
      <div v-if="configs.length === 0 && !saveConfig" class="text-muted text-sm text-center py-4">暂无保存的配置</div>
      <label v-for="c in configs" :key="c.id" class="flex items-start gap-3 p-4 rounded-xl border cursor-pointer transition-all duration-200" :class="selectedConfigId === c.id ? 'border-accent bg-accent/5' : 'border-border/50 hover:border-muted'">
        <input type="radio" :value="c.id" v-model="selectedConfigId" class="mt-0.5 accent-accent cursor-pointer" />
        <div>
          <div class="font-medium text-sm text-fg">{{ c.name }}</div>
          <div class="text-xs text-muted mt-0.5">{{ c.encoder }} &middot; {{ c.mode || '简易' }}</div>
        </div>
      </label>
      <label class="flex items-start gap-3 p-4 rounded-xl border cursor-pointer transition-all duration-200" :class="selectedConfigId === 0 ? 'border-accent bg-accent/5' : 'border-border/50 hover:border-muted'">
        <input type="radio" :value="0" v-model="selectedConfigId" class="mt-0.5 accent-accent cursor-pointer" />
        <div>
          <div class="font-medium text-sm text-fg">新建配置</div>
          <div class="text-xs text-muted mt-0.5">创建新的编码参数配置</div>
        </div>
      </label>
    </div>

    <div v-if="selectedConfigId === 0" class="border-t border-border/30 pt-4">
      <ConfigEditor :config="null" @saved="onNewConfigSaved" />
    </div>

    <div v-if="error" class="bg-destructive/10 border border-destructive/30 text-destructive rounded-lg px-4 py-3 text-sm mt-4">{{ error }}</div>

    <div class="flex gap-3 pt-4">
      <button @click="$emit('back')" class="px-5 py-2.5 border border-border/50 rounded-lg hover:bg-muted transition-all duration-200 text-sm font-medium">上一步</button>
      <button @click="next" class="px-6 py-2.5 bg-accent text-black font-semibold rounded-lg hover:brightness-110 transition-all duration-200 text-sm ml-auto shadow-lg shadow-accent/20 disabled:opacity-50" :disabled="!canProceed">下一步</button>
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
const saveConfig = ref<any>(null)
const error = ref('')

const canProceed = computed(() => {
  if (selectedConfigId.value === 0) return !!saveConfig.value
  return !!selectedConfigId.value
})

onMounted(async () => {
  try {
    configs.value = await api.configs.list()
    if (configs.value.length > 0) selectedConfigId.value = configs.value[0].id
  } catch { /* ignore */ }
})

function onNewConfigSaved(config: any) { saveConfig.value = config }

function next() {
  if (!canProceed.value) { error.value = '请选择或创建编码配置'; return }
  emit('next', {
    config_id: selectedConfigId.value === 0 ? null : selectedConfigId.value,
    new_config: selectedConfigId.value === 0 ? saveConfig.value : null,
  })
}
</script>
