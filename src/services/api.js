import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:6969/api',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json'
  }
})

// UUID валидация (для отладки)
export function isValidUUID(uuid) {
  const UUID_REGEX = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
  return UUID_REGEX.test(uuid)
}

// Интерцептор для добавления токена
api.interceptors.request.use(config => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Интерцептор для обработки ошибок
api.interceptors.response.use(
  response => {
    console.log('✅ Response received:', response.status)
    return response
  },
  error => {
    if (error.message.includes('Network Error')) {
      console.error('🚫 CORS Error detected!')
      console.error('🚫 Check backend CORS configuration')
    }

    if (error.response?.status === 401) {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/login'
    }

    return Promise.reject(error)
  }
)

export default api
