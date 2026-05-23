// src/mocks/adapter.ts
import { MOCK_TASKS, MOCK_PROJECTS, MOCK_DASHBOARD, MOCK_USER } from './data'

const delay = (ms: number) => new Promise(res => setTimeout(res, ms + Math.random() * 500))

// Формируем ответ, структурно идентичный AxiosResponse
const toResponse = <T>(data: T, status = 200) => ({
  data,
  status,
  statusText: status === 200 ? 'OK' : 'Error',
  headers: {},
  config: {}
})

export const mockApi = {
  get: async (url: string) => {
    await delay(400)
    if (url.includes('/auth/profile')) return toResponse(MOCK_USER)
    if (url.includes('/tasks')) return toResponse(MOCK_TASKS)
    if (url.includes('/dashboard')) return toResponse(MOCK_DASHBOARD)
    if (url.includes('/projects')) return toResponse(MOCK_PROJECTS)
    if (url.includes('/users/me')) return toResponse(MOCK_USER)
    return toResponse(null)
  },

  post: async (url: string, data?: any) => {
    await delay(600)
    console.log('📥 [MOCK POST]', url, data)
    if (url.includes('/auth/login')) {
      return toResponse({
        token: 'fake-jwt-token',
        user: MOCK_USER
      })
    }
    if (url.includes('/sprints')) return toResponse({ id: 'sprint-new-1', ...data })
    return toResponse({ success: true })
  },

  patch: async (url: string, data?: any) => {
    await delay(300)
    console.log('📥 [MOCK PATCH]', url, data)
  
    // 👇 НОВОЕ: если обновляют профиль — возвращаем обновлённого пользователя
    if (url.includes('/auth/profile')) {
      // "Сливаем" старые данные с новыми из запроса
      const updatedUser = { ...MOCK_USER, ...data }
      return toResponse(updatedUser)
  }
  
  return toResponse({ success: true })
  },

  put: async (url: string, data?: any) => {
    await delay(300)
    console.log('📥 [MOCK PUT]', url, data)
    return toResponse({ success: true })
  },

  delete: async (url: string) => {
    await delay(300)
    console.log(' [MOCK DELETE]', url)
    return toResponse({ success: true })
  }
}
