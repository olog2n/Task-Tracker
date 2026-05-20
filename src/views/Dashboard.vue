<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-800">Дашборд проекта</h1>

    <!-- Статистика -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-white p-4 rounded-lg shadow-sm border border-gray-200">
        <p class="text-sm text-gray-500">Закрыто в спринте</p>
        <p class="text-2xl font-bold text-gray-800">{{ stats.closed }}</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow-sm border border-gray-200">
        <p class="text-sm text-gray-500">Осталось задач</p>
        <p class="text-2xl font-bold text-gray-800">{{ stats.remaining }}</p>
      </div>
      <div class="bg-white p-4 rounded-lg shadow-sm border border-gray-200">
        <p class="text-sm text-gray-500">Дней до конца спринта</p>
        <p class="text-2xl font-bold text-gray-800">{{ stats.daysLeft }}</p>
      </div>
    </div>

    <!-- Горящие задачи -->
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
      <h2 class="text-lg font-semibold mb-3">⚠️ Задачи с истекающим сроком</h2>
      <ul class="divide-y divide-gray-100">
        <li v-for="task in urgentTasks" :key="task.id" class="py-3 flex justify-between items-center">
          <span class="text-gray-700">{{ task.title }}</span>
          <span
            class="px-2 py-1 rounded-full text-xs font-medium"
            :class="getUrgencyClass(task.due_date)"
          >
            {{ getUrgencyLabel(task.due_date) }}
          </span>
        </li>
        <li v-if="urgentTasks.length === 0" class="py-4 text-center text-gray-400">
          Нет задач с критичным сроком
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
// 🔧 Подключи vue-query здесь для реальных запросов
// const { data: tasks } = useQuery({ queryKey: ['tasks', route.params.projectId], queryFn: () => tasksApi.list(...) })

const route = useRoute()
const stats = ref({ closed: 12, remaining: 5, daysLeft: 8 })
const urgentTasks = ref([
  { id: 1, title: 'Обновить документацию API', due_date: new Date().toISOString() },
  { id: 2, title: 'Исправить таймаут авторизации', due_date: new Date(Date.now() + 7200000).toISOString() },
])

const getUrgencyLabel = (dateStr: string) => {
  const diff = new Date(dateStr).getTime() - Date.now()
  const hours = diff / 3600000
  if (diff < 0) return 'Overdue'
  if (hours < 24) return 'Due today'
  if (hours < 48) return 'Due tomorrow'
  return `Expires in ${Math.ceil(hours)}h`
}

const getUrgencyClass = (dateStr: string) => {
  const diff = new Date(dateStr).getTime() - Date.now()
  return diff < 0 ? 'bg-red-100 text-red-700' : 'bg-yellow-100 text-yellow-700'
}
</script>
