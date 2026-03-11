package models

type AuthUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}