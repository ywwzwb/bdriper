<template>
  <nav class="flex items-center justify-between p-4 bg-card border-b border-border">
    <RouterLink to="/" class="text-accent font-semibold text-lg">BDRiper</RouterLink>
    <div class="flex gap-4">
      <RouterLink to="/" class="text-muted hover:text-fg">概览</RouterLink>
      <RouterLink to="/tasks" class="text-muted hover:text-fg">任务管理</RouterLink>
      <RouterLink to="/logs" class="text-muted hover:text-fg">日志</RouterLink>
      <RouterLink to="/settings" class="text-muted hover:text-fg">设置</RouterLink>
    </div>
  </nav>
</template>
