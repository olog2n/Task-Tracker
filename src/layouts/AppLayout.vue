<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <!-- TopBar -->
    <header class="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between shadow-sm">
      <div class="flex items-center gap-4">
        <span class="font-bold text-lg text-gray-800">TaskEngine</span>
        <!-- 🔧 Селектор проекта: берет projectId из route params -->
        <select
          v-model="currentProjectId"
          class="bg-gray-100 border border-gray-300 text-gray-700 py-1 px-3 rounded-md text-sm"
          @change="navigateToProject"
        >
          <option disabled value="">Выберите проект</option>
          <option v-for="p in mockProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
        </select>
      </div>

      <nav class="flex gap-2">
        <RouterLink
            v-for="link in navLinks"
            :key="link.to"
            :to="`/projects/${currentProjectId}/${link.to}`"
            class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors"
            :class="$route.path.includes(link.to) ? 'bg-blue-100 text-blue-700' : 'text-gray-600 hover:bg-gray-100'"
        >
          {{ link.label }}
        </RouterLink>
      </nav>

      <div class="flex items-center gap-3">
        <span class="text-sm text-gray-500">{{ mockUser.name }}</span>
        <div class="w-8 h-8 rounded-full bg-blue-500 text-white flex items-center justify-center text-xs font-bold">
          {{ mockUser.initials }}
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 p-6 overflow-auto">
      <RouterView :key="$route.fullPath" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const mockProjects = [
  { id: 'proj-1', name: 'Project Aurora' },
  { id: 'proj-2', name: 'Internal Tools' }
]
const mockUser = { name: 'Иван Петров', initials: 'ИП' }

const currentProjectId = ref(route.params.projectId as string)

watch(() => route.params.projectId, (newId) => {
  currentProjectId.value = newId as string
})

const navigateToProject = () => {
  if (currentProjectId.value) {
    router.push(`/projects/${currentProjectId.value}/dashboard`)
  }
}

const navLinks = [
  { to: 'dashboard', label: 'Дашборд' },
  { to: 'board', label: 'Kanban' },
  { to: 'sprints/plan', label: 'Спринты' },
]
</script>
