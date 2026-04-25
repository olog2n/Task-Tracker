import api from './api'

export const projectService = {
  async getProjects() {
    const response = await api.get('/projects')
    return response.data
  },

  async getProjectById(id) {
    const response = await api.get(`/projects/${id}`)
    return response.data
  },

  async createProject(project) {
    const response = await api.post('/projects', {
      name: project.name,
      description: project.description || ''
    })
    return response.data
  },

  async updateProject(id, updates) {
    const response = await api.patch(`/projects/${id}`, updates)
    return response.data
  },

  async deleteProject(id) {
    await api.delete(`/projects/${id}`)
  },

  async getMembers(projectId) {
    const response = await api.get(`/projects/${projectId}/members`)
    return response.data
  },

  async addMember(projectId, userId, role) {
    const response = await api.post(`/projects/${projectId}/members`, {
      user_id: userId,
      role: role || 'member'
    })
    return response.data
  },

  async removeMember(projectId, userId) {
    await api.delete(`/projects/${projectId}/members/${userId}`)
  }
}
