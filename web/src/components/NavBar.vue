<template>
  <nav style="position:sticky;top:0;z-index:50;backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);background:rgba(2,2,3,0.6);border-bottom:1px solid rgba(255,255,255,0.06);">
    <div style="max-width:1200px;margin:0 auto;padding:0 24px;display:flex;align-items:center;height:56px;">
      <router-link to="/" style="display:flex;align-items:center;gap:8px;text-decoration:none;margin-right:40px;">
        <div style="width:8px;height:8px;border-radius:50%;background:#5E6AD2;box-shadow:0 0 12px rgba(94,106,210,0.6);" />
        <span style="font-weight:700;font-size:16px;color:#EDEDEF;letter-spacing:-0.5px;">BDRiper</span>
      </router-link>
      <div style="display:flex;gap:4px;">
        <router-link v-for="tab in tabs" :key="tab.path" :to="tab.path" class="pill" :class="{ active: $route.path === tab.path }">
          {{ tab.name }}
        </router-link>
      </div>
    </div>
  </nav>
</template>
<script setup lang="ts">
const tabs = [
  { name: '概览', path: '/' },
  { name: '任务管理', path: '/tasks' },
  { name: '日志', path: '/logs' },
  { name: '设置', path: '/settings' },
]
</script>
