import axios from 'axios'
import type { ApiError } from '../types'

const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

export const http = axios.create({
  baseURL,
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('helpdesk_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (res) => res,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('helpdesk_token')
      localStorage.removeItem('helpdesk_user')
      if (!window.location.pathname.startsWith('/login') && !window.location.pathname.startsWith('/register')) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

export function getErrorMessage(err: unknown, fallback = 'Something went wrong'): string {
  const ax = err as { response?: { data?: ApiError }; message?: string }
  return ax.response?.data?.error || ax.message || fallback
}
