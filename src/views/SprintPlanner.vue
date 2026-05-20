<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 h-full">
    <!-- Левая часть: Бэклог -->
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 flex flex-col">
      <h2 class="text-lg font-semibold mb-3">Бэклог проекта</h2>
      <div class="flex-1 overflow-y-auto space-y-2 pr-2">
        <div v-for="task in backlogTasks" :key="task.id" class="flex items-center gap-3 p-2 hover:bg-gray-50 rounded-md border border-transparent hover:border-gray-200">
          <input type="checkbox" :value="task.id" v-model="selectedTaskIds" class="w-4 h-4 text-blue-600 rounded focus:ring-blue-500" />
          <div class="flex-1">
            <p class="text-sm font-medium text-gray-800">{{ task.title }}</p>
            <p class="text-xs text-gray-500">{{ task.status }} • {{ task.points }} SP</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Правая часть: Создание спринта -->
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 flex flex-col">
      <h2 class="text-lg font-semibold mb-3">Создать спринт</h2>
      <form @submit.prevent="createSprint" class="space-y-4 flex-1 flex flex-col">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Название</label>
          <input v-model="form.name" type="text" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500" placeholder="Q4 Release Sprint" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Описание</label>
          <textarea v-model="form.desc" rows="3" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-blue-500 focus:border-blue-500"></textarea>
        </div>
        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Начало</label>
            <input v-model="form.start" type="date" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Конец</label>
            <input v-model="form.end" type="date" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm" />
          </div>
        </div>

        <div class="mt-auto pt-4 border-t border-gray-100">
          <p class="text-sm text-gray-600 mb-2">Выбрано задач: <strong>{{ selectedTaskIds.length }}</strong></p>
          <button type="submit" class="w-full bg-blue-600 text-white py-2 rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-50" :disabled="selectedTaskIds.length === 0 || !form.name">
            Создать и добавить задачи
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/client' // 🔧 Используем централизованный клиент
import { tasksApi } from '@/api/tasks'

const route = useRoute()
const projectId = route.params.projectId as string

const backlogTasks = ref([
  { id: 't1', title: 'API валидация', status: 'To Do', points: 5 },
  { id: 't2', title: 'UI кнопок', status: 'To Do', points: 3 },
])

const selectedTaskIds = ref<string[]>([])
const form = ref({ name: '', desc: '', start: '', end: '' })

const createSprint = async () => {
  try {
    //  api уже имеет baseURL: '/api/v0', поэтому префикс не нужен
    const { data: sprint } = await api.post(`/projects/${projectId}/sprints`, {
      name: form.value.name,
      description: form.value.desc,
      start_date: form.value.start,
      end_date: form.value.end
    })

    // Batch обновляем задачи (привязываем к спринту)
    await Promise.all(selectedTaskIds.value.map(taskId =>
      tasksApi.update(taskId, { sprint_id: sprint.id })
    ))

    console.log('✅ Sprint created & tasks assigned')
    // TODO: показать toast / редирект на доску спринта
  } catch (e: any) {
    console.error('❌ Failed to create sprint:', e.response?.data || e.message)
    // TODO: показать ошибку пользователю
  }
}
</script>
