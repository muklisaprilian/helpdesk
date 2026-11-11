<script setup lang="ts">
defineProps<{
  loading?: boolean
  error?: string
  empty?: boolean
  emptyText?: string
}>()
</script>

<template>
  <div v-if="loading" class="state-block">Loading…</div>
  <div v-else-if="error" class="state-block state-error">{{ error }}</div>
  <div v-else-if="empty" class="state-block">{{ emptyText || 'Nothing here yet.' }}</div>
  <slot v-else />
</template>
