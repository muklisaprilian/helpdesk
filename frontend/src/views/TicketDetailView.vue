<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import AppShell from '../components/AppShell.vue'
import StateBlock from '../components/StateBlock.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { allowedNextStatuses, statusLabel } from '../composables/useTicketTransitions'
import { getErrorMessage, http } from '../api/http'
import { useAuthStore } from '../stores/auth'
import type { Priority, Status, Ticket, User } from '../types'

const route = useRoute()
const auth = useAuthStore()

const loading = ref(false)
const saving = ref(false)
const error = ref('')
const actionError = ref('')
const ticket = ref<Ticket | null>(null)
const agents = ref<User[]>([])
const commentBody = ref('')
const commentInternal = ref(false)
const selectedPriority = ref<Priority>('medium')
const selectedAssignee = ref('')

const nextStatuses = computed(() =>
  ticket.value && auth.role ? allowedNextStatuses(ticket.value.status, auth.role) : [],
)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const id = route.params.id as string
    const { data } = await http.get<Ticket>(`/tickets/${id}`)
    ticket.value = data
    selectedPriority.value = data.priority
    selectedAssignee.value = data.assignee_id || ''
    if (auth.isAdmin) {
      const usersRes = await http.get<User[]>('/users')
      agents.value = usersRes.data.filter((u) => u.role === 'agent' || u.role === 'admin')
    } else if (auth.isAgentOrAdmin) {
      agents.value = auth.user ? [auth.user] : []
    }
  } catch (err) {
    error.value = getErrorMessage(err, 'Failed to load ticket')
  } finally {
    loading.value = false
  }
}

async function changeStatus(status: Status) {
  if (!ticket.value) return
  saving.value = true
  actionError.value = ''
  try {
    const { data } = await http.patch<Ticket>(`/tickets/${ticket.value.id}`, { status })
    ticket.value = data
  } catch (err) {
    actionError.value = getErrorMessage(err, 'Status update failed')
  } finally {
    saving.value = false
  }
}

async function claimTicket() {
  if (!ticket.value || !auth.user) return
  saving.value = true
  actionError.value = ''
  try {
    const { data } = await http.patch<Ticket>(`/tickets/${ticket.value.id}`, {
      assignee_id: auth.user.id,
    })
    ticket.value = data
    selectedAssignee.value = data.assignee_id || ''
  } catch (err) {
    actionError.value = getErrorMessage(err, 'Claim failed')
  } finally {
    saving.value = false
  }
}

async function saveMeta() {
  if (!ticket.value || !auth.isAgentOrAdmin) return
  saving.value = true
  actionError.value = ''
  try {
    const payload: { priority: Priority; assignee_id?: string } = {
      priority: selectedPriority.value,
    }
    if (selectedAssignee.value) {
      payload.assignee_id = selectedAssignee.value
    }
    const { data } = await http.patch<Ticket>(`/tickets/${ticket.value.id}`, payload)
    ticket.value = data
  } catch (err) {
    actionError.value = getErrorMessage(err, 'Update failed')
  } finally {
    saving.value = false
  }
}

async function addComment() {
  if (!ticket.value || !commentBody.value.trim()) return
  saving.value = true
  actionError.value = ''
  try {
    await http.post(`/tickets/${ticket.value.id}/comments`, {
      body: commentBody.value.trim(),
      is_internal: commentInternal.value,
    })
    commentBody.value = ''
    commentInternal.value = false
    await load()
  } catch (err) {
    actionError.value = getErrorMessage(err, 'Comment failed')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <AppShell>
    <section class="page">
      <StateBlock :loading="loading" :error="error" :empty="!ticket">
        <template v-if="ticket">
          <div class="page-header">
            <div>
              <p class="muted">#{{ ticket.id.slice(0, 8) }}</p>
              <h1>{{ ticket.title }}</h1>
              <div class="badge-row">
                <StatusBadge :status="ticket.status" />
                <StatusBadge :priority="ticket.priority" />
              </div>
            </div>
          </div>

          <div class="detail-grid">
            <article class="panel">
              <h2>Details</h2>
              <p>{{ ticket.description }}</p>
              <dl class="meta-list">
                <div><dt>Category</dt><dd>{{ ticket.category }}</dd></div>
                <div><dt>Requester</dt><dd>{{ ticket.requester?.name }}</dd></div>
                <div><dt>Assignee</dt><dd>{{ ticket.assignee?.name || 'Unassigned' }}</dd></div>
                <div><dt>Created</dt><dd>{{ new Date(ticket.created_at).toLocaleString() }}</dd></div>
                <div><dt>Updated</dt><dd>{{ new Date(ticket.updated_at).toLocaleString() }}</dd></div>
              </dl>
            </article>

            <aside class="panel">
              <h2>Actions</h2>
              <p v-if="actionError" class="form-error">{{ actionError }}</p>

              <div v-if="nextStatuses.length" class="action-group">
                <p class="muted">Change status</p>
                <div class="btn-row">
                  <button
                    v-for="s in nextStatuses"
                    :key="s"
                    class="btn btn-secondary"
                    type="button"
                    :disabled="saving"
                    @click="changeStatus(s)"
                  >
                    {{ statusLabel(s) }}
                  </button>
                </div>
              </div>

              <div v-if="auth.isAgentOrAdmin" class="action-group">
                <button class="btn btn-primary" type="button" :disabled="saving" @click="claimTicket">
                  Assign to me
                </button>
                <label>
                  Priority
                  <select v-model="selectedPriority">
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                    <option value="urgent">Urgent</option>
                  </select>
                </label>
                <label v-if="auth.isAdmin">
                  Assignee
                  <select v-model="selectedAssignee">
                    <option value="">Unassigned</option>
                    <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name }} ({{ a.role }})</option>
                  </select>
                </label>
                <button class="btn btn-secondary" type="button" :disabled="saving" @click="saveMeta">
                  Save changes
                </button>
              </div>
            </aside>
          </div>

          <section class="panel">
            <h2>Comments</h2>
            <div v-if="!ticket.comments?.length" class="muted">No comments yet.</div>
            <ul class="comment-list">
              <li v-for="c in ticket.comments" :key="c.id" :class="{ internal: c.is_internal }">
                <div class="comment-head">
                  <strong>{{ c.author?.name || 'User' }}</strong>
                  <span class="muted">{{ new Date(c.created_at).toLocaleString() }}</span>
                  <span v-if="c.is_internal" class="badge priority-high">internal</span>
                </div>
                <p>{{ c.body }}</p>
              </li>
            </ul>

            <form class="form" @submit.prevent="addComment">
              <label>
                Add comment
                <textarea v-model="commentBody" rows="3" required />
              </label>
              <label v-if="auth.isAgentOrAdmin" class="checkbox">
                <input v-model="commentInternal" type="checkbox" />
                Internal note (hidden from requester)
              </label>
              <button class="btn btn-primary" type="submit" :disabled="saving || !commentBody.trim()">
                Post comment
              </button>
            </form>
          </section>
        </template>
      </StateBlock>
    </section>
  </AppShell>
</template>
