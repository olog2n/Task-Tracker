<template>
  <div class="min-h-screen bg-gray-100">
    <!-- Header -->
    <nav class="bg-white shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex items-center">
            <h1 class="text-xl font-bold text-gray-900">Tracker</h1>
          </div>
          <div class="flex items-center">
            <span class="text-gray-600 mr-4">{{ authStore.user?.email || 'User' }}</span>
            <button
              @click="handleLogout"
              class="text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              Выйти
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main content -->
    <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div class="px-4 py-6 sm:px-0">
        <!-- Create Task Form -->
        <div class="bg-white rounded-lg shadow p-6 mb-6">
          <h2 class="text-xl font-bold mb-4">Создать задачу</h2>
          <form @submit.prevent="createTask" class="flex gap-4">
            <input
              v-model="newTask.title"
              type="text"
              placeholder="Название задачи"
              class="flex-1 px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
            <select
              v-model="newTask.priority"
              class="px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="low">Низкий</option>
              <option value="medium">Средний</option>
              <option value="high">Высокий</option>
            </select>
            <button
              type="submit"
              class="bg-blue-500 text-white px-6 py-2 rounded-lg hover:bg-blue-600 transition"
              :disabled="loading"
            >
              {{ loading ? '...' : 'Создать' }}
            </button>
          </form>
        </div>

        <!-- Tasks List -->
        <div class="bg-white rounded-lg shadow p-6">
          <h2 class="text-xl font-bold mb-4">Задачи</h2>

          <div v-if="loadingTasks" class="text-center py-8">
            <p class="text-gray-500">Загрузка...</p>
          </div>

          <div v-else-if="tasks.length === 0" class="text-center py-8">
            <p class="text-gray-500">Нет задач. Создайте первую!</p>
          </div>

          <div v-else class="space-y-4">
            <div
              v-for="task in tasks"
              :key="task.id"
              class="border rounded-lg p-4 hover:shadow-md transition"
            >
              <div class="flex justify-between items-start">
                <div>
                  <h3 class="font-semibold text-lg">{{ task.title }}</h3>
                  <p class="text-gray-600 text-sm mt-1">{{ task.description || 'Нет описания' }}</p>
                  <div class="mt-2 flex gap-2">
                    <span
                      :class="{
                        'bg-green-100 text-green-800': task.status === 'done',
                        'bg-yellow-100 text-yellow-800': task.status === 'in_progress',
                        'bg-gray-100 text-gray-800': task.status === 'todo'
                      }"
                      class="px-2 py-1 rounded text-xs font-medium"
                    >
                      {{ statusLabel(task.status) }}
                    </span>
                    <span
                      :class="{
                        'bg-green-100 text-green-800': task.priority === 'high',
                        'bg-yellow-100 text-yellow-800': task.priority === 'medium',
                        'bg-gray-100 text-gray-800': task.priority === 'low'
                      }"
                      class="px-2 py-1 rounded text-xs font-medium"
                    >
                      {{ priorityLabel(task.priority) }}
                    </span>
                  </div>
                </div>
                <div class="flex gap-2">
                  <button
                    @click="toggleTaskStatus(task)"
                    class="text-blue-500 hover:text-blue-700 text-sm"
                  >
                    {{ task.status === 'done' ? 'Вернуть' : 'Готово' }}
                  </button>
                  <button
                    @click="deleteTask(task.id)"
                    class="text-red-500 hover:text-red-700 text-sm"
                  >
                    Удалить
                  </button>
                </div>
              </div>
              <p class="text-gray-400 text-xs mt-2">
                Создано: {{ formatDate(task.created_at) }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'

export default {
  setup() {
    const router = useRouter()
    const authStore = useAuthStore()

    const tasks = ref([])
    const loading = ref(false)
    const loadingTasks = ref(false)
    const newTask = ref({
      title: '',
      description: '',
      priority: 'medium'
    })

    const loadTasks = async () => {
      loadingTasks.value = true
      try {
        const response = await api.get('/tasks')
        tasks.value = response.data.tasks || response.data
      } catch (err) {
        console.error('Failed to load tasks:', err)
      } finally {
        loadingTasks.value = false
      }
    }

    const createTask = async () => {
      loading.value = true
      try {
        const response = await api.post('/tasks', {
          title: newTask.value.title,
          description: newTask.value.description,
          priority: newTask.value.priority
        })
        tasks.value.unshift(response.data)
        newTask.value.title = ''
        newTask.value.description = ''
        newTask.value.priority = 'medium'
      } catch (err) {
        alert('Ошибка создания задачи: ' + (err.response?.data || 'Неизвестная ошибка'))
      } finally {
        loading.value = false
      }
    }

    const toggleTaskStatus = async (task) => {
      try {
        const newStatus = task.status === 'done' ? 'todo' : 'done'
        await api.patch(`/tasks/${task.id}`, { status: newStatus })
        task.status = newStatus
      } catch (err) {
        alert('Ошибка обновления задачи')
      }
    }

    const deleteTask = async (taskId) => {
      if (!confirm('Удалить задачу?')) return
      try {
        await api.delete(`/tasks/${taskId}`)
        tasks.value = tasks.value.filter(t => t.id !== taskId)
      } catch (err) {
        alert('Ошибка удаления задачи')
      }
    }

    const statusLabel = (status) => {
      const labels = {
        todo: 'К выполнению',
        in_progress: 'В работе',
        done: 'Готово'
      }
      return labels[status] || status
    }

    const priorityLabel = (priority) => {
      const labels = {
        low: 'Низкий',
        medium: 'Средний',
        high: 'Высокий'
      }
      return labels[priority] || priority
    }

    const formatDate = (dateString) => {
      if (!dateString) return '—'
      const date = new Date(dateString)
      return date.toLocaleDateString('ru-RU', {
        day: 'numeric',
        month: 'long',
        year: 'numeric'
      })
    }

    const handleLogout = async () => {
      await authStore.logout()
      router.push('/login')
    }

    onMounted(() => {
      loadTasks()
    })

    return {
      tasks,
      loading,
      loadingTasks,
      newTask,
      authStore,
      createTask,
      toggleTaskStatus,
      deleteTask,
      statusLabel,
      priorityLabel,
      formatDate,
      handleLogout
    }
  }
}
</script>
