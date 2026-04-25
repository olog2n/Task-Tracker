<template>
  <div class="task-detail">
    <button @click="$router.push('/tasks')">← Назад</button>

    <div v-if="task">
      <h1>{{ task.title }}</h1>
      <p>{{ task.description }}</p>
      <p>Статус: {{ task.status }}</p>
      <p>Приоритет: {{ task.priority }}</p>
    </div>

    <div v-else>Загрузка...</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useTasksStore } from '@/stores/tasks'

const route = useRoute()
const tasksStore = useTasksStore()

const task = ref(null)

onMounted(async () => {
  await tasksStore.fetchTaskById(route.params.id)
  task.value = tasksStore.currentTask
})
</script>
