<template>
  <router-view />
</template>

<script setup lang="ts">
import { useSettingsStore } from '@/stores/settings'
import { onMounted } from 'vue'

const settings = useSettingsStore()

onMounted(() => {
  settings.loadSettings()
})
</script>
