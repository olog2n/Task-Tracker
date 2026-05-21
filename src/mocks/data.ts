// src/mocks/data.ts

export const MOCK_USER = {
  id: 'u-1',
  name: 'Алексей Разработчик',
  email: 'alex@taskengine.io',
  avatar: null,
  default_project_id: 'proj-1'
}

export const MOCK_PROJECTS = [
  { id: 'proj-1', name: 'Project Aurora' },
  { id: 'proj-2', name: 'Task Engine Core' },
]

// Задачи для Канбана
// export const MOCK_TASKS = [
//   { id: 't-1', title: 'Настроить Vite + Tailwind', status: 'done', urgency: 'high', assignee: 'Алексей', points: 3 },
//   { id: 't-2', title: 'Сверстать модалку создания задачи', status: 'in-progress', urgency: 'medium', assignee: 'Мария', points: 5 },
//   { id: 't-3', title: 'Исправить баг с авторизацией', status: 'review', urgency: 'critical', assignee: 'Алексей', points: 2 },
//   { id: 't-4', title: 'Интегрировать API спринтов', status: 'backlog', urgency: 'low', assignee: null, points: 8 },
//   { id: 't-5', title: 'Добавить Drag & Drop', status: 'in-progress', urgency: 'high', assignee: 'Иван', points: 13 },
//   { id: 't-6', title: 'Написать документацию', status: 'backlog', urgency: 'low', assignee: null, points: 3 },
// ]

// src/mocks/data.ts

type Task = {
  id: string,
  title: string,
  status: string,
  urgency: string,
  assignee: string,
  points: number
}

export const MOCK_TASKS: Task[] = [
  {
    id: 't-1',
    title: 'Настроить Vite + Tailwind',
    status: 'done',
    urgency: 'high',
    assignee: 'Алексей',
    points: 3
  },
  {
    id: 't-2',
    title: 'Сверстать модалку создания задачи',
    status: 'in-progress',
    urgency: 'medium',
    assignee: 'Мария',
    points: 5
  },
  {
    id: 't-3',
    title: 'Исправить баг с авторизацией',
    status: 'review',
    urgency: 'critical',
    assignee: 'Алексей',
    points: 2
  },
  {
    id: 't-4',
    title: 'Интегрировать API спринтов',
    status: 'backlog',
    urgency: 'low',
    assignee: 'null',
    points: 8
  },
  {
    id: 't-5',
    title: 'Добавить Drag & Drop',
    status: 'in-progress',
    urgency: 'high',
    assignee: 'Иван',
    points: 13
  },
]

// Данные для Дашборда
export const MOCK_DASHBOARD = {
  stats: { closed: 12, remaining: 5, daysLeft: 8 },
  urgentTasks: [
    { id: 't-7', title: 'Обновить SSL сертификаты', due_date: new Date().toISOString() },
    { id: 't-8', title: 'Релиз v1.0.4', due_date: new Date(Date.now() + 7200000).toISOString() },
    { id: 't-9', title: 'Аудит безопасности', due_date: new Date(Date.now() - 86400000).toISOString() },
  ]
}
