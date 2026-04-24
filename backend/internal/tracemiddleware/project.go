package tracemiddleware

import (
	"context"
	"database/sql"
	"net/http"
	"tracker/internal/model"
	"tracker/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ProjectAuthMiddleware(projectRepo repository.ProjectRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Получаем project_id из URL
			projectIDStr := chi.URLParam(r, "project_id")
			projectID, err := uuid.Parse(projectIDStr)
			if projectIDStr == "" {
				http.Error(w, "project_id required", http.StatusBadRequest)
				return
			}
			if err != nil {
				http.Error(w, "invalid project_id", http.StatusBadRequest)
				return
			}

			// 2. Получаем user_id из контекста (AuthMiddleware уже проверил токен)
			userID, ok := GetUserIDFromContext(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// 3. Проверяем доступ к проекту
			member, err := projectRepo.GetMemberByID(ctx, projectID, userID)
			if err == sql.ErrNoRows {
				http.Error(w, "access denied to project", http.StatusForbidden)
				return
			}
			if err != nil {
				http.Error(w, "failed to check access", http.StatusInternalServerError)
				return
			}

			// 4. Проверяем, активен ли проект
			_, err = projectRepo.GetProjectByID(ctx, projectID)
			if err != nil {
				http.Error(w, "project not found or inactive", http.StatusNotFound)
				return
			}

			// 5. Добавляем в контекст для использования в хендлере
			ctx = context.WithValue(ctx, ProjectIDKey, projectID)
			ctx = context.WithValue(ctx, ProjectMemberKey, member)

			// 6. Передаём управление следующему хендлеру
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetProjectIDFromContext(r *http.Request) (uuid.UUID, bool) {
	projectID, ok := r.Context().Value(ProjectIDKey).(uuid.UUID)
	return projectID, ok
}

func GetProjectMemberFromContext(r *http.Request) (*model.ProjectMember, bool) {
	member, ok := r.Context().Value(ProjectMemberKey).(*model.ProjectMember)
	return member, ok
}
