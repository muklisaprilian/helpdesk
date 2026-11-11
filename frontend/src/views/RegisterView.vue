<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AppShell from '../components/AppShell.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const form = reactive({ name: '', email: '', password: '' })
const clientError = ref('')

async function onSubmit() {
  clientError.value = ''
  if (form.name.trim().length < 2) {
    clientError.value = 'Name must be at least 2 characters.'
    return
  }
  if (form.password.length < 8) {
    clientError.value = 'Password must be at least 8 characters.'
    return
  }
  try {
    await auth.register(form.name.trim(), form.email, form.password)
    router.push('/tickets')
  } catch {
    // store error
  }
}
</script>

<template>
  <AppShell>
    <section class="auth-card">
      <h1>Create account</h1>
      <p class="muted">New accounts are registered as requesters.</p>
      <form class="form" @submit.prevent="onSubmit">
        <label>
          Name
          <input v-model="form.name" type="text" required minlength="2" />
        </label>
        <label>
          Email
          <input v-model="form.email" type="email" required />
        </label>
        <label>
          Password
          <input v-model="form.password" type="password" required minlength="8" />
        </label>
        <p v-if="clientError || auth.error" class="form-error">{{ clientError || auth.error }}</p>
        <button class="btn btn-primary" type="submit" :disabled="auth.loading">
          {{ auth.loading ? 'Creating…' : 'Register' }}
        </button>
      </form>
      <p class="muted">Already have an account? <RouterLink to="/login">Sign in</RouterLink></p>
    </section>
  </AppShell>
</template>
