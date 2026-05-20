<template>
  <div class="project-detail">
    <button @click="$router.push('/projects')">← Назад</button>

    <div v-if="project">
      <h1>{{ project.name }}</h1>
      <p>{{ project.description }}</p>

      <h2>Участники</h2>
      <ul>
        <li v-for="member in members" :key="member.id">
          {{ member.user_email }} - {{ member.role }}
        </li>
      </ul>
    </div>

    <div v-else>Загрузка...</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useProjectsStore } from '@/stores/projects'

const route = useRoute()
const projectsStore = useProjectsStore()

const project = ref(null)
const members = ref([])

onMounted(async () => {
  await projectsStore.fetchProjectById(route.params.id)
  project.value = projectsStore.currentProject

  const membersData = await projectsStore.$service.getMembers(route.params.id)
  members.value = membersData
})
</script>
