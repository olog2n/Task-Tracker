import { defineStore } from 'pinia'
import { taskService } from '@/services/taskService'

export const useTasksStore = defineStore('tasks', {
  state: () => ({
    tasks: [],
    currentTask: null,
    total: 0,
    limit: 50,
    offset: 0
  }),

  actions: {
    async fetchTasks(filters = {}) {
      const result = await taskService.getTasks({
        ...filters,
        limit: this.limit,
        offset: this.offset
      })

      this.tasks = result.tasks || result
      this.total = result.total || this.tasks.length
      return this.tasks
    },

    async fetchTaskById(id) {
      this.currentTask = await taskService.getTaskById(id)
      return this.currentTask
    },

    async createTask(task) {
      const newTask = await taskService.createTask(task)
      this.tasks.unshift(newTask)
      return newTask
    },

    async updateTask(id, updates) {
      const updated = await taskService.updateTask(id, updates)
      const index = this.tasks.findIndex(t => t.id === id)
      if (index !== -1) {
        this.tasks[index] = updated
      }
      return updated
    },

    async deleteTask(id) {
      await taskService.deleteTask(id)
      this.tasks = this.tasks.filter(t => t.id !== id)
    }
  }
})
