<template>
  <div class="min-h-screen bg-gray-50 flex flex-col">
    <!-- TopBar -->
    <header class="bg-white border-b border-gray-200 px-6 py-3 shadow-sm">
      <div class="flex items-center justify-between">
        <!-- Левая часть: логотип + проект + кнопка -->
        <div class="flex items-center gap-4">
          <span class="font-bold text-lg text-gray-800">TaskEngine</span>
          <select
            v-model="currentProjectId"
            class="bg-gray-100 border border-gray-300 text-gray-700 py-1 px-3 rounded-md text-sm"
            @change="navigateToProject"
          >
            <option disabled value="">Выберите проект</option>
            <option v-for="p in mockProjects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <button 
            v-if="currentProjectId"
            @click="openCreateModal"
            class="px-3 py-1.5 bg-blue-600 text-white rounded-md text-sm font-medium hover:bg-blue-700 transition"
            title="Создать задачу (Ctrl+K)"
          >
            Создать задачу
          </button>
        </div>

        <!-- Центральная часть: навигация -->
        <nav class="flex gap-2">
          <RouterLink
            v-for="link in navLinks"
            :key="link.to"
            :to="`/projects/${currentProjectId}/${link.to}`"
            class="px-3 py-1.5 rounded-md text-sm font-medium transition-colors"
            :class="$route.path.endsWith('/' + link.to) ? 'bg-blue-100 text-blue-700' : 'text-gray-600 hover:bg-gray-100'"
          >
            {{ link.label }}
          </RouterLink>
        </nav>

        <!-- Правая часть: профиль -->
        <RouterLink :to="{ name: 'Profile' }" class="flex items-center gap-2 hover:opacity-80 transition">
          <img 
            v-if="user?.avatar" 
            :src="user.avatar + '?v=' + Date.now()"
            class="w-8 h-8 rounded-full object-cover"
          />
          <div v-else class="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center text-indigo-700 font-medium">
            {{ user?.name?.[0] }}{{ user?.surname?.[0] }}
          </div>
          <span class="text-sm font-medium">{{ user?.name }} {{ user?.surname }}</span>
        </RouterLink>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 p-6 overflow-auto">
      <RouterView :key="$route.fullPath" />
    </main>
    <CreateTaskModal 
      :open="isCreateModalOpen"
      :project-id="currentProjectId || ''"
      @close="closeCreateModal"
      @success="handleTaskCreated"
  />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '@/composables/useAuth'
import type { Task } from '@/types'
import CreateTaskModal from '@/views/CreateTaskModal.vue'


const route = useRoute()
const router = useRouter()
const { user, logout } = useAuth()

const mockProjects = [
  { id: 'proj-1', name: 'Project Aurora' },
  { id: 'proj-2', name: 'Internal Tools' }
]

const currentProjectId = ref<string>((route.params.projectId as string) || '')


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



const isCreateModalOpen = ref(false)

const openCreateModal = () => {
  isCreateModalOpen.value = true
}

const closeCreateModal = () => {
  isCreateModalOpen.value = false
}

const handleTaskCreated = (task: Task) => {
  console.log('✅ Задача создана:', task)
  closeCreateModal()
  // TODO: здесь позже добавим toast и обновление Kanban
}

onMounted(() => {
  const handler = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
      e.preventDefault()
      if (currentProjectId.value) openCreateModal()
    }
  }
  window.addEventListener('keydown', handler)
  return () => window.removeEventListener('keydown', handler)
})
</script>
