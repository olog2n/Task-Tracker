<template>
  <div class="h-full flex flex-col">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-xl font-bold text-gray-800">Kanban-доска</h1>
      <!-- 🔧 Здесь будет селектор спринта -->
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

          <!-- 🔧 ЗДЕСЬ БУДЕТ DROPPABLE ЗОНА от @dnd-kit/vue -->
          <div class="flex-1 space-y-2 overflow-y-auto pr-1">
            <div
              v-for="task in col.tasks"
              :key="task.id"
              class="bg-white p-3 rounded-md shadow-sm border border-gray-200 cursor-grab active:cursor-grabbing hover:shadow-md transition-shadow"
              draggable="true"
              @dragstart="handleDragStart($event, task)"
            >
              <div class="flex justify-between items-start mb-2">
                <span class="text-sm font-medium text-gray-800">{{ task.title }}</span>
                <span class="text-xs px-1.5 py-0.5 rounded bg-blue-100 text-blue-700">{{ task.urgency }}</span>
              </div>
              <div class="flex items-center gap-2 mt-2">
                <div class="w-6 h-6 rounded-full bg-gray-300 text-[10px] flex items-center justify-center">👤</div>
                <span class="text-xs text-gray-500">{{ task.assignee }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
// 🔧 ИНТЕГРАЦИЯ DnD:
// 1. npm i @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
// 2. Оберни колонки в <DndContext>, задачи в <SortableContext>
// 3. В onDragEnd вызывай optimistic update + api.transition()
// 4. При ошибке FSM откатывай состояние и показывай toast

const columns = ref([
  { id: 'backlog', title: 'Backlog', tasks: [{ id: 1, title: 'Фича A', urgency: 'Low', assignee: 'Иван' }] },
  { id: 'progress', title: 'In Progress', tasks: [{ id: 2, title: 'Баг B', urgency: 'High', assignee: 'Мария' }] },
  { id: 'review', title: 'Review', tasks: [] },
  { id: 'done', title: 'Done', tasks: [{ id: 3, title: 'Документация', urgency: 'Medium', assignee: 'Алексей' }] },
])

const handleDragStart = (e: DragEvent, task: any) => {
  // 🔧 Нативный DnD для MVP. Замени на @dnd-kit для production
  e.dataTransfer?.setData('text/plain', JSON.stringify(task))
}
</script>
