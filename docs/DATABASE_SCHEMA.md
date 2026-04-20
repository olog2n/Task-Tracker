# Issue Tracker — Database Schema

## Overview

This document describes the database schema for all supported databases
(SQLite, PostgreSQL, MySQL). Comments in migrations may differ due to
database-specific capabilities.

---

## Table: users

Stores user accounts and authentication data.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER/SERIAL | NO | AUTO | Primary key |
| email | TEXT/VARCHAR | NO | - | Unique email address |
| password_hash | TEXT | NO | - | Bcrypt password hash |
| is_active | BOOLEAN | NO | TRUE | Soft delete flag |
| deleted_at | DATETIME | YES | NULL | Deactivation timestamp |
| deleted_by | INTEGER | YES | NULL | Who deactivated (FK → users.id) |
| token_version | INTEGER | NO | 1 | For token invalidation |
| created_at | DATETIME | NO | NOW | Account creation time |
| last_login | DATETIME | YES | NULL | Last successful login |

---

## Table: projects

Kanban boards / project containers.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER/SERIAL | NO | AUTO | Primary key |
| name | TEXT/VARCHAR | NO | - | Project name |
| description | TEXT | YES | NULL | Project description |
| owner_id | INTEGER | YES | NULL | Owner (FK → users.id) |
| is_active | BOOLEAN | NO | TRUE | Active flag |
| created_at | DATETIME | NO | NOW | Creation timestamp |
| updated_at | DATETIME | NO | NOW | Last update timestamp |
| deleted_at | DATETIME | YES | NULL | Soft delete timestamp |

---

## Table: project_members

User access to projects with roles.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER/SERIAL | NO | AUTO | Primary key |
| project_id | INTEGER | NO | - | Project (FK → projects.id) |
| user_id | INTEGER | NO | - | User (FK → users.id) |
| role | TEXT/VARCHAR | NO | 'member' | admin, member, viewer |
| joined_at | DATETIME | NO | NOW | When joined |
| added_by | INTEGER | YES | NULL | Who added (FK → users.id) |

**Constraints:**
- UNIQUE(project_id, user_id) — one entry per user per project

---

## Table: tasks

Individual tasks within projects.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER/SERIAL | NO | AUTO | Primary key |
| project_id | INTEGER | YES | NULL | Project (FK → projects.id) |
| title | TEXT/VARCHAR | NO | - | Task title |
| author | TEXT/VARCHAR | NO | - | Author email (denormalized) |
| author_id | INTEGER | NO | - | Author ID (FK → users.id) |
| description | TEXT | YES | NULL | Task description |
| executor | TEXT/VARCHAR | YES | NULL | Assigned user email |
| status | TEXT/VARCHAR | NO | 'backlog' | backlog, in_progress, done |
| created_at | DATETIME | NO | NOW | Creation timestamp |
| updated_at | DATETIME | NO | NOW | Last update timestamp |

---

## Table: audit_log

System-wide audit trail.

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | INTEGER/SERIAL | NO | AUTO | Primary key |
| actor_id | INTEGER | NO | - | Who performed action (FK → users.id) |
| action | TEXT/VARCHAR | NO | - | Action type (create, update, delete) |
| target_type | TEXT/VARCHAR | NO | - | Entity type (user, task, project) |
| target_id | INTEGER | YES | NULL | Entity ID |
| old_value | JSON/JSONB | YES | NULL | Previous state |
| new_value | JSON/JSONB | YES | NULL | New state |
| ip_address | TEXT/INET | YES | NULL | Client IP |
| user_agent | TEXT | YES | NULL | Client User-Agent |
| created_at | DATETIME | NO | NOW | Event timestamp |