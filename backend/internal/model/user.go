package model

import (
	"time"

	"github.com/google/uuid"
)

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         User      `json:"user"`
}

type User struct {
	ID                   uuid.UUID `json:"id"`
	Email                string    `json:"email"`
	Name                 string    `json:"name"`
	PasswordHash         string    `json:"-"`
	IsActive             bool      `json:"is_active"`
	DeletedAt            time.Time `json:"deveted_at,omitempty"`
	DeletedBy            uuid.UUID `json:"-"`
	TokenVersion         int       `json:"-"`
	RequirePasswordReset bool      `json:"-"`
	ReactivatedAt        time.Time `json:"-"`
	ReactivatedBy        uuid.UUID `json:"-"`
	CreatedAt            time.Time `json:"created_at"`
	LastLogin            time.Time `json:"last_login,omitempty"`
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != time.Time{}
}

func (u *User) CanLogin() bool {
	// return u.IsActive && !u.IsDeleted()
	return true
}

// TODO: Update it with correct roles
func (u *User) IsAdmin() bool {
	return true
}
