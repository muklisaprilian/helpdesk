import type { Role, Status } from '../types'

const transitions: Record<string, Role[]> = {
  'open->in_progress': ['agent', 'admin'],
  'in_progress->resolved': ['agent', 'admin'],
  'resolved->closed': ['requester', 'agent', 'admin'],
  'resolved->open': ['requester'],
  'open->closed': ['requester', 'admin'],
}

export function allowedNextStatuses(current: Status, role: Role | null): Status[] {
  if (!role) return []
  const next: Status[] = []
  for (const [key, roles] of Object.entries(transitions)) {
    const [from, to] = key.split('->') as [Status, Status]
    if (from === current && roles.includes(role)) {
      next.push(to)
    }
  }
  return next
}

export function statusLabel(status: Status): string {
  return status.replace('_', ' ')
}
