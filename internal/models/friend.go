package models

type FriendItem struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}