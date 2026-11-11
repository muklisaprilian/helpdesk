export type Role = 'requester' | 'agent' | 'admin'
export type Priority = 'low' | 'medium' | 'high' | 'urgent'
export type Status = 'open' | 'in_progress' | 'resolved' | 'closed'

export interface User {
  id: string
  name: string
  email: string
  role: Role
  created_at: string
}

export interface UserBrief {
  id: string
  name: string
  email: string
  role: Role
}

export interface Comment {
  id: string
  ticket_id: string
  author_id: string
  body: string
  is_internal: boolean
  created_at: string
  author?: UserBrief
}

export interface Ticket {
  id: string
  title: string
  description: string
  category: string
  priority: Priority
  status: Status
  requester_id: string
  assignee_id: string | null
  created_at: string
  updated_at: string
  requester?: UserBrief
  assignee?: UserBrief | null
  comments?: Comment[]
}

export interface PaginatedTickets {
  data: Ticket[]
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  error: string
  details?: string
}
