import axios from 'axios'
import router from '/@router'

const api = axios.create({
  baseURL: '/api/v0', // 🔧 Проверь префикс, если бэкенд меняет версию
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

// Request Interceptor: инжект токена
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response Interceptor: обработка ошибок
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 🔧 Очистка сессии и редирект
      localStorage.removeItem('auth_token')
      router.push('/login')
    }
    // Можно добавить глобальный toast для 403/500
    return Promise.reject(error)
  }
)

export default api
