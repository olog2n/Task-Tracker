import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'

// Views
import LoginView from '@/views/LoginView.vue'
import RegisterView from '@/views/RegisterView.vue'
import AppLayout from '@/layouts/AppLayout.vue'
import Dashboard from '@/views/Dashboard.vue'
import KanbanBoard from '@/views/KanbanBoard.vue'
import SprintPlanner from '@/views/SprintPlanner.vue'
import UnderConstruction from '@/views/UnderConstruction.vue'

const routes: RouteRecordRaw[] = [
  // 🔹 Гостевые маршруты (без защиты)
  { path: '/login', component: LoginView, meta: { guest: true } },
  { path: '/register', component: RegisterView, meta: { guest: true } },

  //  Защищённые маршруты проекта
  {
    path: '/projects/:projectId',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      { path: '', redirect: 'dashboard' },
      { path: 'dashboard', component: Dashboard, meta: { title: 'Дашборд' } },
      { path: 'board', component: KanbanBoard, meta: { title: 'Kanban' } },
      { path: 'sprints/plan', component: SprintPlanner, meta: { title: 'Планирование спринта' } },
      { path: 'processes/builder', component: UnderConstruction, meta: { title: 'Конструктор процессов' } },
      { path: 'admin', component: UnderConstruction, meta: { title: 'Админ-панель' } },
    ],
  },
  // Fallback
  { path: '/:pathMatch(.*)*', redirect: '/login' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 🔐 Глобальный guard
router.beforeEach((to, from, next) => {
  const isAuthenticated = !!localStorage.getItem('auth_token')

  // Если маршрут требует авторизации, а токена нет → login
  if (to.meta.requiresAuth && !isAuthenticated) {
    return next({ path: '/login', query: { redirect: to.fullPath } })
  }

  // Если гостевой маршрут, а пользователь уже вошёл → dashboard
  if (to.meta.guest && isAuthenticated) {
    return next('/projects/default/dashboard') // 🔧 Замени на реальный redirect или первый проект
  }

  next()
})

export default router
