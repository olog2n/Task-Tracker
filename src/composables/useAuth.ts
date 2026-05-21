// src/composables/useAuth.ts
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authApi, AuthUser } from '@/api/auth' // 🔧 Импортируем тип

const token = ref<string | null>(localStorage.getItem('auth_token'))
const user = ref<AuthUser | null>(JSON.parse(localStorage.getItem('auth_user') || 'null'))

export function useAuth() {
  const router = useRouter()
  const isAuthenticated = computed(() => !!token.value)

  const setSession = (newToken: string, userData: AuthUser) => {
    token.value = newToken
    user.value = userData
    localStorage.setItem('auth_token', newToken)
    localStorage.setItem('auth_user', JSON.stringify(userData))
  }

  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('auth_token')
    localStorage.removeItem('auth_user')
    router.push('/login')
  }

  const login = async (email: string, password: string) => {
    const { data } = await authApi.login({ email, password })
    setSession(data.token, data.user)

    const redirect = router.currentRoute.value.query.redirect as string
    // 🔧 Теперь TS знает, что default_project_id может существовать
    const targetProject = data.user.default_project_id || 'proj-1'

    router.push(redirect || `/projects/${targetProject}/dashboard`)
  }

  const register = async (name: string, email: string, password: string) => {
    const { data } = await authApi.register({ name, email, password })
    setSession(data.token, data.user)

    const targetProject = data.user.default_project_id || 'proj-1'
    router.push(`/projects/${targetProject}/dashboard`)
  }

  return { token, user, isAuthenticated, login, register, logout }
}
