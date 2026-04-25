import api from './api'

export const taskService = {
  async getTasks(filters = {}) {
    const params = new URLSearchParams()

    if (filters.project_id) params.append('project_id', filters.project_id)
    if (filters.status_id) params.append('status_id', filters.status_id)
    if (filters.assignee_id) params.append('assignee_id', filters.assignee_id)
    if (filters.priority) params.append('priority', filters.priority)
    if (filters.search) params.append('search', filters.search)
    if (filters.limit) params.append('limit', filters.limit)
    if (filters.offset) params.append('offset', filters.offset)

    const response = await api.get(`/tasks?${params.toString()}`)
    return response.data
  },

  async getTaskById(id) {
    const response = await api.get(`/tasks/${id}`)
    return response.data
  },

  async createTask(task) {
    const response = await api.post('/tasks', {
      title: task.title,
      description: task.description || '',
      project_id: task.project_id,
      status_id: task.status_id,
      priority: task.priority || 'medium',
      assignee_id: task.assignee_id || null
    })
    return response.data
  },

  async updateTask(id, updates) {
    const response = await api.patch(`/tasks/${id}`, updates)
    return response.data
  },

  async deleteTask(id) {
    await api.delete(`/tasks/${id}`)
  }
}
