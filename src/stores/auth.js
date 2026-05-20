import { defineStore } from 'pinia'
import { authService } from '@/services/authService'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    accessToken: localStorage.getItem('access_token'),
    refreshToken: localStorage.getItem('refresh_token')
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken,
    userId: (state) => state.user?.id || null,
    userEmail: (state) => state.user?.email || null
  },

  actions: {
    async login(email, password) {
      const data = await authService.login(email, password)
      this.accessToken = data.access_token
      this.refreshToken = data.refresh_token
      await this.fetchUser()
    },

    async register(email, password) {
      await authService.register(email, password)
    },

    async logout() {
      await authService.logout()
      this.user = null
      this.accessToken = null
      this.refreshToken = null
    },

    async fetchUser() {
      try {
        this.user = await authService.getCurrentUser()
      } catch (error) {
        this.user = null
      }
    }
  }
})
