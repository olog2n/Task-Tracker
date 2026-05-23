// src/api/auth.ts
import { A } from 'vue-router/dist/router-CWoNjPRp.mjs'
import api,  { ApiResponse } from './client'
import apiClient from '@/api/client'

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
  login: string
  email?: string
  default_project_id?: string // 🔧 Опционально, т.к. может не вернуться при ошибке
  surname?: string
  department?: string
  avatar?: string
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

// Обновления профиля
export interface UpdateProfilePayload {
  name?: string
  surname?: string
  department?: string
  password?: string
  passwordConfirm?: string
}

// Получение текущего пользователя
export async function getProfile(): Promise<ApiResponse<AuthUser>> {
  return apiClient.get<AuthUser>('/api/v0/auth/profile')
}

// Обновление профиля
export async function updateProfile(
  payload: UpdateProfilePayload
): Promise<ApiResponse<AuthUser>> {
  const cleanPayload = { ...payload }
  if (!cleanPayload.password) {
    delete cleanPayload.password
    delete cleanPayload.passwordConfirm
  }
  return apiClient.patch<AuthUser>('/api/v0/auth/profile', cleanPayload)  
}