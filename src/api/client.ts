import axios, { AxiosResponse, AxiosRequestConfig } from 'axios'
// import router from '@/router'
import { mockApi } from '@/mocks/adapter'

const USE_MOCKS = true

export type ApiResponse<T = any> = {
  data: T
  status: number
  statusText: string
  headers: Record<string, string>
  config?: any
}

const api = {
  get: <T>(url: string, config?: AxiosRequestConfig) =>
    USE_MOCKS ? (mockApi.get(url) as Promise<ApiResponse<T>>) : axios.get<T>(url, config),

  post: <T>(url: string, data?: any, config?: AxiosRequestConfig) =>
    USE_MOCKS ? (mockApi.post(url, data) as Promise<ApiResponse<T>>) : axios.post<T>(url, data, config),

  patch: <T>(url: string, data?: any, config?: AxiosRequestConfig) =>
    USE_MOCKS ? (mockApi.patch(url, data) as Promise<ApiResponse<T>>) : axios.patch<T>(url, data, config),

  put: <T>(url: string, data?: any, config?: AxiosRequestConfig) =>
    USE_MOCKS ? (mockApi.put(url, data) as Promise<ApiResponse<T>>) : axios.put<T>(url, data, config),

  delete: <T>(url: string, config?: AxiosRequestConfig) =>
    USE_MOCKS ? (mockApi.delete(url) as Promise<ApiResponse<T>>) : axios.delete<T>(url, config),

  interceptors: axios.interceptors
}

// const api = axios.create({
//   get: (url: string, config?: any) => USE_MOCKS ? mockApi.get(url) : axios.get(url, config),
//   post: (url: string, data?: any, config?: any) => USE_MOCKS ? mockApi.post(url, data) : axios.post(url, data, config),
//   patch: (url: string, data?: any, config?: any) => USE_MOCKS ? mockApi.patch(url, data) : axios.patch(url, data, config),
//   put: (url: string, data?: any, config?: any) => USE_MOCKS ? mockApi.put(url, data) : axios.put(url, data, config),
//   delete: (url: string, config?: any) => USE_MOCKS ? mockApi.delete(url) : axios.delete(url, config),
//   interceptors: axios.interceptors

//   // baseURL: '/api/v0', // 🔧 Проверь префикс, если бэкенд меняет версию
//   // timeout: 10000,
//   // headers: { 'Content-Type': 'application/json' },
// })

if (USE_MOCKS && !localStorage.getItem('auth_token')) {
  localStorage.setItem('auth_token', 'mock-token')
  localStorage.setItem('auth_user', JSON.stringify({ id: 'u-1', name: 'Dev Mode User' }))
}

// Request Interceptor: инжект токена
// api.interceptors.request.use((config) => {
//   const token = localStorage.getItem('auth_token')
//   if (token) {
//     config.headers.Authorization = `Bearer ${token}`
//   }
//   return config
// })

// Response Interceptor: обработка ошибок
// api.interceptors.response.use(
//   (response) => response,
//   (error) => {
//     if (error.response?.status === 401) {
//       // 🔧 Очистка сессии и редирект
//       localStorage.removeItem('auth_token')
//       router.push('/login')
//     }
//     // Можно добавить глобальный toast для 403/500
//     return Promise.reject(error)
//   }
// )

export default api
