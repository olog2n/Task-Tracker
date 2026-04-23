import { defineStore } from 'pinia'
import api from '@/services/api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    isAuthenticated: false
  }),

  actions: {
    async login(email, password) {
      const response = await api.post('/auth/login', { email, password })
      this.isAuthenticated = true
      this.user = response.data.user
      return response.data
    },

    async register(email, password) {
      const response = await api.post('/auth/register', { email, password })
      this.isAuthenticated = true
      this.user = response.data.user
      return response.data
    },

    async logout() {
      await api.post('/auth/logout')
      this.isAuthenticated = false
      this.user = null
    }
  }
})
