package models

type UnreadCountItem struct {
	UserID      string `json:"user_id"`
	UnreadCount int    `json:"unread_count"`
}