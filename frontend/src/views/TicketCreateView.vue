<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppShell from '../components/AppShell.vue'
import { getErrorMessage, http } from '../api/http'
import type { Priority, Ticket } from '../types'

const router = useRouter()
const loading = ref(false)
const error = ref('')
const success = ref('')
const form = reactive({
  title: '',
  description: '',
  category: '',
  priority: 'medium' as Priority,
})

async function onSubmit() {
  error.value = ''
  success.value = ''
  if (form.title.trim().length < 3) {
    error.value = 'Title must be at least 3 characters.'
    return
  }
  if (form.description.trim().length < 3) {
    error.value = 'Description must be at least 3 characters.'
    return
  }
  if (form.category.trim().length < 2) {
    error.value = 'Category is required.'
    return
  }

  loading.value = true
  try {
    const { data } = await http.post<Ticket>('/tickets', form)
    success.value = 'Ticket created.'
    router.push(`/tickets/${data.id}`)
  } catch (err) {
    error.value = getErrorMessage(err, 'Failed to create ticket')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AppShell>
    <section class="page narrow">
      <h1>Create ticket</h1>
      <form class="form" @submit.prevent="onSubmit">
        <label>
          Title
          <input v-model="form.title" type="text" required minlength="3" maxlength="200" />
        </label>
        <label>
          Category
          <input v-model="form.category" type="text" required placeholder="e.g. Network, Hardware" />
        </label>
        <label>
          Priority
          <select v-model="form.priority">
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="urgent">Urgent</option>
          </select>
        </label>
        <label>
          Description
          <textarea v-model="form.description" rows="6" required minlength="3" />
        </label>
        <p v-if="error" class="form-error">{{ error }}</p>
        <p v-if="success" class="form-success">{{ success }}</p>
        <button class="btn btn-primary" type="submit" :disabled="loading">
          {{ loading ? 'Submitting…' : 'Create ticket' }}
        </button>
      </form>
    </section>
  </AppShell>
</template>
