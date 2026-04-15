package model

type User struct {
	ID        int    `json:"id"`
	LastLogin int    `json:"last_login"`
	CreatedAt int    `json:"created_at"`
	Email     string `json:"email"`
	Password  string `json:"-"`
}
