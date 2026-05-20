import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authApi } from '@/api/auth'

const token = ref<string | null>(localStorage.getItem('auth_token'))
const user = ref<{ id: string; name: string } | null>(JSON.parse(localStorage.getItem('auth_user') || 'null'))

export function useAuth() {
  const router = useRouter()
  const isAuthenticated = computed(() => !!token.value)

  const setSession = (newToken: string, userData: any) => {
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
    //  Редирект на сохранённый путь или дашборд
    const redirect = router.currentRoute.value.query.redirect as string
    router.push(redirect || `/projects/${data.user.default_project_id}/dashboard`)
  }

  const register = async (name: string, email: string, password: string) => {
    const { data } = await authApi.register({ name, email, password })
    setSession(data.token, data.user)
    router.push(`/projects/${data.user.default_project_id}/dashboard`)
  }

  return { token, user, isAuthenticated, login, register, logout }
}
