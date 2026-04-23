package tracemiddleware

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"tracker/internal/model"
	"tracker/internal/repository"

	"github.com/go-chi/chi/v5"
)

func ProjectAuthMiddleware(projectRepo repository.ProjectRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 1. Получаем project_id из URL
			projectIDStr := chi.URLParam(r, "project_id")
			if projectIDStr == "" {
				http.Error(w, "project_id required", http.StatusBadRequest)
				return
			}

			projectID, err := strconv.Atoi(projectIDStr)
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
			member, err := projectRepo.GetMember(ctx, projectID, userID)
			if err == sql.ErrNoRows {
				http.Error(w, "access denied to project", http.StatusForbidden)
				return
			}
			if err != nil {
				http.Error(w, "failed to check access", http.StatusInternalServerError)
				return
			}

			// 4. Проверяем, активен ли проект
			project, err := projectRepo.GetByID(ctx, projectID)
			if err != nil || !project.IsActive {
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

func GetProjectIDFromContext(r *http.Request) (int, bool) {
	projectID, ok := r.Context().Value(ProjectIDKey).(int)
	return projectID, ok
}

func GetProjectMemberFromContext(r *http.Request) (*model.ProjectMember, bool) {
	member, ok := r.Context().Value(ProjectMemberKey).(*model.ProjectMember)
	return member, ok
}
