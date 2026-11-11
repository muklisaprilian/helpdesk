<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="app-shell">
    <header class="topbar">
      <div class="brand">
        <RouterLink to="/tickets">Helpdesk</RouterLink>
      </div>
      <nav v-if="auth.isAuthenticated" class="nav">
        <RouterLink to="/tickets">Tickets</RouterLink>
        <RouterLink to="/tickets/new">New ticket</RouterLink>
        <RouterLink v-if="auth.isAdmin" to="/users">Users</RouterLink>
      </nav>
      <div v-if="auth.user" class="session">
        <span class="session-meta">{{ auth.user.name }} · {{ auth.user.role }}</span>
        <button class="btn btn-ghost" type="button" @click="logout">Logout</button>
      </div>
    </header>
    <main class="main">
      <slot />
    </main>
  </div>
</template>
