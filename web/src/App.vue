<template>
  <div class="min-h-screen bg-bg text-fg font-sans">
    <NavBar />
    <main class="p-6">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import NavBar from './components/NavBar.vue'
import { connect } from './ws'

onMounted(() => {
  connect()
})
</script>
