<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import AppShell from '../components/AppShell.vue'
import StateBlock from '../components/StateBlock.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { getErrorMessage, http } from '../api/http'
import type { PaginatedTickets, Ticket } from '../types'

const loading = ref(false)
const error = ref('')
const tickets = ref<Ticket[]>([])
const totalPages = ref(1)
const total = ref(0)

const filters = reactive({
  search: '',
  status: '',
  priority: '',
  page: 1,
  per_page: 10,
  sort_by: 'created_at',
  sort_dir: 'desc',
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await http.get<PaginatedTickets>('/tickets', { params: { ...filters } })
    tickets.value = data.data
    totalPages.value = data.total_pages
    total.value = data.total
    filters.page = data.page
  } catch (err) {
    error.value = getErrorMessage(err, 'Failed to load tickets')
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  filters.page = 1
  load()
}

function changePage(delta: number) {
  filters.page = Math.min(Math.max(1, filters.page + delta), totalPages.value)
  load()
}

onMounted(load)
</script>

<template>
  <AppShell>
    <section class="page">
      <div class="page-header">
        <div>
          <h1>Tickets</h1>
          <p class="muted">{{ total }} ticket(s)</p>
        </div>
        <RouterLink class="btn btn-primary" to="/tickets/new">New ticket</RouterLink>
      </div>

      <form class="filters" @submit.prevent="applyFilters">
        <input v-model="filters.search" type="search" placeholder="Search title…" />
        <select v-model="filters.status">
          <option value="">All statuses</option>
          <option value="open">Open</option>
          <option value="in_progress">In progress</option>
          <option value="resolved">Resolved</option>
          <option value="closed">Closed</option>
        </select>
        <select v-model="filters.priority">
          <option value="">All priorities</option>
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
          <option value="urgent">Urgent</option>
        </select>
        <select v-model="filters.sort_by">
          <option value="created_at">Created</option>
          <option value="updated_at">Updated</option>
          <option value="priority">Priority</option>
          <option value="status">Status</option>
          <option value="title">Title</option>
        </select>
        <select v-model="filters.sort_dir">
          <option value="desc">Desc</option>
          <option value="asc">Asc</option>
        </select>
        <button class="btn btn-secondary" type="submit">Apply</button>
      </form>

      <StateBlock :loading="loading" :error="error" :empty="!tickets.length" empty-text="No tickets found.">
        <div class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>Title</th>
                <th>Status</th>
                <th>Priority</th>
                <th>Category</th>
                <th>Requester</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in tickets" :key="t.id">
                <td>
                  <RouterLink :to="`/tickets/${t.id}`">{{ t.title }}</RouterLink>
                </td>
                <td><StatusBadge :status="t.status" /></td>
                <td><StatusBadge :priority="t.priority" /></td>
                <td>{{ t.category }}</td>
                <td>{{ t.requester?.name || '—' }}</td>
                <td>{{ new Date(t.updated_at).toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="pagination">
          <button class="btn btn-ghost" type="button" :disabled="filters.page <= 1" @click="changePage(-1)">
            Previous
          </button>
          <span>Page {{ filters.page }} / {{ totalPages }}</span>
          <button
            class="btn btn-ghost"
            type="button"
            :disabled="filters.page >= totalPages"
            @click="changePage(1)"
          >
            Next
          </button>
        </div>
      </StateBlock>
    </section>
  </AppShell>
</template>
