package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"tracker/internal/model"
	"tracker/internal/repository"
	"tracker/internal/tracemiddleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

// @Summary		Create a new project
// @Description	Create a new project with the given name and description
// @Tags			projects
// @Accept			json
// @Produce		json
// @Param			request	body		model.ProjectInput	true	"Project creation data"
// @Success		201		{object}	model.Project
// @Failure		400		{object}	map[string]string	"Invalid request body"
// @Failure		401		{object}	map[string]string	"Unauthorized"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/api/projects [post]
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

	input.OwnerID = userID
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	projectID, err := h.projectRepo.CreateProject(ctx, &input)
	if err != nil {
		log.Printf("create project error: %v", err)
		http.Error(w, "failed to create project", http.StatusInternalServerError)
		return
	}

	member := &model.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      model.RoleProjectAdmin,
		JoinedAt:  time.Now(),
		AddedBy:   userID,
	}
	if err := h.projectRepo.AddProjectMember(ctx, member); err != nil {
		log.Printf("add member error: %v", err)
		http.Error(w, "failed to add project owner as member", http.StatusInternalServerError)
		return
	}

	input.ID = projectID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

// @Summary		Get all projects
// @Description	Get a list of all projects accessible by the user
// @Tags			projects
// @Produce		json
// @Success		200	{array}		model.Project
// @Failure		401	{object}	map[string]string	"Unauthorized"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/api/projects [get]
func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	_, ok := tracemiddleware.GetUserIDFromContext(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	projects, err := h.projectRepo.GetAllProjects(ctx, 50, 0)
	if err != nil {
		log.Printf("get projects error: %v", err)
		http.Error(w, "failed to get projects", http.StatusInternalServerError)
		return
	}

	if projects == nil {
		projects = make([]*model.Project, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// @Summary		Get project by ID
// @Description	Get a specific project by its UUID
// @Tags			projects
// @Produce		json
// @Param			id	path		string	true	"Project UUID"
// @Success		200	{object}	model.Project
// @Failure		400	{object}	map[string]string	"Invalid project ID"
// @Failure		404	{object}	map[string]string	"Project not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/api/projects/{id} [get]
func (h *ProjectHandler) GetProjectById(w http.ResponseWriter, r *http.Request) {
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

	project, err := h.projectRepo.GetProjectByID(ctx, projectID)
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

// @Summary		Update project
// @Description	Update an existing project
// @Tags			projects
// @Accept			json
// @Produce		json
// @Param			id		path		string				true	"Project UUID"
// @Param			request	body		model.ProjectInput	true	"Project update data"
// @Success		200		{object}	model.Project
// @Failure		400		{object}	map[string]string	"Invalid request"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		404		{object}	map[string]string	"Project not found"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/api/projects/{id} [put]
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

	var input model.ProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if _, err := h.projectRepo.UpdateProject(ctx, projectID, &input, member.ID); err != nil {
		log.Printf("update project error: %v", err)
		http.Error(w, "failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(input)
}

// @Summary		Delete project
// @Description	Delete a project by its UUID
// @Tags			projects
// @Produce		json
// @Param			id	path	string	true	"Project UUID"
// @Success		204	"No content"
// @Failure		400	{object}	map[string]string	"Invalid project ID"
// @Failure		403	{object}	map[string]string	"Forbidden"
// @Failure		404	{object}	map[string]string	"Project not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/api/projects/{id} [delete]
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

	if err := h.projectRepo.DeleteProject(ctx, projectID); err != nil {
		log.Printf("delete project error: %v", err)
		http.Error(w, "failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary		Get project members
// @Description	Get all members of a project
// @Tags			projects
// @Produce		json
// @Param			id	path		string	true	"Project UUID"
// @Success		200	{array}		model.ProjectMember
// @Failure		400	{object}	map[string]string	"Invalid project ID"
// @Failure		404	{object}	map[string]string	"Project not found"
// @Failure		500	{object}	map[string]string	"Internal server error"
// @Router			/api/projects/{id}/members [get]
func (h *ProjectHandler) GetProjectMembers(w http.ResponseWriter, r *http.Request) {
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

	members, err := h.projectRepo.GetProjectMembers(ctx, projectID)
	if err != nil {
		log.Printf("get members error: %v", err)
		http.Error(w, "failed to get members", http.StatusInternalServerError)
		return
	}

	if members == nil {
		members = make([]*model.ProjectMember, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// @Summary		Add project member
// @Description	Add a new member to a project
// @Tags			projects
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"Project UUID"
// @Param			request	body		object	true	"Member data {user_id: string, role: string}"
// @Success		201		{object}	model.ProjectMember
// @Failure		400		{object}	map[string]string	"Invalid request"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		404		{object}	map[string]string	"Project not found"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/api/projects/{id}/members [post]
func (h *ProjectHandler) AddProjectMember(w http.ResponseWriter, r *http.Request) {
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
		UserID uuid.UUID         `json:"user_id"`
		Role   model.ProjectRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.UserID == uuid.Nil {
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
		AddedBy:   member.UserID,
	}

	if err := h.projectRepo.AddProjectMember(ctx, newMember); err != nil {
		log.Printf("add member error: %v", err)
		http.Error(w, "failed to add member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newMember)
}

// @Summary		Update member role
// @Description	Update a member's role in a project
// @Tags			projects
// @Accept			json
// @Produce		json
// @Param			id		path		string	true	"Project UUID"
// @Param			userId	path		string	true	"User UUID"
// @Param			request	body		object	true	"Role data {role: string}"
// @Success		200		{object}	model.ProjectMember
// @Failure		400		{object}	map[string]string	"Invalid request"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		404		{object}	map[string]string	"Member not found"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/api/projects/{id}/members/{userId} [put]
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
	userID, err := uuid.Parse(userIDStr)
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

// @Summary		Remove project member
// @Description	Remove a member from a project
// @Tags			projects
// @Produce		json
// @Param			id		path	string	true	"Project UUID"
// @Param			userId	path	string	true	"User UUID"
// @Success		204		"No content"
// @Failure		400		{object}	map[string]string	"Invalid ID"
// @Failure		403		{object}	map[string]string	"Forbidden"
// @Failure		404		{object}	map[string]string	"Member not found"
// @Failure		500		{object}	map[string]string	"Internal server error"
// @Router			/api/projects/{id}/members/{userId} [delete]
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
	userID, err := uuid.Parse(userIDStr)
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
