package model

type User struct {
	ID        int    `json:"id"`
	LastLogin int    `json:"last_login"`
	Email     string `json:"email"`
	Password  string `json:"-"`
}
