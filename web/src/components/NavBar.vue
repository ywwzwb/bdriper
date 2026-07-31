<template>
  <nav class="flex items-center justify-between p-4 bg-card border-b border-border">
    <RouterLink to="/" class="text-accent font-semibold text-lg">BDRiper</RouterLink>
    <div class="flex gap-4">
      <RouterLink to="/" class="text-muted hover:text-fg">Overview</RouterLink>
      <RouterLink to="/tasks" class="text-muted hover:text-fg">Tasks</RouterLink>
      <RouterLink to="/logs" class="text-muted hover:text-fg">Logs</RouterLink>
      <RouterLink to="/settings" class="text-muted hover:text-fg">Settings</RouterLink>
    </div>
  </nav>
</template>
