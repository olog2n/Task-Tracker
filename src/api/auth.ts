// src/api/auth.ts
import api from './client'

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  email: string
  password: string
  name: string
}

// 🔧 Расширяем интерфейс пользователя
export interface AuthUser {
  id: string
  name: string
  email?: string
  default_project_id?: string // 🔧 Опционально, т.к. может не вернуться при ошибке
}

// 🔧 Интерфейс ответа от сервера
export interface AuthResponse {
  token: string
  user: AuthUser
}

export const authApi = {
  login: (data: LoginPayload) => api.post<AuthResponse>('/auth/login', data),
  register: (data: RegisterPayload) => api.post<AuthResponse>('/auth/register', data),
}
