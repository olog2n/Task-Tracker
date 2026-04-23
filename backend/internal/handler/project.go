package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"tracker/internal/model"
	"tracker/internal/repository"
	"tracker/internal/tracemiddleware"

	"github.com/go-chi/chi/v5"
)

type ProjectHandler struct {
	projectRepo repository.ProjectRepository
	userRepo    repository.UserRepository
}

func NewProjectHandler(projectRepo repository.ProjectRepository, userRepo repository.UserRepository) *ProjectHandler {
	return &ProjectHandler{
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

// @Summary      Create a new project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body model.Project true "Project data"
// @Success      201  {object}  model.Project
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /api/projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	userID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input model.Project
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	input.OwnerID = sql.NullInt64{Int64: int64(userID), Valid: true}
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	projectID, err := h.projectRepo.Create(ctx, &input)
	if err != nil {
		log.Printf("create project error: %v", err)
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	member := &model.ProjectMember{
		ProjectID: int(projectID),
		UserID:    userID,
		Role:      model.RoleProjectAdmin,
		JoinedAt:  time.Now(),
		AddedBy:   sql.NullInt64{Int64: int64(userID), Valid: true},
	}
	if err := h.projectRepo.AddMember(ctx, member); err != nil {
		log.Printf("add member error: %v", err)
		http.Error(w, "failed to add project owner as member", http.StatusInternalServerError)
		return
	}

	input.ID = int(projectID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

// @Summary      Get all projects
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Project
// @Failure      401  {object}  map[string]string
// @Router       /api/projects [get]
func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	userID, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := h.projectRepo.GetAll(ctx, userID)
	if err != nil {
		log.Printf("get projects error: %v", err)
		http.Error(w, "failed to get projects", http.StatusInternalServerError)
		return
	}

	if projects == nil {
		projects = make([]model.Project, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// @Summary      Get a project by ID
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        project_id path int true "Project ID"
// @Success      200  {object}  model.Project
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/projects/{project_id} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	projectID, ok := tracemiddleware.GetProjectIDFromContext(r)
	if !ok {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(ctx, projectID)
	if err == sql.ErrNoRows {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("get project error: %v", err)
		http.Error(w, "failed to get project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

// @Summary      Update a project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        project_id path int true "Project ID"
// @Param        request body model.Project true "Project data"
// @Success      200  {object}  model.Project
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/projects/{project_id} [put]
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	member, ok := tracemiddleware.GetProjectMemberFromContext(r)
	if !ok || !member.CanEdit() {
		http.Error(w, "forbidden - project admin only", http.StatusForbidden)
		return
	}

	projectID, _ := tracemiddleware.GetProjectIDFromContext(r)

	var input model.Project
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	input.ID = projectID
	input.UpdatedAt = time.Now()

	if err := h.projectRepo.Update(ctx, &input); err != nil {
		log.Printf("update project error: %v", err)
		http.Error(w, "failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(input)
}

// @Summary      Delete a project
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        project_id path int true "Project ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /api/projects/{project_id} [delete]
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	member, ok := tracemiddleware.GetProjectMemberFromContext(r)
	if !ok || !member.CanDelete() {
		http.Error(w, "forbidden - project admin only", http.StatusForbidden)
		return
	}

	projectID, _ := tracemiddleware.GetProjectIDFromContext(r)

	if err := h.projectRepo.Delete(ctx, projectID); err != nil {
		log.Printf("delete project error: %v", err)
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Get project members
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        project_id path int true "Project ID"
// @Success      200  {array}   model.ProjectMember
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/projects/{project_id}/members [get]
func (h *ProjectHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	projectID, ok := tracemiddleware.GetProjectIDFromContext(r)
	if !ok {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	members, err := h.projectRepo.GetMembers(ctx, projectID)
	if err != nil {
		log.Printf("get members error: %v", err)
		http.Error(w, "failed to get members", http.StatusInternalServerError)
		return
	}

	if members == nil {
		members = make([]model.ProjectMember, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// @Summary      Add a member to project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        project_id path int true "Project ID"
// @Param        request body AddMemberInput true "Member data"
// @Success      201  {object}  model.ProjectMember
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/projects/{project_id}/members [post]
func (h *ProjectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	member, ok := tracemiddleware.GetProjectMemberFromContext(r)
	if !ok || !member.CanManageMembers() {
		http.Error(w, "forbidden - project admin only", http.StatusForbidden)
		return
	}

	projectID, _ := tracemiddleware.GetProjectIDFromContext(r)

	var input struct {
		UserID int               `json:"user_id"`
		Role   model.ProjectRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.UserID == 0 {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if input.Role == "" {
		input.Role = model.RoleProjectMember
	}

	_, err := h.userRepo.GetByID(ctx, input.UserID)
	if err == sql.ErrNoRows {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "failed to get user", http.StatusInternalServerError)
		return
	}

	newMember := &model.ProjectMember{
		ProjectID: projectID,
		UserID:    input.UserID,
		Role:      input.Role,
		JoinedAt:  time.Now(),
		AddedBy:   sql.NullInt64{Int64: int64(member.UserID), Valid: true},
	}

	if err := h.projectRepo.AddMember(ctx, newMember); err != nil {
		log.Printf("add member error: %v", err)
		http.Error(w, "failed to add member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMember)
}

// @Summary      Update member role
// @Tags         projects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        project_id path int true "Project ID"
// @Param        user_id path int true "User ID"
// @Param        request body UpdateRoleInput true "New role"
// @Success      200  {object}  model.ProjectMember
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/projects/{project_id}/members/{user_id} [put]
func (h *ProjectHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	member, ok := tracemiddleware.GetProjectMemberFromContext(r)
	if !ok || !member.CanManageMembers() {
		http.Error(w, "forbidden - project admin only", http.StatusForbidden)
		return
	}

	projectID, _ := tracemiddleware.GetProjectIDFromContext(r)
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	var input struct {
		Role model.ProjectRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.projectRepo.UpdateMemberRole(ctx, projectID, userID, input.Role); err != nil {
		log.Printf("update role error: %v", err)
		http.Error(w, "failed to update role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"project_id": projectID,
		"user_id":    userID,
		"role":       input.Role,
	})
}

// @Summary      Remove member from project
// @Tags         projects
// @Produce      json
// @Security     BearerAuth
// @Param        project_id path int true "Project ID"
// @Param        user_id path int true "User ID"
// @Success      204
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/projects/{project_id}/members/{user_id} [delete]
func (h *ProjectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	member, ok := tracemiddleware.GetProjectMemberFromContext(r)
	if !ok || !member.CanManageMembers() {
		http.Error(w, "forbidden - project admin only", http.StatusForbidden)
		return
	}

	projectID, _ := tracemiddleware.GetProjectIDFromContext(r)
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	if userID == member.UserID {
		http.Error(w, "cannot remove yourself", http.StatusBadRequest)
		return
	}

	if err := h.projectRepo.RemoveMember(ctx, projectID, userID); err != nil {
		log.Printf("remove member error: %v", err)
		http.Error(w, "failed to remove member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
