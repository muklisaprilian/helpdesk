<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppShell from '../components/AppShell.vue'
import StateBlock from '../components/StateBlock.vue'
import { getErrorMessage, http } from '../api/http'
import type { Role, User } from '../types'

const loading = ref(false)
const error = ref('')
const users = ref<User[]>([])
const savingId = ref<string | null>(null)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await http.get<User[]>('/users')
    users.value = data
  } catch (err) {
    error.value = getErrorMessage(err, 'Failed to load users')
  } finally {
    loading.value = false
  }
}

async function changeRole(user: User, role: Role) {
  savingId.value = user.id
  error.value = ''
  try {
    const { data } = await http.patch<User>(`/users/${user.id}/role`, { role })
    users.value = users.value.map((u) => (u.id === data.id ? data : u))
  } catch (err) {
    error.value = getErrorMessage(err, 'Failed to update role')
  } finally {
    savingId.value = null
  }
}

onMounted(load)
</script>

<template>
  <AppShell>
    <section class="page">
      <div class="page-header">
        <div>
          <h1>Users</h1>
          <p class="muted">Admin only — manage roles</p>
        </div>
      </div>

      <StateBlock :loading="loading" :error="error" :empty="!users.length">
        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Role</th>
                <th>Joined</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id">
                <td>{{ u.name }}</td>
                <td>{{ u.email }}</td>
                <td>
                  <select
                    :value="u.role"
                    :disabled="savingId === u.id"
                    @change="changeRole(u, ($event.target as HTMLSelectElement).value as Role)"
                  >
                    <option value="requester">requester</option>
                    <option value="agent">agent</option>
                    <option value="admin">admin</option>
                  </select>
                </td>
                <td>{{ new Date(u.created_at).toLocaleDateString() }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </StateBlock>
    </section>
  </AppShell>
</template>
