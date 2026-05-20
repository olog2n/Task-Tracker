import api from './client'

export const tasksApi = {
  list: (projectId: string, params?: Record<string, any>) =>
    api.get(`/tasks`, { params: { project_id: projectId, ...params } }),

  create: (projectId: string, data: any) =>
    api.post(`/tasks`, { ...data, project_id: projectId }),

  update: (taskId: string, data: any) =>
    api.patch(`/tasks/${taskId}`, data),

  transition: (taskId: string, targetStatusId: string, reason?: string) =>
    api.post(`/tasks/${taskId}/transitions`, { target_status_id: targetStatusId, reason }),
}
