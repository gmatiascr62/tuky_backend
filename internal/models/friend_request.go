package models

import "time"

type FriendRequestItem struct {
	ID        string    `json:"id"`
	FromID    string    `json:"from_user_id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}