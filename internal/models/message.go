package models

import "time"

type MessageItem struct {
	ID         string     `json:"id"`
	FromUserID string     `json:"from_user_id"`
	ToUserID   string     `json:"to_user_id"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
	ReadAt     *time.Time `json:"read_at"`
}