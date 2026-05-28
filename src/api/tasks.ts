import api from './client'
import type { Task, TaskCreatePayload, ApiResponse } from '@/types'

export const tasksApi = {
  list: (projectId: string, params?: Record<string, any>) =>
    api.get<ApiResponse<Task[]>>(`/tasks`, { 
      params: { project_id: projectId, ...params } 
    }),

  // ✅ СОЗДАНИЕ — строго по спецификации API V0
  create: (payload: TaskCreatePayload) =>
    api.post<ApiResponse<Task>>(`/tasks`, payload),

  update: (taskId: string, data: Partial<Task>) =>
    api.patch<ApiResponse<Task>>(`/tasks/${taskId}`, data),

  transition: (taskId: string, targetStatusId: string, reason?: string) =>
    api.post<ApiResponse<Task>>(`/tasks/${taskId}/transitions`, { 
      target_status_id: targetStatusId, 
      reason 
    }),
}