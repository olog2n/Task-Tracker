<!-- src/components/TaskDetailModal.vue -->
<template>
  <Teleport to="body">
    <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <!-- Backdrop -->
      <div 
        class="absolute inset-0 bg-black/50 backdrop-blur-sm"
        @click="emit('close')"
      />
      
      <!-- 🔹 Модалка: убран flex, теперь просто блок -->
      <div class="relative bg-white rounded-xl shadow-xl w-full max-w-lg p-6 max-h-[85vh] overflow-y-auto block">
        
        <!-- Кнопка закрытия -->
        <button 
          @click="emit('close')"
          class="absolute top-4 right-4 p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-lg transition z-10"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>

        <!-- Заголовок -->
        <div class="pr-8 mb-6 block">
          <h3 class="text-lg font-semibold text-gray-900">
            {{ task?.title || 'Задача' }}
          </h3>
        </div>

        <!-- 🔹 Поля задачи: block, а не flex! -->
        <div class="space-y-3 mb-6 block">
          <div class="block">
            <strong class="text-gray-700">Описание: </strong> 
            <span class="text-gray-600">{{ task?.description || 'Нет описания' }}</span>
          </div>
          <div class="block">
            <strong class="text-gray-700">Исполнитель: </strong> 
            <span class="text-gray-600">{{ task?.assignee || 'Не назначен' }}</span>
          </div>
          <div class="block">
            <strong class="text-gray-700">Дедлайн: </strong> 
            <span class="text-gray-600">{{ task?.due_date || 'Не указан' }}</span>
          </div>
          <div class="block">
            <strong class="text-gray-700">Приоритет: </strong> 
            <span class="text-gray-600">{{ task?.urgency || 'Не указан' }}</span>
          </div>
          <div class="block">
            <strong class="text-gray-700">Статус: </strong> 
            <span class="text-gray-600">{{ task?.status || 'Нет статуса' }}</span>
          </div>
        </div>

        <!-- 🔹 Комментарии: ТОЖЕ block, внизу модалки -->
        <div v-if="task?.comments?.length" class="border-t pt-4 block">
          <h4 class="text-sm font-semibold text-gray-700 mb-3 block">Комментарии</h4>
          <div class="space-y-2 max-h-40 overflow-y-auto pr-1 block">
            <div v-for="c in task.comments" :key="c.id" class="p-3 bg-gray-50 rounded-lg text-sm block">
              <div class="flex justify-between items-start mb-1">
                <span class="font-medium text-gray-700">{{ c.assignee }}</span>
                <span class="text-xs text-gray-400 flex-shrink-0 ml-2">
                  {{ new Date(c.date).toLocaleString('ru-RU') }}
                </span>
              </div>
              <p class="text-gray-800 whitespace-pre-wrap break-words">{{ c.text }}</p>
            </div>
          </div>
        </div>

      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">

    const props = defineProps<{
        modelValue: boolean;
        task: any | null;
        projectId: string;
    }>();

    const emit = defineEmits<{
        (e: 'update:modelValue', value: boolean): void;
        (e: 'close'): void;
    }>();
</script>

<style scoped>
/* Стили только для этого компонента */
</style>