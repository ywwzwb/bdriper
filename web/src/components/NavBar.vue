<template>
  <nav class="flex items-center justify-between px-6 py-3 bg-card/80 backdrop-blur border-b border-border/50">
    <RouterLink to="/" class="flex items-center gap-2.5 group">
      <div class="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
        <svg class="w-5 h-5 text-black" fill="none" stroke="currentColor" viewBox="0 0 24 24"><polygon points="12,2 22,8.5 22,15.5 12,22 2,15.5 2,8.5"/></svg>
      </div>
      <span class="text-fg font-bold text-lg tracking-tight">BDRiper</span>
    </RouterLink>

    <div class="flex gap-1 bg-muted/50 rounded-xl p-1">
      <RouterLink
        v-for="link in links"
        :key="link.to"
        :to="link.to"
        class="relative px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200"
        :class="route.path === link.to ? 'text-fg bg-card shadow-sm' : 'text-muted hover:text-fg'"
        active-class=""
        :exact="link.to === '/'"
      >
        {{ link.label }}
      </RouterLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()

const links = [
  { to: '/', label: '概览' },
  { to: '/tasks', label: '任务管理' },
  { to: '/logs', label: '日志' },
  { to: '/settings', label: '设置' },
]
</script>
