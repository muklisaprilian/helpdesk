<script setup lang="ts">
import type { Priority, Status } from '../types'
import { statusLabel } from '../composables/useTicketTransitions'

defineProps<{
  status?: Status
  priority?: Priority
}>()
</script>

<template>
  <span v-if="status" class="badge" :class="`status-${status}`">{{ statusLabel(status) }}</span>
  <span v-else-if="priority" class="badge" :class="`priority-${priority}`">{{ priority }}</span>
</template>
