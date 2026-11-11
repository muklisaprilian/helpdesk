import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { http, getErrorMessage } from '../api/http'
import type { AuthResponse, Role, User } from '../types'

const TOKEN_KEY = 'helpdesk_token'
const USER_KEY = 'helpdesk_user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<User | null>(readUser())
  const loading = ref(false)
  const error = ref('')

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isAgentOrAdmin = computed(() => user.value?.role === 'agent' || user.value?.role === 'admin')
  const role = computed<Role | null>(() => user.value?.role ?? null)

  function readUser(): User | null {
    const raw = localStorage.getItem(USER_KEY)
    if (!raw) return null
    try {
      return JSON.parse(raw) as User
    } catch {
      return null
    }
  }

  function persist(res: AuthResponse) {
    token.value = res.token
    user.value = res.user
    localStorage.setItem(TOKEN_KEY, res.token)
    localStorage.setItem(USER_KEY, JSON.stringify(res.user))
  }

  function clear() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  async function login(email: string, password: string) {
    loading.value = true
    error.value = ''
    try {
      const { data } = await http.post<AuthResponse>('/auth/login', { email, password })
      persist(data)
      return data.user
    } catch (err) {
      error.value = getErrorMessage(err, 'Login failed')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function register(name: string, email: string, password: string) {
    loading.value = true
    error.value = ''
    try {
      const { data } = await http.post<AuthResponse>('/auth/register', { name, email, password })
      persist(data)
      return data.user
    } catch (err) {
      error.value = getErrorMessage(err, 'Registration failed')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    if (!token.value) return null
    try {
      const { data } = await http.get<User>('/me')
      user.value = data
      localStorage.setItem(USER_KEY, JSON.stringify(data))
      return data
    } catch {
      clear()
      return null
    }
  }

  function logout() {
    clear()
  }

  return {
    token,
    user,
    loading,
    error,
    isAuthenticated,
    isAdmin,
    isAgentOrAdmin,
    role,
    login,
    register,
    fetchMe,
    logout,
  }
})
