package models

import "time"

type Profile struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Bio       *string   `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}