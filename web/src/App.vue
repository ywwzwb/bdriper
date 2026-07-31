<template>
  <div style="min-height:100vh;background:linear-gradient(180deg,#0a0a0f 0%,#020203 100%);">
    <div class="blob" style="width:600px;height:600px;background:#5E6AD2;top:-200px;right:-100px;" />
    <div class="blob" style="width:400px;height:400px;background:#22C55E;bottom:-100px;left:-100px;opacity:0.05;" />
    <NavBar />
    <main style="position:relative;z-index:10;max-width:1200px;margin:0 auto;padding:24px;">
      <RouterView />
    </main>
  </div>
</template>
<script setup lang="ts">
import NavBar from './components/NavBar.vue'
</script>
