<template>
  <div class="projects-view">
    <h1>Проекты</h1>

    <button @click="showCreateModal = true">Создать проект</button>

    <div class="project-list">
      <div
        v-for="project in projects"
        :key="project.id"
        class="project-card"
        @click="openProject(project.id)"
      >
        <h3>{{ project.name }}</h3>
        <p>{{ project.description }}</p>
      </div>
    </div>

    <div v-if="showCreateModal" class="modal">
      <h2>Новый проект</h2>
      <input v-model="newProject.name" placeholder="Название" />
      <textarea v-model="newProject.description" placeholder="Описание" />
      <button @click="createProject">Создать</button>
      <button @click="showCreateModal = false">Отмена</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useProjectsStore } from '@/stores/projects'

const router = useRouter()
const projectsStore = useProjectsStore()

const projects = ref([])
const showCreateModal = ref(false)
const newProject = ref({ name: '', description: '' })

const loadProjects = async () => {
  await projectsStore.fetchProjects()
  projects.value = projectsStore.projects
}

const openProject = (projectId) => {
  router.push(`/projects/${projectId}`)
}

const createProject = async () => {
  await projectsStore.createProject(newProject.value)
  showCreateModal.value = false
  newProject.value = { name: '', description: '' }
  await loadProjects()
}

onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
.projects-view { padding: 20px; }
.project-list { display: grid; gap: 20px; margin-top: 20px; }
.project-card {
  border: 1px solid #ddd;
  padding: 15px;
  border-radius: 8px;
  cursor: pointer;
}
.project-card:hover { background: #f5f5f5; }
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
.modal input, .modal textarea {
  display: block;
  width: 100%;
  margin: 10px 0;
  padding: 10px;
}
</style>
