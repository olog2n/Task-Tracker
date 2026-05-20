<template>
  <div class="tasks-view">
    <h1>Задачи</h1>

    <div class="filters">
      <select v-model="filters.project_id">
        <option value="">Все проекты</option>
        <option
          v-for="project in projects"
          :key="project.id"
          :value="project.id"
        >
          {{ project.name }}
        </option>
      </select>

      <select v-model="filters.priority">
        <option value="">Все приоритеты</option>
        <option value="low">Низкий</option>
        <option value="medium">Средний</option>
        <option value="high">Высокий</option>
      </select>

      <button @click="showCreateModal = true">Создать задачу</button>
    </div>

    <div class="task-list">
      <div
        v-for="task in tasks"
        :key="task.id"
        class="task-card"
        @click="openTask(task.id)"
      >
        <h3>{{ task.title }}</h3>
        <p>{{ task.description }}</p>
        <span class="priority">{{ task.priority }}</span>
      </div>
    </div>

    <div v-if="showCreateModal" class="modal">
      <h2>Новая задача</h2>
      <input v-model="newTask.title" placeholder="Заголовок" required />
      <textarea v-model="newTask.description" placeholder="Описание" />
      <select v-model="newTask.priority">
        <option value="low">Низкий</option>
        <option value="medium">Средний</option>
        <option value="high">Высокий</option>
      </select>
      <button @click="createTask">Создать</button>
      <button @click="showCreateModal = false">Отмена</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useTasksStore } from '@/stores/tasks'
import { useProjectsStore } from '@/stores/projects'

const router = useRouter()
const tasksStore = useTasksStore()
const projectsStore = useProjectsStore()

const tasks = ref([])
const projects = ref([])
const filters = ref({ project_id: '', priority: '' })
const showCreateModal = ref(false)
const newTask = ref({ title: '', description: '', priority: 'medium' })

const loadTasks = async () => {
  tasks.value = await tasksStore.fetchTasks(filters.value)
}

const loadProjects = async () => {
  await projectsStore.fetchProjects()
  projects.value = projectsStore.projects
}

const openTask = (taskId) => {
  router.push(`/tasks/${taskId}`)
}

const createTask = async () => {
  await tasksStore.createTask(newTask.value)
  showCreateModal.value = false
  newTask.value = { title: '', description: '', priority: 'medium' }
  await loadTasks()
}

onMounted(() => {
  loadTasks()
  loadProjects()
})
</script>

<style scoped>
.tasks-view { padding: 20px; }
.filters { margin-bottom: 20px; }
.task-list { display: grid; gap: 20px; }
.task-card {
  border: 1px solid #ddd;
  padding: 15px;
  border-radius: 8px;
  cursor: pointer;
}
.task-card:hover { background: #f5f5f5; }
.priority {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.modal {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.3);
}
.modal input, .modal textarea, .modal select {
  display: block;
  width: 100%;
  margin: 10px 0;
  padding: 10px;
}
</style>
