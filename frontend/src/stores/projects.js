import { defineStore } from 'pinia'
import { projectService } from '@/services/projectService'

export const useProjectsStore = defineStore('projects', {
  state: () => ({
    projects: [],
    currentProject: null
  }),

  actions: {
    async fetchProjects() {
      this.projects = await projectService.getProjects()
    },

    async fetchProjectById(id) {
      this.currentProject = await projectService.getProjectById(id)
    },

    async createProject(project) {
      const newProject = await projectService.createProject(project)
      this.projects.push(newProject)
      return newProject
    },

    async updateProject(id, updates) {
      const updated = await projectService.updateProject(id, updates)
      const index = this.projects.findIndex(p => p.id === id)
      if (index !== -1) {
        this.projects[index] = updated
      }
      return updated
    },

    async deleteProject(id) {
      await projectService.deleteProject(id)
      this.projects = this.projects.filter(p => p.id !== id)
    }
  }
})
