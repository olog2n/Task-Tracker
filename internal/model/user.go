package model

import (
	"database/sql"
	"time"
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
	ID                   int           `json:"id"`
	Email                string        `json:"email"`
	PasswordHash         string        `json:"-"`
	IsActive             bool          `json:"is_active"`
	DeletedAt            sql.NullTime  `json:"deveted_at,omitempty"`
	DeletedBy            sql.NullInt64 `json:"-"`
	TokenVersion         int           `json:"-"`
	RequirePasswordReset bool          `json:"-"`
	ReactivatedAt        sql.NullTime  `json:"-"`
	ReactivatedBy        sql.NullInt64 `json:"-"`
	CreatedAt            time.Time     `json:"created_at"`
	LastLogin            sql.NullTime  `json:"last_login,omitempty"`
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt.Valid
}

func (u *User) CanLogin() bool {
	return u.IsActive && !u.IsDeleted()
}

// TODO: Update it with correct roles
func (u *User) IsAdmin() bool {
	return u.ID == 1 //NOTE: Boilerplate, first user is admin
}
