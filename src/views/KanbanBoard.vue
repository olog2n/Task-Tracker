<template>
  <div class="h-full flex flex-col">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-xl font-bold text-gray-800">Kanban-доска</h1>
      <span class="text-sm text-gray-500">Спринт: Sprint-42</span>
    </div>

    <!-- Канбан колонки -->
    <div class="flex-1 overflow-x-auto pb-4">
      <div class="flex gap-4 min-w-max h-full">
        <div v-for="col in columns" :key="col.id" class="w-72 flex flex-col bg-gray-100 rounded-lg p-3">
          <h3 class="font-semibold text-gray-700 mb-3 flex justify-between">
            {{ col.title }}
            <span class="text-xs bg-gray-200 px-2 py-0.5 rounded-full">{{ col.tasks.length }}</span>
          </h3>

          <!-- DROPPABLE ЗОНА (нативный HTML5 DnD для MVP) -->
          <div
            class="flex-1 space-y-2 overflow-y-auto pr-1 min-h-[100px]"
            @dragover.prevent
            @drop="handleDrop($event, col.id)"
          >
            <div
              v-for="task in col.tasks"
              :key="task.id"
              class="bg-white p-3 rounded-md shadow-sm border border-gray-200 cursor-grab active:cursor-grabbing hover:shadow-md transition-shadow"
              draggable="true"
              @dragstart="handleDragStart($event, task)"
            >
              <div class="flex justify-between items-start mb-2">
                <span class="text-sm font-medium text-gray-800">{{ task.title }}</span>
                <span
                  class="text-xs px-1.5 py-0.5 rounded"
                  :class="getUrgencyBadgeClass(task.urgency)"
                >
                  {{ task.urgency }}
                </span>
              </div>
              <div class="flex items-center gap-2 mt-2">
                <div class="w-6 h-6 rounded-full bg-gray-300 text-[10px] flex items-center justify-center">👤</div>
                <span class="text-xs text-gray-500">{{ task.assignee || 'Не назначен' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import api from '@/api/client'

// 🔧 1. Интерфейсы для строгой типизации
export interface Task {
  id: string
  title: string
  status: 'backlog' | 'in-progress' | 'review' | 'done'
  urgency: 'low' | 'medium' | 'high' | 'critical'
  assignee: string | null
  points?: number
  due_date?: string
}

interface Column {
  id: Task['status']
  title: string
  status: Task['status']
  tasks: Task[]
}

// 🔧 2. Явная типизация реактивных переменных
const allTasks = ref<Task[]>([])

// 🔧 3. Вычисляемые колонки (группировка задач по статусу)
const columns = computed<Column[]>(() => {
  // Явно указываем тип для массива определений
  const defs: Array<{ id: Task['status']; title: string; status: Task['status'] }> = [
    { id: 'backlog', title: 'Backlog', status: 'backlog' },
    { id: 'in-progress', title: 'In Progress', status: 'in-progress' },
    { id: 'review', title: 'Review', status: 'review' },
    { id: 'done', title: 'Done', status: 'done' },
  ]

  return defs.map(def => ({
    ...def,
    tasks: allTasks.value.filter(t => t.status === def.status)
  }))
})

// 🔧 4. Хелперы для UI (видны в шаблоне)
const getUrgencyBadgeClass = (urgency: Task['urgency']): string => {
  const map: Record<Task['urgency'], string> = {
    low: 'bg-gray-100 text-gray-700',
    medium: 'bg-blue-100 text-blue-700',
    high: 'bg-orange-100 text-orange-700',
    critical: 'bg-red-100 text-red-700',
  }
  return map[urgency]
}

// 🔧 5. Обработчики Drag & Drop (нативный HTML5 API)
const handleDragStart = (e: DragEvent, task: Task) => {
  // Сохраняем ID задачи в DataTransfer для использования в onDrop
  e.dataTransfer?.setData('text/plain', task.id)
  e.dataTransfer!.effectAllowed = 'move'
}

const handleDrop = async (e: DragEvent, targetStatus: Task['status']) => {
  e.preventDefault()
  const taskId = e.dataTransfer?.getData('text/plain')
  if (!taskId) return

  // 🔧 Оптимистичное обновление (сразу меняем локальный стейт)
  const taskIndex = allTasks.value.findIndex(t => t.id === taskId)
  if (taskIndex === -1) return

  const oldStatus = allTasks.value[taskIndex].status

  // Меняем статус локально
  allTasks.value[taskIndex].status = targetStatus

  try {
    // 🔧 Отправляем запрос на бэкенд (или мок)
    await api.patch(`/tasks/${taskId}`, {
      current_status_id: targetStatus // 🔧 Маппинг на контракт бэкенда
    })
    // Если нужно, можно вызвать invalidateQueries vue-query здесь
  } catch (err) {
    // 🔧 Откат при ошибке (FSM guard failed, network error, etc.)
    console.error('Failed to transition task, rolling back', err)
    if (allTasks.value[taskIndex]) {
      allTasks.value[taskIndex].status = oldStatus
    }
    // TODO: показать toast с ошибкой
  }
}

// 🔧 6. Загрузка данных при монтировании
onMounted(async () => {
  // 🔧 Явный дженерик <Task[]> для правильной типизации ответа
  const { data } = await api.get<Task[]>('/tasks')
  allTasks.value = data
})
</script>
