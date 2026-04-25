import api from './api'

export const authService = {
  async register(email, password) {
    console.log('📤 Register request:', { email, password })
    const response = await api.post('/auth/register', { email, password })
    console.log('📥 Register response:', response.data)  // 👈 Добавь это
    return response.data
  },

  async login(email, password) {
  console.log('📤 Login request:', { email, password })
  const response = await api.post('/auth/login', { email, password })

  console.log('📥 Login response:', response.data)  // 👈 Посмотри что приходит!
  console.log('🔑 access_token:', response.data.access_token)
  console.log('🔑 refresh_token:', response.data.refresh_token)

  // 👇 Адаптируй под формат бэкенда
  const accessToken = response.data.access_token
  const refreshToken = response.data.refresh_token

  if (accessToken) {
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)
    console.log('✅ Tokens saved')
  } else {
    console.error('❌ No token in response!')
    console.error('Response keys:', Object.keys(response.data))
  }

  return response.data
  },

  async getCurrentUser() {
    const response = await api.get('/auth/me')
    console.log('📥 Me response:', response.data)
    return response.data
  }
}
