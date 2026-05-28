<!-- src/components/CreateTaskModal.vue -->
<template>
  <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="close">
    <div class="bg-white rounded-lg shadow-xl w-full max-w-lg p-6">
      <h3 class="text-lg font-semibold mb-4">Создать задачу</h3>
      
      <form @submit.prevent="submit" class="space-y-4">
        <!-- Название (обязательно) -->
        <div>
          <label class="block text-sm font-medium mb-1">Название *</label>
          <input 
            v-model="form.title" 
            class="w-full border rounded px-3 py-2"
            placeholder="Кратко опишите задачу..."
            required
            autofocus
          />
        </div>
        
        <!-- Описание (опционально) -->
        <div>
          <label class="block text-sm font-medium mb-1">Описание</label>
          <textarea 
            v-model="form.description" 
            class="w-full border rounded px-3 py-2"
            rows="3"
            placeholder="Детали, шаги, ожидаемый результат..."
          />
        </div>
        
        <!-- Дедлайн -->
        <div>
          <label class="block text-sm font-medium mb-1">Дедлайн</label>
          <input 
            v-model="form.due_date" 
            type="date"
            class="w-full border rounded px-3 py-2"
          />
        </div>
        
        <!-- Ошибка -->
        <div v-if="error" class="text-red-600 text-sm">{{ error }}</div>
        
        <!-- Кнопки -->
        <div class="flex justify-end gap-2 pt-2">
          <button 
            type="button" 
            @click="close"
            class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded"
            :disabled="loading"
          >
            Отмена
          </button>
          <button 
            type="submit"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            :disabled="loading || !form.title.trim()"
          >
            {{ loading ? 'Создание...' : 'Создать' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { tasksApi } from '@/api/tasks'
import type { TaskCreatePayload, Task } from '@/types'

const props = defineProps<{
  open: boolean
  projectId: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success', task: Task): void
}>()

const form = ref<TaskCreatePayload>({
  title: '',
  project_id: props.projectId,
  description: '',
  due_date: undefined,
})

const loading = ref(false)
const error = ref<string | null>(null)

const submit = async () => {
  if (!form.value.title.trim()) return
  
  loading.value = true
  error.value = null
  
  try {
    const response = await tasksApi.create(form.value)
    
    if (response.data?.error) {
      error.value = response.data.error.message
      return
    }
    
    emit('success', response.data.data as Task)
    close()
  } catch (e: any) {
    error.value = e.response?.data?.error?.message || 'Не удалось создать задачу'
  } finally {
    loading.value = false
  }
}

const close = () => {
  if (!loading.value) {
    form.value = { title: '', project_id: props.projectId, description: '', due_date: undefined }
    error.value = null
    emit('close')
  }
}
</script>