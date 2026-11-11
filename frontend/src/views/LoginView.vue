<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import AppShell from '../components/AppShell.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const form = reactive({ email: '', password: '' })
const clientError = ref('')

async function onSubmit() {
  clientError.value = ''
  if (!form.email || !form.password) {
    clientError.value = 'Email and password are required.'
    return
  }
  try {
    await auth.login(form.email, form.password)
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/tickets'
    router.push(redirect)
  } catch {
    // error shown from store
  }
}
</script>

<template>
  <AppShell>
    <section class="auth-card">
      <h1>Sign in</h1>
      <p class="muted">Internal helpdesk for employees and agents.</p>
      <form class="form" @submit.prevent="onSubmit">
        <label>
          Email
          <input v-model="form.email" type="email" autocomplete="username" required />
        </label>
        <label>
          Password
          <input v-model="form.password" type="password" autocomplete="current-password" required />
        </label>
        <p v-if="clientError || auth.error" class="form-error">{{ clientError || auth.error }}</p>
        <button class="btn btn-primary" type="submit" :disabled="auth.loading">
          {{ auth.loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
      <p class="muted">No account? <RouterLink to="/register">Register</RouterLink></p>
    </section>
  </AppShell>
</template>
