export type UUID = string;

export interface ApiResponse<T> {
  data?: T;
  meta?: { cursor?: string; limit?: number; total?: number };
  error?: { code: string; message: string; details?: Record<string, any> };
}

export interface Task {
  id: UUID;
  title: string;
  project_id: UUID | null;
  process_id: UUID | null;
  current_status_id: UUID;
  assignee_id: UUID | null;
  creator_id: UUID;
  created_at: string;
  due_date: string | null;
  description: string | null;
  classification_id: UUID | null;
  urgency_id: UUID | null;
  sprint_id: UUID | null;
  parent_id: UUID | null;
  capacity: number;
  tags: string[];
  comments_container_id: UUID | null;
  attachments_container_id: UUID | null;
  updated_at: string;
}

export interface TaskCreatePayload {
  title: string;
  project_id: UUID;
  description?: string;
  classification_id?: UUID;
  urgency_id?: UUID;
  due_date?: string;
  tags?: string[];
}

export interface AuthUser {
  id: UUID;
  name: string;
  login: string;
  email?: string;
  default_project_id?: UUID;
  surname?: string;
  department?: string;
  avatar?: string;
}