// src/mocks/adapter.ts
import { MOCK_TASKS, MOCK_PROJECTS, MOCK_DASHBOARD, MOCK_USER } from './data'

const delay = (ms: number) => new Promise(res => setTimeout(res, ms + Math.random() * 500))

// Формируем ответ, структурно идентичный AxiosResponse
const toResponse = <T>(data: T, status = 200) => ({
  data,
  status,
  statusText: status === 200 ? 'OK' : 'Error',
  headers: {},
  config: {}
})

export const mockApi = {
  get: async (url: string) => {
    await delay(400)
    if (url.includes('/auth/profile')) return toResponse(MOCK_USER)
    if (url.includes('/tasks')) return toResponse(MOCK_TASKS)
    if (url.includes('/dashboard')) return toResponse(MOCK_DASHBOARD)
    if (url.includes('/projects')) return toResponse(MOCK_PROJECTS)
    if (url.includes('/users/me')) return toResponse(MOCK_USER)
    return toResponse(null)
  },

post: async (url: string, data?: any) => {
  await delay(600);
  console.log('📥 [MOCK POST]', url, data);

  if (url.includes('/auth/login')) {
    return toResponse({ 
      token: 'mock-token', 
      user: MOCK_USER 
    });
  }

  // ✅ Создание задачи
  if (url.includes('/tasks') && !url.includes('/transitions')) {
    const newId = `task-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const payload = data || {};
    
    const newTask = {
      id: newId,
      title: payload.title,
      project_id: payload.project_id,
      description: payload.description || null,
      classification_id: payload.classification_id || null,
      urgency_id: payload.urgency_id || null,
      due_date: payload.due_date || null,
      tags: payload.tags || [],
      // Бэкенд проставляет автоматически (имитируем)
      process_id: '00000000-0000-0000-0000-000000000100',
      current_status_id: '00000000-0000-0000-0000-000000000001',
      creator_id: 'u-1',
      created_at: new Date().toISOString(),
      assignee_id: null,
      sprint_id: null,
      capacity: 0,
      parent_id: null,
      comments_container_id: `cc-${newId}`,
      attachments_container_id: `ac-${newId}`,
      updated_at: new Date().toISOString(),
    };

    return toResponse(newTask, 201);
  }

  if (url.includes('/sprints')) return toResponse({ id: 'sprint-new', ...data });
  return toResponse({ success: true });
},

  patch: async (url: string, data?: any) => {
    await delay(300)
    console.log('📥 [MOCK PATCH]', url, data)
  
    // 👇 НОВОЕ: если обновляют профиль — возвращаем обновлённого пользователя
    if (url.includes('/auth/profile')) {
      // "Сливаем" старые данные с новыми из запроса
      const updatedUser = { ...MOCK_USER, ...data }
      return toResponse(updatedUser)
  }
  
  return toResponse({ success: true })
  },

  put: async (url: string, data?: any) => {
    await delay(300)
    console.log('📥 [MOCK PUT]', url, data)
    return toResponse({ success: true })
  },

  delete: async (url: string) => {
    await delay(300)
    console.log(' [MOCK DELETE]', url)
    return toResponse({ success: true })
  }
}
