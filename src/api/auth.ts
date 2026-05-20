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

export const authApi = {
  login: (data: LoginPayload) => api.post<{ token: string; user: { id: string; name: string } }>('/auth/login', data),
  register: (data: RegisterPayload) => api.post<{ token: string; user: { id: string; name: string } }>('/auth/register', data),
}
